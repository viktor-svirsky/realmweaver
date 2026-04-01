package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/claude"
	"realmweaver/engine"
	"realmweaver/storage"
	"realmweaver/world"
)

// OpCode defines message types between client and server.
const (
	OpCodePlayerAction int64 = 1
	OpCodeNarrative    int64 = 2
	OpCodeGameState    int64 = 3
	OpCodeMechanical   int64 = 4
	OpCodeError        int64 = 7
	OpCodeQuickActions int64 = 8
)

// MatchState holds the server-side state for a gameplay session.
type MatchState struct {
	UserID       string              `json:"user_id"`
	Character    *engine.Character   `json:"character"`
	Phase        engine.GamePhase    `json:"phase"`
	Combat       *engine.CombatState `json:"combat,omitempty"`
	Region       *world.Region       `json:"region"`
	NPCs         []world.NPC         `json:"npcs"`
	Events       []string            `json:"events"`
	Language     string              `json:"language"`
	Players      map[string]*PlayerState `json:"players,omitempty"`
	WorldDB      *world.WorldDB      `json:"-"`
	ClaudeClient *claude.Client      `json:"-"`
	Pool         *claude.WorkerPool  `json:"-"`
}

// PlayerState tracks a single player in a multiplayer match.
type PlayerState struct {
	UserID    string            `json:"user_id"`
	Character *engine.Character `json:"character"`
	Presence  runtime.Presence  `json:"-"`
}

// PlayerActionMsg is the message sent by the client.
type PlayerActionMsg struct {
	Action string `json:"action"`
	Target string `json:"target,omitempty"`
	ItemID string `json:"item_id,omitempty"`
}

// GameplayMatch implements the Nakama match handler interface.
type GameplayMatch struct{}

func (m *GameplayMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		Phase:        engine.PhaseExploring,
		Events:       []string{},
		Language:     "en",
		Players:      make(map[string]*PlayerState),
		ClaudeClient: claude.NewClient(),
		Pool:         claude.NewWorkerPool(10), // Max 10 concurrent Claude calls
	}

	worldDB, err := connectWorldDB()
	if err != nil {
		logger.Error("Failed to connect to world DB: %v", err)
	} else {
		state.WorldDB = world.NewWorldDB(worldDB)
	}

	if charID, ok := params["character_id"].(string); ok {
		if userID, ok := params["user_id"].(string); ok {
			state.UserID = userID
			character, err := storage.LoadCharacter(ctx, nk, userID, charID)
			if err != nil {
				logger.Error("Failed to load character: %v", err)
			} else if character != nil {
				state.Character = character
				if state.WorldDB != nil {
					region, _ := state.WorldDB.GetRegion(ctx, character.RegionX, character.RegionY)
					if region != nil {
						state.Region = region
						state.NPCs, _ = state.WorldDB.GetNPCsInRegion(ctx, region.ID)
					}
				}
			}
		}
	}
	if lang, ok := params["language"].(string); ok && lang != "" {
		state.Language = lang
	}
	logger.Info("Match created: language=%s, character=%v", state.Language, state.Character != nil)

	return state, 1, ""
}

func (m *GameplayMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	return state, true, ""
}

func (m *GameplayMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	ms := state.(*MatchState)

	for _, p := range presences {
		sendGameState(dispatcher, p, ms)
		sendQuickActions(dispatcher, p, ms)

		// Update player location in world DB
		if ms.Character != nil && ms.WorldDB != nil {
			ms.WorldDB.UpdatePlayerLocation(ctx, &world.PlayerLocation{
				UserID:         ms.UserID,
				CharacterID:    ms.Character.ID,
				CharacterName:  ms.Character.Name,
				CharacterClass: string(ms.Character.Class),
				CharacterLevel: ms.Character.Level,
				RegionX:        ms.Character.RegionX,
				RegionY:        ms.Character.RegionY,
				Karma:          ms.Character.Karma,
			})
		}

		if ms.Character != nil && ms.Region != nil {
			// Async narration is safe here — reads ms for context but does not mutate state.
			// MatchJoin must return quickly; Claude call can take seconds.
			go func(presence runtime.Presence) {
				narrateExploration(ctx, logger, dispatcher, presence, ms, "look around and describe the area")
			}(p)
		}
	}

	return ms
}

func (m *GameplayMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	ms := state.(*MatchState)
	if ms.Character != nil && ms.UserID != "" {
		storage.SaveCharacter(ctx, nk, ms.UserID, ms.Character)
		// Remove player from world map
		if ms.WorldDB != nil {
			ms.WorldDB.RemovePlayerLocation(ctx, ms.UserID)
		}
	}
	return ms
}

func (m *GameplayMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	ms := state.(*MatchState)

	for _, msg := range messages {
		if msg.GetOpCode() != OpCodePlayerAction {
			continue
		}

		var action PlayerActionMsg
		if err := json.Unmarshal(msg.GetData(), &action); err != nil {
			sendError(dispatcher, msg, "Invalid action format")
			continue
		}

		presence := msg
		handleAction(ctx, logger, nk, dispatcher, presence, ms, &action)
	}

	return ms
}

func (m *GameplayMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	ms := state.(*MatchState)
	if ms.Character != nil && ms.UserID != "" {
		storage.SaveCharacter(ctx, nk, ms.UserID, ms.Character)
	}
	return ms
}

func (m *GameplayMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

func handleAction(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, sender runtime.MatchData, ms *MatchState, action *PlayerActionMsg) {
	if ms.Character == nil {
		sendError(dispatcher, sender, "No character loaded")
		return
	}

	switch ms.Phase {
	case engine.PhaseInCombat:
		handleCombatAction(ctx, logger, nk, dispatcher, sender, ms, action)
	case engine.PhaseExploring:
		handleExploreAction(ctx, logger, nk, dispatcher, sender, ms, action)
	case engine.PhaseInDialogue:
		handleDialogueAction(ctx, logger, dispatcher, sender, ms, action)
	case engine.PhaseInShop:
		handleShopAction(ctx, logger, dispatcher, sender, ms, action)
	}

	if ms.UserID != "" {
		storage.SaveCharacter(ctx, nk, ms.UserID, ms.Character)
	}

	sendGameState(dispatcher, sender, ms)
	sendQuickActions(dispatcher, sender, ms)
}

func handleCombatAction(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, sender runtime.MatchData, ms *MatchState, action *PlayerActionMsg) {
	if ms.Combat == nil {
		ms.Phase = engine.PhaseExploring
		return
	}

	var result engine.ActionResult

	// Handle "use_skill:<skill_id>" quick action format
	if strings.HasPrefix(action.Action, "use_skill:") {
		action.Target = strings.TrimPrefix(action.Action, "use_skill:")
		action.Action = "use_skill"
	}

	switch action.Action {
	case "attack":
		enemies := ms.Combat.AliveEnemies()
		if len(enemies) == 0 {
			return
		}
		target := &ms.Combat.Enemies[0]
		for i := range ms.Combat.Enemies {
			if ms.Combat.Enemies[i].HP > 0 {
				if action.Target == "" || ms.Combat.Enemies[i].ID == action.Target {
					target = &ms.Combat.Enemies[i]
					break
				}
			}
		}
		result = engine.ResolveMeleeAttack(ms.Character, target)

	case "defend":
		result = engine.ResolveDefend(ms.Character, ms.Combat)

	case "flee":
		result = engine.ResolveFlee(ms.Character)
		if result.Success {
			// Reset buffs/debuffs from combat
			ms.Character.RecalcDerived()
			ms.Phase = engine.PhaseExploring
			ms.Combat = nil
		}

	case "use_item":
		desc, err := engine.UseItem(ms.Character, action.ItemID)
		if err != nil {
			sendError(dispatcher, sender, err.Error())
			return
		}
		result = engine.ActionResult{
			Action:  "use_item",
			Actor:   ms.Character.Name,
			Success: true,
			Details: desc,
		}

	case "use_skill":
		skill := engine.FindSkillByID(action.Target)
		if skill == nil {
			sendError(dispatcher, sender, "Unknown skill")
			return
		}
		if ms.Character.Mana < skill.ManaCost {
			sendError(dispatcher, sender, "Not enough mana")
			return
		}
		available := engine.GetAvailableSkills(ms.Character.Class, ms.Character.Level)
		found := false
		for _, s := range available {
			if s.ID == skill.ID {
				found = true
				break
			}
		}
		if !found {
			sendError(dispatcher, sender, "Skill not available")
			return
		}

		var skillTarget interface{}
		switch skill.TargetType {
		case "enemy":
			for i := range ms.Combat.Enemies {
				if ms.Combat.Enemies[i].HP > 0 {
					if action.ItemID == "" || ms.Combat.Enemies[i].ID == action.ItemID {
						skillTarget = &ms.Combat.Enemies[i]
						break
					}
				}
			}
		case "all_enemies":
			var targets []*engine.Enemy
			for i := range ms.Combat.Enemies {
				if ms.Combat.Enemies[i].HP > 0 {
					targets = append(targets, &ms.Combat.Enemies[i])
				}
			}
			skillTarget = targets
		case "self":
			skillTarget = nil
		}

		skillResult := engine.ResolveSkill(ms.Character, skill, skillTarget)
		result = engine.ActionResult{
			Action:     "skill_attack",
			Actor:      ms.Character.Name,
			Target:     skill.Name,
			Damage:     skillResult.Damage,
			DamageType: skillResult.DamageType,
			Success:    true,
			Details:    skillResult.Description,
		}

	default:
		enemies := ms.Combat.AliveEnemies()
		if len(enemies) > 0 {
			result = engine.ResolveMeleeAttack(ms.Character, &ms.Combat.Enemies[0])
		}
	}

	sendMechanical(dispatcher, sender, &result)
	narrateAction(ctx, logger, dispatcher, sender, ms, &result, action.Action)

	if ms.Combat != nil && ms.Combat.IsOver() {
		xp := engine.CalculateXP(ms.Combat.Enemies)
		gold, lootItems := engine.GenerateLootFromEnemies(ms.Combat.Enemies, ms.Character.Level)

		leveledUp := ms.Character.GainXP(xp)
		ms.Character.Gold += gold
		for _, item := range lootItems {
			engine.AddItem(ms.Character, item)
		}

		// Karma: -50 per enemy killed (working off PK sins)
		for _, enemy := range ms.Combat.Enemies {
			if enemy.HP <= 0 {
				ms.Character.ReduceKarma(50)
			}
		}

		// Build loot notification
		lootText := engine.FormatLootNotification(gold, lootItems)

		// Reset buffs/debuffs from combat (AC, stats back to base + equipment)
		ms.Character.RecalcDerived()

		victoryResult := engine.ActionResult{
			Action:      "combat_victory",
			Actor:       ms.Character.Name,
			Success:     true,
			Victory:     true,
			CombatEnded: true,
			XPGained:    xp,
			LeveledUp:   leveledUp,
			LootDropped: lootItems,
			Details:     fmt.Sprintf("Victory! +%d XP%s\n%s", xp, func() string { if leveledUp { return " LEVEL UP!" }; return "" }(), lootText),
		}
		sendMechanical(dispatcher, sender, &victoryResult)
		narrateAction(ctx, logger, dispatcher, sender, ms, &victoryResult, "victory after defeating all enemies")
		ms.Phase = engine.PhaseExploring
		ms.Combat = nil
		return
	}

	if ms.Combat != nil && ms.Phase == engine.PhaseInCombat {
		for i := range ms.Combat.Enemies {
			enemy := &ms.Combat.Enemies[i]
			if enemy.HP <= 0 {
				continue
			}
			enemyResult := engine.ResolveEnemyAttack(enemy, ms.Character, ms.Combat.PlayerDefending)
			sendMechanical(dispatcher, sender, &enemyResult)
			narrateAction(ctx, logger, dispatcher, sender, ms, &enemyResult, "enemy attacks")

			if !ms.Character.IsAlive() {
				deathResult := engine.ActionResult{
					Action:      "player_death",
					Actor:       ms.Character.Name,
					Success:     false,
					CombatEnded: true,
					Details:     "You have fallen in battle.",
				}
				sendMechanical(dispatcher, sender, &deathResult)
				narrateAction(ctx, logger, dispatcher, sender, ms, &deathResult, "player death")
				ms.Phase = engine.PhaseExploring
				ms.Combat = nil
				ms.Character.RecalcDerived() // Reset buffs/debuffs
				ms.Character.HP = ms.Character.MaxHP / 2
				ms.Character.Mana = ms.Character.MaxMana / 2
				return
			}
		}
	}
}

func handleExploreAction(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, sender runtime.MatchData, ms *MatchState, action *PlayerActionMsg) {
	switch action.Action {
	case "search":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "search the area carefully")
	case "look":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "look around and describe the area")
	case "rest":
		ms.Character.Rest()
		result := engine.ActionResult{
			Action:  "rest",
			Actor:   ms.Character.Name,
			Success: true,
			Details: "Fully rested. HP and Mana restored.",
		}
		sendMechanical(dispatcher, sender, &result)
		narrateExploration(ctx, logger, dispatcher, sender, ms, "rest at the current location")
	case "enter_dungeon":
		room := 1
		enemies := engine.GoblinCaveEnemies(room)
		ms.Combat = engine.InitCombat(ms.Character, enemies)
		ms.Phase = engine.PhaseInCombat
		narrateAction(ctx, logger, dispatcher, sender, ms, &engine.ActionResult{
			Action:  "combat_start",
			Actor:   ms.Character.Name,
			Success: true,
			Details: "Entering the Goblin Cave. Combat begins!",
		}, "entering the Goblin Cave and encountering enemies")
	case "travel_complete":
		// Reload character from storage to get updated position
		if ms.UserID != "" {
			updated, err := storage.LoadCharacter(ctx, nk, ms.UserID, ms.Character.ID)
			if err == nil && updated != nil {
				ms.Character = updated
				// Reload region for new position
				if ms.WorldDB != nil {
					region, _ := ms.WorldDB.GetRegion(ctx, ms.Character.RegionX, ms.Character.RegionY)
					if region != nil {
						ms.Region = region
						ms.NPCs, _ = ms.WorldDB.GetNPCsInRegion(ctx, region.ID)
					}
					// Update world location
					ms.WorldDB.UpdatePlayerLocation(ctx, &world.PlayerLocation{
						UserID:         ms.UserID,
						CharacterID:    ms.Character.ID,
						CharacterName:  ms.Character.Name,
						CharacterClass: string(ms.Character.Class),
						CharacterLevel: ms.Character.Level,
						RegionX:        ms.Character.RegionX,
						RegionY:        ms.Character.RegionY,
						Karma:          ms.Character.Karma,
					})
				}
				logger.Info("Player traveled to (%d,%d) - %s (%s)", ms.Character.RegionX, ms.Character.RegionY, ms.Region.Name, ms.Region.Biome)
				// Narrate arriving at the new location — Claude will use the player's language
				narrateExploration(ctx, logger, dispatcher, sender, ms, fmt.Sprintf("you just traveled to a new region called %s (a %s area). Describe your arrival and what you see in this new place", ms.Region.Name, ms.Region.Biome))
			}
		}
	case "tavern":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "enter The Wanderer's Rest tavern")
	case "forge":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "visit the Ironheart Forge blacksmith shop")
	case "chapel":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "visit the Chapel of Dawn")
	case "square":
		narrateExploration(ctx, logger, dispatcher, sender, ms, "walk to the Town Square and look around the market stalls")
	default:
		// Check for dynamic NPC talk actions (talk_<name>)
		if strings.HasPrefix(action.Action, "talk_") {
			ms.Phase = engine.PhaseInDialogue
			npcName := strings.TrimPrefix(action.Action, "talk_")
			npcName = strings.ReplaceAll(npcName, "_", " ")
			narrateDialogue(ctx, logger, dispatcher, sender, ms, npcName, "greet and introduce yourself")
			return
		}
		// Sanitize free-form player input for Claude
		text := action.Action
		if len(text) > 200 {
			text = text[:200]
		}
		narrateExploration(ctx, logger, dispatcher, sender, ms, text)
	}
}

func handleDialogueAction(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.MatchData, ms *MatchState, action *PlayerActionMsg) {
	switch action.Action {
	case "leave":
		ms.Phase = engine.PhaseExploring
	case "ask_quest":
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, "ask about available quests or work")
	case "ask_rumors":
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, "ask about any rumors or interesting news")
	case "trade":
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, "ask about buying and selling items")
	case "ask_cave":
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, "ask about the Goblin Cave and what dangers lurk there")
	case "greet":
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, "greet warmly and introduce yourself")
	default:
		// Sanitize free-form player input for Claude
		text := action.Action
		if len(text) > 200 {
			text = text[:200]
		}
		narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, text)
	}
}

func handleShopAction(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.MatchData, ms *MatchState, action *PlayerActionMsg) {
	if action.Action == "leave" {
		ms.Phase = engine.PhaseExploring
		return
	}
	narrateDialogue(ctx, logger, dispatcher, sender, ms, action.Target, action.Action)
}

// narrateExploration uses tiered AI: cache for revisits, haiku for exploration, sonnet for first visits.
func narrateExploration(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.Presence, ms *MatchState, playerText string) {
	var regionSummary *world.RegionSummary
	if ms.Region != nil {
		s := world.BuildRegionSummary(ms.Region, ms.NPCs)
		regionSummary = &s
	}

	gameCtx := &claude.GameContext{
		Character:    ms.Character,
		Location:     regionSummary,
		Phase:        ms.Phase,
		RecentEvents: lastN(ms.Events, 5),
		PlayerAction: playerText,
		Language:     ms.Language,
	}

	systemPrompt := claude.SelectSystemPrompt(ms.Phase, ms.Language)
	userMessage := claude.BuildUserMessage(gameCtx)

	tier := claude.ClassifyAction(playerText, ms.Phase)
	cacheKey := ""
	if ms.Region != nil {
		cacheKey = fmt.Sprintf("explore:%d:%d:%s:%s", ms.Region.X, ms.Region.Y, playerText, ms.Language)
	}

	logger.Info("Narrating: tier=%d, language=%s, action=%s [%s]", tier, ms.Language, playerText, ms.Pool.StatsString())

	resp, err := ms.Pool.Generate(ms.UserID, tier, systemPrompt, userMessage, cacheKey, nil, ms.Character, ms.Language)
	if err != nil {
		logger.Error("Claude error: %v", err)
		sendNarrative(dispatcher, sender, "The world shimmers for a moment, then settles. You take in your surroundings.")
		return
	}

	sendNarrative(dispatcher, sender, resp.Narrative)
}

// narrateAction uses tiered AI: templates for routine combat, haiku for dramatic moments.
func narrateAction(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.Presence, ms *MatchState, result *engine.ActionResult, playerText string) {
	tier := claude.ClassifyAction(result.Action, ms.Phase)

	// Template tier — no API call needed
	if tier == claude.TierTemplate {
		resp, _ := ms.Pool.Generate(ms.UserID, tier, "", "", "", result, ms.Character, ms.Language)
		sendNarrative(dispatcher, sender, resp.Narrative)
		return
	}

	var regionSummary *world.RegionSummary
	if ms.Region != nil {
		s := world.BuildRegionSummary(ms.Region, ms.NPCs)
		regionSummary = &s
	}

	gameCtx := &claude.GameContext{
		Character:    ms.Character,
		Location:     regionSummary,
		Phase:        ms.Phase,
		MechResult:   result,
		RecentEvents: lastN(ms.Events, 5),
		PlayerAction: playerText,
		Language:     ms.Language,
	}

	systemPrompt := claude.SelectSystemPrompt(ms.Phase, ms.Language)
	userMessage := claude.BuildUserMessage(gameCtx)

	logger.Info("Narrating action: tier=%d, action=%s [%s]", tier, result.Action, ms.Pool.StatsString())

	resp, err := ms.Pool.Generate(ms.UserID, tier, systemPrompt, userMessage, "", result, ms.Character, ms.Language)
	if err != nil {
		logger.Error("Claude error: %v", err)
		sendNarrative(dispatcher, sender, result.Details)
		return
	}

	sendNarrative(dispatcher, sender, resp.Narrative)
}

// narrateDialogue always uses Sonnet — NPC dialogue is the premium experience.
func narrateDialogue(ctx context.Context, logger runtime.Logger, dispatcher runtime.MatchDispatcher, sender runtime.Presence, ms *MatchState, npcName, playerText string) {
	var npcCtx *claude.NPCContext
	for _, npc := range ms.NPCs {
		if strings.EqualFold(npc.Name, npcName) || npcName == "" {
			npcCtx = &claude.NPCContext{
				Name:           npc.Name,
				Occupation:     npc.Occupation,
				Personality:    npc.PersonalityPrompt,
				Disposition:    0,
				LocationDetail: npc.LocationDetail,
			}
			break
		}
	}

	if npcCtx == nil && len(ms.NPCs) > 0 {
		npc := ms.NPCs[0]
		npcCtx = &claude.NPCContext{
			Name:           npc.Name,
			Occupation:     npc.Occupation,
			Personality:    npc.PersonalityPrompt,
			Disposition:    0,
			LocationDetail: npc.LocationDetail,
		}
	}

	var regionSummary *world.RegionSummary
	if ms.Region != nil {
		s := world.BuildRegionSummary(ms.Region, ms.NPCs)
		regionSummary = &s
	}

	gameCtx := &claude.GameContext{
		Character:    ms.Character,
		Location:     regionSummary,
		Phase:        engine.PhaseInDialogue,
		NPCContext:   npcCtx,
		RecentEvents: lastN(ms.Events, 5),
		PlayerAction: playerText,
		Language:     ms.Language,
	}

	systemPrompt := claude.SelectSystemPrompt(engine.PhaseInDialogue, ms.Language)
	userMessage := claude.BuildUserMessage(gameCtx)

	npcCacheKey := ""
	if npcCtx != nil {
		npcCacheKey = fmt.Sprintf("dialogue:%s:%s:%s", npcCtx.Name, playerText, ms.Language)
	}

	logger.Info("Narrating dialogue: NPC=%s, action=%s [%s]", npcName, playerText, ms.Pool.StatsString())

	resp, err := ms.Pool.Generate(ms.UserID, claude.TierSonnet, systemPrompt, userMessage, npcCacheKey, nil, ms.Character, ms.Language)
	if err != nil {
		logger.Error("Claude dialogue error: %v", err)
		sendNarrative(dispatcher, sender, "The NPC regards you silently.")
		return
	}

	sendNarrative(dispatcher, sender, resp.Narrative)
}

// Helper functions

func sendGameState(dispatcher runtime.MatchDispatcher, presence runtime.Presence, ms *MatchState) {
	data, _ := json.Marshal(map[string]interface{}{
		"character": ms.Character,
		"phase":     ms.Phase,
		"combat":    ms.Combat,
	})
	dispatcher.BroadcastMessage(OpCodeGameState, data, []runtime.Presence{presence}, nil, true)
}

func sendQuickActions(dispatcher runtime.MatchDispatcher, presence runtime.Presence, ms *MatchState) {
	if ms.Phase == engine.PhaseExploring {
		// Build dynamic actions based on current region
		actions := []engine.QuickAction{
			{ID: "look", Label: "Look Around", Icon: "magnifier"},
			{ID: "travel", Label: "Travel", Icon: "map"},
		}
		// Add region structures as locations
		if ms.Region != nil {
			structures := world.ParseStructures(ms.Region.Structures)
			for _, s := range structures {
				switch s.Type {
				case "tavern":
					actions = append(actions, engine.QuickAction{ID: "tavern", Label: s.Name, Icon: "tavern"})
				case "shop":
					actions = append(actions, engine.QuickAction{ID: "forge", Label: s.Name, Icon: "anvil"})
				case "square":
					actions = append(actions, engine.QuickAction{ID: "square", Label: s.Name, Icon: "market"})
				case "house":
					actions = append(actions, engine.QuickAction{ID: "chapel", Label: s.Name, Icon: "chapel"})
				case "dungeon":
					actions = append(actions, engine.QuickAction{ID: "enter_dungeon", Label: s.Name, Icon: "dungeon"})
				}
			}
		}
		// Add NPCs from current region
		for _, npc := range ms.NPCs {
			npcID := "talk_" + strings.ToLower(strings.ReplaceAll(npc.Name, " ", "_"))
			actions = append(actions, engine.QuickAction{ID: npcID, Label: "Talk to " + npc.Name, Icon: "npc"})
		}
		actions = append(actions, engine.QuickAction{ID: "rest", Label: "Rest", Icon: "campfire"})
		data, _ := json.Marshal(actions)
		dispatcher.BroadcastMessage(OpCodeQuickActions, data, []runtime.Presence{presence}, nil, true)
	} else {
		actions := engine.QuickActionsForPhaseWithChar(ms.Phase, ms.Combat, ms.Character)
		data, _ := json.Marshal(actions)
		dispatcher.BroadcastMessage(OpCodeQuickActions, data, []runtime.Presence{presence}, nil, true)
	}
}

func sendNarrative(dispatcher runtime.MatchDispatcher, presence runtime.Presence, text string) {
	data, _ := json.Marshal(map[string]string{"text": text})
	dispatcher.BroadcastMessage(OpCodeNarrative, data, []runtime.Presence{presence}, nil, true)
}

func sendMechanical(dispatcher runtime.MatchDispatcher, sender runtime.Presence, result *engine.ActionResult) {
	data, _ := json.Marshal(result)
	dispatcher.BroadcastMessage(OpCodeMechanical, data, []runtime.Presence{sender}, nil, true)
}

func sendError(dispatcher runtime.MatchDispatcher, sender runtime.Presence, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	dispatcher.BroadcastMessage(OpCodeError, data, []runtime.Presence{sender}, nil, true)
}

func lastN(events []string, n int) []string {
	if len(events) <= n {
		return events
	}
	return events[len(events)-n:]
}

func connectWorldDB() (*sql.DB, error) {
	dsn := os.Getenv("WORLD_DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("WORLD_DB_DSN environment variable is required")
	}
	return sql.Open("postgres", dsn)
}
