package claude

import (
	"fmt"
	"math/rand"

	"realmweaver/engine"
)

// NarrationTier determines which AI tier (or template) to use.
type NarrationTier int

const (
	TierTemplate NarrationTier = iota // No AI call — use pre-written template
	TierHaiku                         // Fast/cheap model for routine narration
	TierSonnet                        // Full model for important moments
)

// ClassifyAction determines which narration tier to use for a given action.
func ClassifyAction(action string, phase engine.GamePhase) NarrationTier {
	switch phase {
	case engine.PhaseInCombat:
		switch action {
		case "attack", "defend", "flee", "use_item", "enemy_attack":
			return TierTemplate // Routine combat — templates
		case "combat_start", "combat_victory", "player_death":
			return TierHaiku // Dramatic moments — cheap AI
		}
	case engine.PhaseInDialogue:
		return TierSonnet // NPC dialogue always uses full AI
	case engine.PhaseExploring:
		switch action {
		case "rest":
			return TierTemplate
		case "look", "search":
			return TierHaiku
		default:
			return TierHaiku // Location visits use cheap AI
		}
	}
	return TierHaiku
}

// TemplateNarrate returns a pre-written narration for routine actions,
// selecting templates based on language.
func TemplateNarrate(result *engine.ActionResult, character *engine.Character, language string) string {
	templates := getTemplateSet(language)
	var text string

	switch result.Action {
	case "melee_attack":
		if result.Hit {
			text = pickRandom(templates.attackHit, character.Name, result.Target, result.Damage, weaponName(character))
		} else {
			text = pickRandom(templates.attackMiss, character.Name, result.Target, weaponName(character))
		}
	case "enemy_attack":
		if result.Hit {
			text = pickRandom(templates.enemyHit, result.Actor, character.Name, result.Damage)
		} else {
			text = pickRandom(templates.enemyMiss, result.Actor, character.Name)
		}
	case "defend":
		text = pickRandom(templates.defend, character.Name)
	case "flee":
		if result.Success {
			text = pickRandom(templates.fleeSuccess, character.Name)
		} else {
			text = pickRandom(templates.fleeFail, character.Name)
		}
	case "use_item":
		text = result.Details
	case "rest":
		text = pickRandom(templates.rest, character.Name)
	default:
		text = result.Details
	}

	return text
}

// templateSet holds all template pools for a single language.
type templateSet struct {
	attackHit   []string
	attackMiss  []string
	enemyHit    []string
	enemyMiss   []string
	defend      []string
	fleeSuccess []string
	fleeFail    []string
	rest        []string
}

// getTemplateSet returns the template set for the given language, falling back to English.
func getTemplateSet(language string) *templateSet {
	switch language {
	case "uk":
		return &ukTemplates
	default:
		return &enTemplates
	}
}

// enTemplates is the English template set (populated in init()).
var enTemplates templateSet

// ukTemplates is the Ukrainian template set.
var ukTemplates = templateSet{
	attackHit: []string{
		"%[1]s вдаряє %[2]s своєю зброєю %[4]s — влучний удар на %[3]d шкоди!",
	},
	attackMiss: []string{
		"%[1]s замахується на %[2]s, але промахується!",
	},
	enemyHit: []string{
		"%[1]s атакує %[2]s і завдає %[3]d шкоди!",
	},
	enemyMiss: []string{
		"%[1]s атакує %[2]s, але промахується!",
	},
	defend: []string{
		"%s приймає оборонну позицію. (+2 AC)",
	},
	fleeSuccess: []string{
		"%s тікає з бою!",
	},
	fleeFail: []string{
		"%s намагається втекти, але вороги блокують шлях!",
	},
	rest: []string{
		"%s відпочиває. HP та Мана повністю відновлені.",
	},
}

func weaponName(c *engine.Character) string {
	if c.Equipment.Weapon != nil {
		return c.Equipment.Weapon.Name
	}
	return "fists"
}

func pickRandom(templates []string, args ...interface{}) string {
	tmpl := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(tmpl, args...)
}

// init populates the English template set with explicit argument order.
func init() {
	enTemplates = templateSet{
		attackHit: []string{
			"%[1]s strikes %[2]s with their %[4]s — a solid hit for %[3]d damage!",
			"%[1]s lands a blow against %[2]s! The %[4]s connects hard. %[3]d damage dealt.",
			"With a sharp swing, %[1]s catches %[2]s off guard with their %[4]s. %[3]d damage!",
			"%[1]s lunges forward, %[4]s biting into %[2]s. %[3]d damage!",
		},
		attackMiss: []string{
			"%[1]s swings at %[2]s with their %[3]s, but misses!",
			"%[1]s lunges at %[2]s — the %[3]s cuts only air.",
			"A wild swing from %[1]s! %[2]s dodges the %[3]s with ease.",
			"%[1]s overextends, and %[2]s sidesteps the attack.",
		},
		enemyHit: []string{
			"%[1]s slashes at %[2]s, dealing %[3]d damage!",
			"%[1]s strikes! %[2]s takes %[3]d damage.",
			"A vicious blow from %[1]s catches %[2]s — %[3]d damage!",
			"%[1]s lashes out, wounding %[2]s for %[3]d damage.",
		},
		enemyMiss: []string{
			"%[1]s attacks %[2]s but misses!",
			"%[1]s swings wildly — %[2]s dodges!",
			"%[2]s deflects the attack from %[1]s.",
			"%[1]s lunges, but %[2]s is too quick.",
		},
		defend: []string{
			"%s raises their guard, bracing for the next attack. (+2 AC)",
			"%s takes a defensive stance, shield ready. (+2 AC)",
			"%s hunkers down, watching for incoming blows. (+2 AC)",
		},
		fleeSuccess: []string{
			"%s dashes toward the exit — and escapes!",
			"%s breaks away from combat and flees to safety!",
			"With a burst of speed, %s escapes the fight!",
		},
		fleeFail: []string{
			"%s tries to run, but the enemies block the way!",
			"%s stumbles while fleeing — no escape!",
			"The enemies cut off %s's retreat!",
		},
		rest: []string{
			"%s finds a quiet spot and rests. HP and Mana fully restored.",
			"%s sits down, catches their breath, and feels refreshed. Fully healed.",
			"A moment of peace. %s rests and recovers all HP and Mana.",
		},
	}
}

// IsFirstVisit checks if a region narration should use full AI (first time) or cache.
func IsFirstVisit(regionKey string, visitedCache map[string]bool) bool {
	if visitedCache[regionKey] {
		return false
	}
	visitedCache[regionKey] = true
	return true
}
