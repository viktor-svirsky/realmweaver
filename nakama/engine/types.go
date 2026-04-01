package engine

import "fmt"

// Character classes
type Class string

const (
	ClassWarrior     Class = "warrior"
	ClassMage        Class = "mage"
	ClassRogue       Class = "rogue"
	ClassCleric      Class = "cleric"
	ClassRanger      Class = "ranger"
	ClassPaladin     Class = "paladin"
	ClassNecromancer Class = "necromancer"
	ClassBerserker   Class = "berserker"
)

// Stats represents the six core ability scores.
type Stats struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

// Modifier returns the D&D-style modifier for a stat value: (stat - 10) / 2.
func Modifier(stat int) int {
	return (stat - 10) / 2
}

// StartingStats returns base stats for each class.
func StartingStats(class Class) Stats {
	switch class {
	case ClassWarrior:
		return Stats{STR: 16, DEX: 12, CON: 14, INT: 8, WIS: 10, CHA: 10}
	case ClassMage:
		return Stats{STR: 8, DEX: 12, CON: 10, INT: 16, WIS: 14, CHA: 10}
	case ClassRogue:
		return Stats{STR: 10, DEX: 16, CON: 12, INT: 14, WIS: 10, CHA: 8}
	case ClassCleric:
		return Stats{STR: 12, DEX: 10, CON: 14, INT: 10, WIS: 16, CHA: 8}
	case ClassRanger:
		return Stats{STR: 10, DEX: 16, CON: 12, INT: 10, WIS: 14, CHA: 8}
	case ClassPaladin:
		return Stats{STR: 14, DEX: 10, CON: 12, INT: 8, WIS: 10, CHA: 16}
	case ClassNecromancer:
		return Stats{STR: 8, DEX: 10, CON: 14, INT: 16, WIS: 10, CHA: 12}
	case ClassBerserker:
		return Stats{STR: 18, DEX: 12, CON: 14, INT: 8, WIS: 8, CHA: 10}
	default:
		return Stats{STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10}
	}
}

// EquipSlot represents where an item can be equipped.
type EquipSlot string

const (
	SlotWeapon  EquipSlot = "weapon"
	SlotOffhand EquipSlot = "offhand"
	SlotArmor   EquipSlot = "armor"
	SlotHelmet  EquipSlot = "helmet"
	SlotBoots   EquipSlot = "boots"
	SlotRing1   EquipSlot = "ring1"
	SlotRing2   EquipSlot = "ring2"
	SlotAmulet  EquipSlot = "amulet"
)

// ItemType categorizes items.
type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeMisc       ItemType = "misc"
)

// DamageType for weapons and spells.
type DamageType string

const (
	DamagePhysical DamageType = "physical"
	DamageFire     DamageType = "fire"
	DamageIce      DamageType = "ice"
	DamageLightning DamageType = "lightning"
	DamageHoly     DamageType = "holy"
	DamageNecrotic DamageType = "necrotic"
)

// Item represents any item in the game.
type Item struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ItemType    `json:"type"`
	Slot        EquipSlot   `json:"slot,omitempty"`
	DamageDice  int         `json:"damage_dice,omitempty"`
	DamageCount int         `json:"damage_count,omitempty"`
	DamageType  DamageType  `json:"damage_type,omitempty"`
	ACBonus     int         `json:"ac_bonus,omitempty"`
	HealAmount  int         `json:"heal_amount,omitempty"`
	ManaRestore int         `json:"mana_restore,omitempty"`
	Weight      float64     `json:"weight"`
	Value       int         `json:"value"`
	Quality     ItemQuality `json:"quality,omitempty"`
	Mods        []ItemMod   `json:"mods,omitempty"`
}

// Equipment holds the currently equipped items.
type Equipment struct {
	Weapon  *Item `json:"weapon,omitempty"`
	Offhand *Item `json:"offhand,omitempty"`
	Armor   *Item `json:"armor,omitempty"`
	Helmet  *Item `json:"helmet,omitempty"`
	Boots   *Item `json:"boots,omitempty"`
	Ring1   *Item `json:"ring1,omitempty"`
	Ring2   *Item `json:"ring2,omitempty"`
	Amulet  *Item `json:"amulet,omitempty"`
}

// GetSlot returns the equipped item for a given slot.
func (e *Equipment) GetSlot(slot EquipSlot) *Item {
	switch slot {
	case SlotWeapon:
		return e.Weapon
	case SlotOffhand:
		return e.Offhand
	case SlotArmor:
		return e.Armor
	case SlotHelmet:
		return e.Helmet
	case SlotBoots:
		return e.Boots
	case SlotRing1:
		return e.Ring1
	case SlotRing2:
		return e.Ring2
	case SlotAmulet:
		return e.Amulet
	}
	return nil
}

// SetSlot equips an item in a slot, returning any previously equipped item.
func (e *Equipment) SetSlot(slot EquipSlot, item *Item) *Item {
	var prev *Item
	switch slot {
	case SlotWeapon:
		prev, e.Weapon = e.Weapon, item
	case SlotOffhand:
		prev, e.Offhand = e.Offhand, item
	case SlotArmor:
		prev, e.Armor = e.Armor, item
	case SlotHelmet:
		prev, e.Helmet = e.Helmet, item
	case SlotBoots:
		prev, e.Boots = e.Boots, item
	case SlotRing1:
		prev, e.Ring1 = e.Ring1, item
	case SlotRing2:
		prev, e.Ring2 = e.Ring2, item
	case SlotAmulet:
		prev, e.Amulet = e.Amulet, item
	}
	return prev
}

// TotalACBonus sums AC from all equipped armor pieces.
func (e *Equipment) TotalACBonus() int {
	total := 0
	for _, item := range []*Item{e.Armor, e.Helmet, e.Boots, e.Offhand} {
		if item != nil {
			total += item.ACBonus
		}
	}
	return total
}

// Character represents a player character.
type Character struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Class     Class     `json:"class"`
	Level     int       `json:"level"`
	XP        int       `json:"xp"`
	Stats     Stats     `json:"stats"`
	HP        int       `json:"hp"`
	MaxHP     int       `json:"max_hp"`
	Mana      int       `json:"mana"`
	MaxMana   int       `json:"max_mana"`
	AC        int       `json:"ac"`
	Gold         int       `json:"gold"`
	Karma        int       `json:"karma"`          // PK karma (0=clean, higher=worse)
	PKCount      int       `json:"pk_count"`       // innocent player kills
	PvPCount     int       `json:"pvp_count"`      // consensual PvP kills
	Flagged      bool      `json:"flagged"`        // purple name (attacked someone)
	FlaggedUntil int64     `json:"flagged_until"`  // unix timestamp when flag expires
	Equipment    Equipment `json:"equipment"`
	Inventory []Item    `json:"inventory"`
	RegionX   int       `json:"region_x"`
	RegionY   int       `json:"region_y"`
}

// XPForLevel returns the XP needed to reach a given level.
func XPForLevel(level int) int {
	return level * level * 100
}

// GamePhase tracks what the player is currently doing.
type GamePhase string

const (
	PhaseExploring  GamePhase = "exploring"
	PhaseInCombat   GamePhase = "in_combat"
	PhaseInDialogue GamePhase = "in_dialogue"
	PhaseInShop     GamePhase = "in_shop"
	PhaseTraveling  GamePhase = "traveling"
)

// Enemy represents a hostile creature.
type Enemy struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	AC     int    `json:"ac"`
	STR    int    `json:"str"`
	DEX    int    `json:"dex"`
	Damage string `json:"damage"` // e.g. "1d6+2"
	XP     int    `json:"xp"`     // XP awarded on kill
}

// CombatState tracks an ongoing combat encounter.
type CombatState struct {
	Enemies        []Enemy  `json:"enemies"`
	TurnOrder      []string `json:"turn_order"` // IDs in initiative order
	CurrentTurn    int      `json:"current_turn"`
	Round          int      `json:"round"`
	PlayerDefending bool    `json:"player_defending"`
}

// ActionResult is sent to Claude for narration.
type ActionResult struct {
	Action       string            `json:"action"`        // e.g. "melee_attack", "skill_check", "explore"
	Actor        string            `json:"actor"`
	Target       string            `json:"target,omitempty"`
	Roll         int               `json:"roll,omitempty"`
	Modifier     int               `json:"modifier,omitempty"`
	DC           int               `json:"dc,omitempty"`
	Hit          bool              `json:"hit,omitempty"`
	Damage       int               `json:"damage,omitempty"`
	DamageType   DamageType        `json:"damage_type,omitempty"`
	Success      bool              `json:"success"`
	Details      string            `json:"details"`
	HPRemaining  int               `json:"hp_remaining,omitempty"`
	EnemyDefeated bool             `json:"enemy_defeated,omitempty"`
	CombatEnded  bool              `json:"combat_ended,omitempty"`
	Victory      bool              `json:"victory,omitempty"`
	XPGained     int               `json:"xp_gained,omitempty"`
	LeveledUp    bool              `json:"leveled_up,omitempty"`
	LootDropped  []Item            `json:"loot_dropped,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// String returns a mechanical summary for display alongside narrative.
func (r *ActionResult) MechanicalSummary() string {
	switch r.Action {
	case "melee_attack", "spell_attack":
		if r.Hit {
			return fmt.Sprintf("Roll: %d+%d vs AC %d — HIT for %d %s damage (%s: %d HP remaining)",
				r.Roll, r.Modifier, r.DC, r.Damage, r.DamageType, r.Target, r.HPRemaining)
		}
		return fmt.Sprintf("Roll: %d+%d vs AC %d — MISS", r.Roll, r.Modifier, r.DC)
	case "skill_check":
		if r.Success {
			return fmt.Sprintf("Skill check: %d+%d vs DC %d — SUCCESS", r.Roll, r.Modifier, r.DC)
		}
		return fmt.Sprintf("Skill check: %d+%d vs DC %d — FAILURE", r.Roll, r.Modifier, r.DC)
	default:
		return r.Details
	}
}

// ClaudeResponse is the structured response from Claude.
type ClaudeResponse struct {
	Narrative string      `json:"narrative"`
	Hints     ClaudeHints `json:"hints"`
}

// ClaudeHints are advisory suggestions from Claude, validated by the engine.
type ClaudeHints struct {
	XPSuggestion       int                  `json:"xp_suggestion,omitempty"`
	DispositionChanges []DispositionChange   `json:"disposition_changes,omitempty"`
	QuestHooks         []string             `json:"quest_hooks,omitempty"`
	Mood               string               `json:"mood,omitempty"`
}

type DispositionChange struct {
	NPCID string `json:"npc_id"`
	Delta int    `json:"delta"`
}

// QuickAction represents a context-sensitive button for the UI.
type QuickAction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// QuickActionsForPhase returns available actions for the current game phase.
func QuickActionsForPhase(phase GamePhase, combat *CombatState) []QuickAction {
	return QuickActionsForPhaseWithChar(phase, combat, nil)
}

// QuickActionsForPhaseWithChar returns available actions including class skills.
func QuickActionsForPhaseWithChar(phase GamePhase, combat *CombatState, character *Character) []QuickAction {
	switch phase {
	case PhaseInCombat:
		actions := []QuickAction{
			{ID: "attack", Label: "Attack", Icon: "sword"},
			{ID: "defend", Label: "Defend", Icon: "shield"},
			{ID: "use_item", Label: "Use Item", Icon: "potion"},
			{ID: "flee", Label: "Flee", Icon: "run"},
		}
		if character != nil {
			skills := GetAvailableSkills(character.Class, character.Level)
			for _, skill := range skills {
				if character.Mana >= skill.ManaCost {
					actions = append(actions, QuickAction{
						ID:    "use_skill:" + skill.ID,
						Label: fmt.Sprintf("%s (%d MP)", skill.Name, skill.ManaCost),
						Icon:  "magic",
					})
				}
			}
		}
		return actions
	case PhaseExploring:
		return []QuickAction{
			{ID: "look", Label: "Look Around", Icon: "magnifier"},
			{ID: "open_map", Label: "Travel", Icon: "map"},
			{ID: "tavern", Label: "Tavern", Icon: "tavern"},
			{ID: "forge", Label: "Forge", Icon: "anvil"},
			{ID: "square", Label: "Town Square", Icon: "market"},
			{ID: "chapel", Label: "Chapel", Icon: "chapel"},
			{ID: "enter_dungeon", Label: "Goblin Cave", Icon: "dungeon"},
			{ID: "talk_marta", Label: "Talk to Marta", Icon: "npc"},
			{ID: "talk_theron", Label: "Talk to Theron", Icon: "npc"},
			{ID: "talk_corin", Label: "Talk to Elder Corin", Icon: "npc"},
			{ID: "talk_lina", Label: "Talk to Sister Lina", Icon: "npc"},
			{ID: "talk_pip", Label: "Talk to Pip", Icon: "npc"},
			{ID: "rest", Label: "Rest", Icon: "campfire"},
		}
	case PhaseInDialogue:
		return []QuickAction{
			{ID: "ask_quest", Label: "Ask about quests", Icon: "scroll"},
			{ID: "ask_rumors", Label: "Ask for rumors", Icon: "ear"},
			{ID: "ask_cave", Label: "Ask about the cave", Icon: "dungeon"},
			{ID: "trade", Label: "Trade", Icon: "coins"},
			{ID: "leave", Label: "Leave conversation", Icon: "door"},
		}
	case PhaseInShop:
		return []QuickAction{
			{ID: "buy", Label: "Buy", Icon: "coins"},
			{ID: "sell", Label: "Sell", Icon: "coins"},
			{ID: "leave", Label: "Leave", Icon: "door"},
		}
	}
	return nil
}
