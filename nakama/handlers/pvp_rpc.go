package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/engine"
	"realmweaver/storage"
)

// PvPResult represents the outcome of a PvP round.
type PvPResult struct {
	AttackerName   string `json:"attacker_name"`
	DefenderName   string `json:"defender_name"`
	AttackerRoll   int    `json:"attacker_roll"`
	DefenderRoll   int    `json:"defender_roll"`
	Damage         int    `json:"damage"`
	AttackerHP     int    `json:"attacker_hp"`
	DefenderHP     int    `json:"defender_hp"`
	AttackerMaxHP  int    `json:"attacker_max_hp"`
	DefenderMaxHP  int    `json:"defender_max_hp"`
	Winner         string `json:"winner,omitempty"`
	Narrative      string `json:"narrative"`
}

// RPCPvPChallenge initiates a PvP duel with another player.
// Resolves instantly (auto-accept for MVP). Both players must be in the same region.
func RPCPvPChallenge(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	attackerID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || attackerID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		AttackerCharID string `json:"attacker_char_id"`
		DefenderUserID string `json:"defender_user_id"`
		DefenderCharID string `json:"defender_char_id"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	// Load both characters
	attacker, err := storage.LoadCharacter(ctx, nk, attackerID, req.AttackerCharID)
	if err != nil || attacker == nil {
		return "", runtime.NewError("attacker character not found", 5)
	}

	defender, err := storage.LoadCharacter(ctx, nk, req.DefenderUserID, req.DefenderCharID)
	if err != nil || defender == nil {
		return "", runtime.NewError("defender character not found", 5)
	}

	// Must be in same region
	if attacker.RegionX != defender.RegionX || attacker.RegionY != defender.RegionY {
		return "", runtime.NewError("players must be in the same region", 3)
	}

	// Resolve PvP: 3 rounds, each player attacks, highest total damage wins
	var rounds []PvPResult
	attackerHP := attacker.HP
	defenderHP := defender.HP

	for round := 1; round <= 5; round++ {
		// Attacker strikes
		atkRoll := engine.RollD20() + engine.Modifier(attacker.Stats.STR)
		if atkRoll >= defender.AC {
			dmg := engine.RollDice(1, 8) + engine.Modifier(attacker.Stats.STR)
			if dmg < 1 { dmg = 1 }
			defenderHP -= dmg
			rounds = append(rounds, PvPResult{
				AttackerName:  attacker.Name,
				DefenderName:  defender.Name,
				AttackerRoll:  atkRoll,
				Damage:        dmg,
				AttackerHP:    attackerHP,
				DefenderHP:    defenderHP,
				AttackerMaxHP: attacker.MaxHP,
				DefenderMaxHP: defender.MaxHP,
				Narrative:     fmt.Sprintf("%s strikes %s for %d damage!", attacker.Name, defender.Name, dmg),
			})
		} else {
			rounds = append(rounds, PvPResult{
				AttackerName:  attacker.Name,
				DefenderName:  defender.Name,
				AttackerRoll:  atkRoll,
				AttackerHP:    attackerHP,
				DefenderHP:    defenderHP,
				AttackerMaxHP: attacker.MaxHP,
				DefenderMaxHP: defender.MaxHP,
				Narrative:     fmt.Sprintf("%s swings at %s but misses!", attacker.Name, defender.Name),
			})
		}

		if defenderHP <= 0 {
			rounds[len(rounds)-1].Winner = attacker.Name
			break
		}

		// Defender strikes back
		defRoll := engine.RollD20() + engine.Modifier(defender.Stats.STR)
		if defRoll >= attacker.AC {
			dmg := engine.RollDice(1, 8) + engine.Modifier(defender.Stats.STR)
			if dmg < 1 { dmg = 1 }
			attackerHP -= dmg
			rounds = append(rounds, PvPResult{
				AttackerName:  defender.Name,
				DefenderName:  attacker.Name,
				AttackerRoll:  defRoll,
				Damage:        dmg,
				AttackerHP:    defenderHP,
				DefenderHP:    attackerHP,
				AttackerMaxHP: defender.MaxHP,
				DefenderMaxHP: attacker.MaxHP,
				Narrative:     fmt.Sprintf("%s retaliates, hitting %s for %d damage!", defender.Name, attacker.Name, dmg),
			})
		} else {
			rounds = append(rounds, PvPResult{
				AttackerName:  defender.Name,
				DefenderName:  attacker.Name,
				AttackerRoll:  defRoll,
				AttackerHP:    defenderHP,
				DefenderHP:    attackerHP,
				AttackerMaxHP: defender.MaxHP,
				DefenderMaxHP: attacker.MaxHP,
				Narrative:     fmt.Sprintf("%s counter-attacks but %s dodges!", defender.Name, attacker.Name),
			})
		}

		if attackerHP <= 0 {
			rounds[len(rounds)-1].Winner = defender.Name
			break
		}
	}

	// Determine winner if no KO
	winner := ""
	if attackerHP <= 0 {
		winner = defender.Name
	} else if defenderHP <= 0 {
		winner = attacker.Name
	} else if attackerHP > defenderHP {
		winner = attacker.Name
	} else {
		winner = defender.Name
	}

	// Apply results
	if attackerHP < 1 { attackerHP = 1 }
	if defenderHP < 1 { defenderHP = 1 }
	attacker.HP = attackerHP
	defender.HP = defenderHP

	// Attacker always gets flagged (purple) for initiating PvP
	attacker.SetFlagged()

	// Karma based on L2 PK system
	var droppedItems []engine.Item
	if winner == attacker.Name {
		// Attacker won — check defender's state
		if !defender.IsPurple() && !defender.IsRed() {
			// Defender was WHITE (innocent) — attacker gets PK karma
			attacker.AddPKKarma(1000)
		} else {
			// Defender was PURPLE or RED — consensual PvP, no karma penalty
			attacker.AddPvPKill()
		}
		// Roll item drop on loser (defender) based on defender's karma
		droppedItems = engine.RollItemDrop(defender)
	} else {
		// Defender won — defender was attacked, so they fought back legitimately.
		// Defender never gets PK karma for killing their attacker.
		defender.AddPvPKill()
		// Roll item drop on loser (attacker) based on attacker's karma
		droppedItems = engine.RollItemDrop(attacker)
	}

	// Gold reward: loser pays 10% of their gold to winner
	reward := 0
	if winner == attacker.Name {
		reward = defender.Gold / 10
		defender.Gold -= reward
		attacker.Gold += reward
	} else {
		reward = attacker.Gold / 10
		attacker.Gold -= reward
		defender.Gold += reward
	}

	// Save both
	storage.SaveCharacter(ctx, nk, attackerID, attacker)
	storage.SaveCharacter(ctx, nk, req.DefenderUserID, defender)

	data, _ := json.Marshal(map[string]interface{}{
		"rounds":          rounds,
		"winner":          winner,
		"gold_reward":     reward,
		"attacker_hp":     attackerHP,
		"defender_hp":     defenderHP,
		"dropped_items":   droppedItems,
		"attacker_karma":  attacker.Karma,
		"defender_karma":  defender.Karma,
	})
	return string(data), nil
}

// RPCCoopHelp lets a player boost another player's stats temporarily (blessing).
func RPCCoopHelp(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	helperID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || helperID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		HelperCharID  string `json:"helper_char_id"`
		TargetUserID  string `json:"target_user_id"`
		TargetCharID  string `json:"target_char_id"`
		HelpType      string `json:"help_type"` // "heal", "buff_str", "buff_ac"
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	helper, err := storage.LoadCharacter(ctx, nk, helperID, req.HelperCharID)
	if err != nil || helper == nil {
		return "", runtime.NewError("helper character not found", 5)
	}

	target, err := storage.LoadCharacter(ctx, nk, req.TargetUserID, req.TargetCharID)
	if err != nil || target == nil {
		return "", runtime.NewError("target character not found", 5)
	}

	// Must be in same region
	if helper.RegionX != target.RegionX || helper.RegionY != target.RegionY {
		return "", runtime.NewError("players must be in the same region", 3)
	}

	var result string

	switch req.HelpType {
	case "heal":
		// Costs 5 mana from helper
		if helper.Mana < 5 {
			return "", runtime.NewError("not enough mana to heal", 3)
		}
		helper.Mana -= 5
		amount := engine.RollDice(2, 6) + 2
		healed := target.Heal(amount)
		result = fmt.Sprintf("%s heals %s for %d HP! (Cost: 5 mana)", helper.Name, target.Name, healed)

	case "buff_str":
		if helper.Mana < 3 {
			return "", runtime.NewError("not enough mana", 3)
		}
		// Cap: STR can only be buffed +4 above starting value to prevent exploit
		maxSTR := engine.StartingStats(target.Class).STR + target.Level + 4
		if target.Stats.STR >= maxSTR {
			return "", runtime.NewError("target STR already at buff cap", 3)
		}
		helper.Mana -= 3
		target.Stats.STR += 2
		if target.Stats.STR > maxSTR {
			target.Stats.STR = maxSTR
		}
		target.RecalcDerived()
		result = fmt.Sprintf("%s empowers %s with +2 STR! (STR: %d)", helper.Name, target.Name, target.Stats.STR)

	case "buff_ac":
		if helper.Mana < 3 {
			return "", runtime.NewError("not enough mana", 3)
		}
		// Cap: AC can only be buffed +4 above base
		baseAC := 10 + engine.Modifier(target.Stats.DEX) + target.Equipment.TotalACBonus()
		maxAC := baseAC + 4
		if target.AC >= maxAC {
			return "", runtime.NewError("target AC already at buff cap", 3)
		}
		helper.Mana -= 3
		target.AC += 2
		if target.AC > maxAC {
			target.AC = maxAC
		}
		result = fmt.Sprintf("%s shields %s with +2 AC! (AC: %d)", helper.Name, target.Name, target.AC)

	default:
		return "", runtime.NewError("invalid help type: heal, buff_str, buff_ac", 3)
	}

	storage.SaveCharacter(ctx, nk, helperID, helper)
	storage.SaveCharacter(ctx, nk, req.TargetUserID, target)

	data, _ := json.Marshal(map[string]string{"result": result})
	return string(data), nil
}
