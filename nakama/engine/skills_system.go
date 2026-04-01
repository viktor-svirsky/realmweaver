package engine

import "fmt"

// Skill represents a class ability that can be used in combat.
type Skill struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ManaCost    int         `json:"mana_cost"`
	MinLevel    int         `json:"min_level"`
	Class       Class       `json:"class"`
	Effect      SkillEffect `json:"effect"`
	TargetType  string      `json:"target_type"` // "self", "enemy", "ally", "all_enemies"
}

// SkillEffect defines the mechanical effects of a skill.
type SkillEffect struct {
	Damage         int        `json:"damage,omitempty"`
	DamageDice     int        `json:"damage_dice,omitempty"`
	DamageCount    int        `json:"damage_count,omitempty"`
	DamageType     DamageType `json:"damage_type,omitempty"`
	Heal           int        `json:"heal,omitempty"`
	HealDice       int        `json:"heal_dice,omitempty"`
	HealCount      int        `json:"heal_count,omitempty"`
	BuffStat       string     `json:"buff_stat,omitempty"`
	BuffAmount     int        `json:"buff_amount,omitempty"`
	BuffDuration   int        `json:"buff_duration,omitempty"`
	Debuff         string     `json:"debuff,omitempty"`
	DebuffDuration int        `json:"debuff_duration,omitempty"`
}

// SkillResult holds the outcome of executing a skill.
type SkillResult struct {
	SkillName    string     `json:"skill_name"`
	Damage       int        `json:"damage,omitempty"`
	DamageType   DamageType `json:"damage_type,omitempty"`
	Healed       int        `json:"healed,omitempty"`
	BuffApplied  string     `json:"buff_applied,omitempty"`
	DebuffApplied string    `json:"debuff_applied,omitempty"`
	Description  string     `json:"description"`
	TargetsHit   int        `json:"targets_hit,omitempty"`
}

// allSkills is the master skill registry.
var allSkills = []Skill{
	// Warrior skills
	{
		ID: "warrior_power_strike", Name: "Power Strike", Description: "A devastating blow with full force",
		ManaCost: 3, MinLevel: 1, Class: ClassWarrior, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 8, DamageType: DamagePhysical},
	},
	{
		ID: "warrior_shield_bash", Name: "Shield Bash", Description: "Bash with your shield, stunning the target",
		ManaCost: 4, MinLevel: 3, Class: ClassWarrior, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 1, DamageDice: 6, DamageType: DamagePhysical, Debuff: "stun", DebuffDuration: 1},
	},
	{
		ID: "warrior_battle_cry", Name: "Battle Cry", Description: "A mighty shout that bolsters your strength",
		ManaCost: 5, MinLevel: 5, Class: ClassWarrior, TargetType: "self",
		Effect: SkillEffect{BuffStat: "str", BuffAmount: 3, BuffDuration: 3},
	},

	// Mage skills
	{
		ID: "mage_fireball", Name: "Fireball", Description: "Hurl a ball of fire at your enemy",
		ManaCost: 5, MinLevel: 1, Class: ClassMage, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 3, DamageDice: 6, DamageType: DamageFire},
	},
	{
		ID: "mage_ice_shard", Name: "Ice Shard", Description: "Launch a freezing shard that slows the target",
		ManaCost: 4, MinLevel: 3, Class: ClassMage, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 6, DamageType: DamageIce, Debuff: "slow", DebuffDuration: 2},
	},
	{
		ID: "mage_lightning_bolt", Name: "Lightning Bolt", Description: "A devastating bolt of lightning",
		ManaCost: 8, MinLevel: 5, Class: ClassMage, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 4, DamageDice: 6, DamageType: DamageLightning},
	},
	{
		ID: "mage_arcane_shield", Name: "Arcane Shield", Description: "Surround yourself with a protective arcane barrier",
		ManaCost: 6, MinLevel: 7, Class: ClassMage, TargetType: "self",
		Effect: SkillEffect{BuffStat: "ac", BuffAmount: 4, BuffDuration: 3},
	},

	// Rogue skills
	{
		ID: "rogue_backstab", Name: "Backstab", Description: "Strike from the shadows for massive damage",
		ManaCost: 3, MinLevel: 1, Class: ClassRogue, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 3, DamageDice: 6, DamageType: DamagePhysical},
	},
	{
		ID: "rogue_poison_blade", Name: "Poison Blade", Description: "Coat your blade in poison, dealing damage over time",
		ManaCost: 4, MinLevel: 3, Class: ClassRogue, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 1, DamageDice: 4, DamageType: DamagePhysical, Debuff: "poison", DebuffDuration: 2},
	},
	{
		ID: "rogue_shadow_step", Name: "Shadow Step", Description: "Meld with shadows, boosting your agility",
		ManaCost: 5, MinLevel: 5, Class: ClassRogue, TargetType: "self",
		Effect: SkillEffect{BuffStat: "dex", BuffAmount: 5, BuffDuration: 2},
	},

	// Cleric skills
	{
		ID: "cleric_holy_light", Name: "Holy Light", Description: "Channel divine energy to heal wounds",
		ManaCost: 4, MinLevel: 1, Class: ClassCleric, TargetType: "self",
		Effect: SkillEffect{HealDice: 8, Heal: 4, HealCount: 2},
	},
	{
		ID: "cleric_smite", Name: "Smite", Description: "Strike with holy wrath",
		ManaCost: 5, MinLevel: 3, Class: ClassCleric, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 8, DamageType: DamageHoly},
	},
	{
		ID: "cleric_divine_shield", Name: "Divine Shield", Description: "Surround yourself with divine protection",
		ManaCost: 6, MinLevel: 5, Class: ClassCleric, TargetType: "self",
		Effect: SkillEffect{BuffStat: "ac", BuffAmount: 5, BuffDuration: 3},
	},
	{
		ID: "cleric_resurrection", Name: "Resurrection", Description: "Divine grace restores you from the brink of death",
		ManaCost: 10, MinLevel: 7, Class: ClassCleric, TargetType: "self",
		Effect: SkillEffect{Heal: 50}, // Special: heals to 50% HP
	},

	// Ranger skills
	{
		ID: "ranger_aimed_shot", Name: "Aimed Shot", Description: "Take careful aim for a precise ranged attack",
		ManaCost: 3, MinLevel: 1, Class: ClassRanger, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 8, DamageType: DamagePhysical},
	},
	{
		ID: "ranger_entangling_roots", Name: "Entangling Roots", Description: "Summon roots to ensnare and slow your enemy",
		ManaCost: 4, MinLevel: 3, Class: ClassRanger, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 1, DamageDice: 4, DamageType: DamagePhysical, Debuff: "slow", DebuffDuration: 2},
	},
	{
		ID: "ranger_rain_of_arrows", Name: "Rain of Arrows", Description: "Unleash a volley of arrows on all enemies",
		ManaCost: 7, MinLevel: 5, Class: ClassRanger, TargetType: "all_enemies",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 6, DamageType: DamagePhysical},
	},

	// Paladin skills
	{
		ID: "paladin_holy_strike", Name: "Holy Strike", Description: "Channel holy energy through your weapon",
		ManaCost: 3, MinLevel: 1, Class: ClassPaladin, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 6, DamageType: DamageHoly},
	},
	{
		ID: "paladin_lay_on_hands", Name: "Lay on Hands", Description: "Heal wounds through the power of faith",
		ManaCost: 5, MinLevel: 3, Class: ClassPaladin, TargetType: "self",
		Effect: SkillEffect{HealDice: 6, Heal: 2, HealCount: 3},
	},
	{
		ID: "paladin_divine_wrath", Name: "Divine Wrath", Description: "Smite your enemy with holy fire and heal from their pain",
		ManaCost: 8, MinLevel: 5, Class: ClassPaladin, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 3, DamageDice: 8, DamageType: DamageHoly},
	},

	// Necromancer skills
	{
		ID: "necromancer_life_drain", Name: "Life Drain", Description: "Siphon life force from your enemy",
		ManaCost: 4, MinLevel: 1, Class: ClassNecromancer, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 6, DamageType: DamageNecrotic},
	},
	{
		ID: "necromancer_bone_shield", Name: "Bone Shield", Description: "Surround yourself with orbiting bones for protection",
		ManaCost: 4, MinLevel: 3, Class: ClassNecromancer, TargetType: "self",
		Effect: SkillEffect{BuffStat: "ac", BuffAmount: 3, BuffDuration: 3},
	},
	{
		ID: "necromancer_soul_rend", Name: "Soul Rend", Description: "Tear at the enemy's very soul",
		ManaCost: 7, MinLevel: 5, Class: ClassNecromancer, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 3, DamageDice: 8, DamageType: DamageNecrotic},
	},
	{
		ID: "necromancer_dark_resurrection", Name: "Dark Resurrection", Description: "Dark power pulls you back from the brink of death",
		ManaCost: 12, MinLevel: 7, Class: ClassNecromancer, TargetType: "self",
		Effect: SkillEffect{Heal: 30}, // Special: heals to 30% HP
	},

	// Berserker skills
	{
		ID: "berserker_reckless_strike", Name: "Reckless Strike", Description: "A wild swing that sacrifices defense for power",
		ManaCost: 2, MinLevel: 1, Class: ClassBerserker, TargetType: "enemy",
		Effect: SkillEffect{DamageCount: 3, DamageDice: 8, DamageType: DamagePhysical, BuffStat: "ac", BuffAmount: -2, BuffDuration: 1},
	},
	{
		ID: "berserker_rage", Name: "Rage", Description: "Enter a berserker rage, gaining strength but losing defense",
		ManaCost: 4, MinLevel: 3, Class: ClassBerserker, TargetType: "self",
		Effect: SkillEffect{BuffStat: "str", BuffAmount: 5, BuffDuration: 3},
	},
	{
		ID: "berserker_whirlwind", Name: "Whirlwind", Description: "Spin in a devastating arc, hitting all nearby enemies",
		ManaCost: 6, MinLevel: 5, Class: ClassBerserker, TargetType: "all_enemies",
		Effect: SkillEffect{DamageCount: 2, DamageDice: 8, DamageType: DamagePhysical},
	},
}

// GetClassSkills returns all skills for a class.
func GetClassSkills(class Class) []Skill {
	var skills []Skill
	for _, s := range allSkills {
		if s.Class == class {
			skills = append(skills, s)
		}
	}
	return skills
}

// GetAvailableSkills returns skills unlocked at the current level for a class.
func GetAvailableSkills(class Class, level int) []Skill {
	var skills []Skill
	for _, s := range allSkills {
		if s.Class == class && s.MinLevel <= level {
			skills = append(skills, s)
		}
	}
	return skills
}

// FindSkillByID returns a skill by its ID, or nil if not found.
func FindSkillByID(id string) *Skill {
	for i := range allSkills {
		if allSkills[i].ID == id {
			return &allSkills[i]
		}
	}
	return nil
}

// ResolveSkill executes a skill and returns the result.
// For "enemy" target, pass a *Enemy. For "all_enemies", pass a []*Enemy slice.
// For "self" target, target is ignored.
func ResolveSkill(character *Character, skill *Skill, target interface{}) SkillResult {
	result := SkillResult{
		SkillName:  skill.Name,
		DamageType: skill.Effect.DamageType,
	}

	switch skill.TargetType {
	case "enemy":
		enemy, ok := target.(*Enemy)
		if !ok || enemy == nil {
			result.Description = "No valid target"
			return result
		}
		_ = enemy // validated above; used in resolveSkillOnEnemy
		deductMana(character, skill.ManaCost)
		result = resolveSkillOnEnemy(character, skill, target, result)

	case "all_enemies":
		enemies, ok := target.([]*Enemy)
		if !ok || len(enemies) == 0 {
			result.Description = "No valid targets"
			return result
		}
		_ = enemies // validated above; used in resolveSkillOnAllEnemies
		deductMana(character, skill.ManaCost)
		result = resolveSkillOnAllEnemies(character, skill, target, result)

	case "self":
		deductMana(character, skill.ManaCost)
		result = resolveSkillOnSelf(character, skill, result)
	}

	return result
}

func deductMana(character *Character, cost int) {
	character.Mana -= cost
	if character.Mana < 0 {
		character.Mana = 0
	}
}

func resolveSkillOnEnemy(character *Character, skill *Skill, target interface{}, result SkillResult) SkillResult {
	enemy := target.(*Enemy)

	// Calculate damage
	if skill.Effect.DamageCount > 0 && skill.Effect.DamageDice > 0 {
		result.Damage = RollDice(skill.Effect.DamageCount, skill.Effect.DamageDice)
		result.Damage += skill.Effect.Damage // flat bonus
		enemy.HP -= result.Damage
		if enemy.HP < 0 {
			enemy.HP = 0
		}
		result.TargetsHit = 1
	}

	// Apply debuff
	if skill.Effect.Debuff != "" {
		result.DebuffApplied = fmt.Sprintf("%s for %d turns", skill.Effect.Debuff, skill.Effect.DebuffDuration)
	}

	// Special: Life Drain heals for half damage
	if skill.ID == "necromancer_life_drain" && result.Damage > 0 {
		healed := result.Damage / 2
		character.Heal(healed)
		result.Healed = healed
	}

	// Special: Divine Wrath heals for half damage
	if skill.ID == "paladin_divine_wrath" && result.Damage > 0 {
		healed := result.Damage / 2
		character.Heal(healed)
		result.Healed = healed
	}

	// Special: Poison Blade adds extra poison damage (2d4)
	if skill.ID == "rogue_poison_blade" {
		poisonDmg := RollDice(2, 4)
		result.Damage += poisonDmg
		enemy.HP -= poisonDmg
		if enemy.HP < 0 {
			enemy.HP = 0
		}
	}

	// Self AC debuff for Reckless Strike
	if skill.Effect.BuffStat == "ac" && skill.Effect.BuffAmount < 0 {
		character.AC += skill.Effect.BuffAmount
		result.BuffApplied = fmt.Sprintf("%+d AC for %d turns", skill.Effect.BuffAmount, skill.Effect.BuffDuration)
	}

	result.Description = buildSkillDescription(skill, result, enemy.Name)
	return result
}

func resolveSkillOnAllEnemies(character *Character, skill *Skill, target interface{}, result SkillResult) SkillResult {
	enemies := target.([]*Enemy)

	totalDamage := 0
	hitCount := 0
	for _, enemy := range enemies {
		if enemy.HP <= 0 {
			continue
		}
		dmg := RollDice(skill.Effect.DamageCount, skill.Effect.DamageDice)
		dmg += skill.Effect.Damage
		enemy.HP -= dmg
		if enemy.HP < 0 {
			enemy.HP = 0
		}
		totalDamage += dmg
		hitCount++
	}

	result.Damage = totalDamage
	result.TargetsHit = hitCount
	result.Description = fmt.Sprintf("%s uses %s, hitting %d enemies for %d total %s damage!",
		character.Name, skill.Name, hitCount, totalDamage, skill.Effect.DamageType)
	return result
}

func resolveSkillOnSelf(character *Character, skill *Skill, result SkillResult) SkillResult {
	// Healing skills
	if skill.Effect.HealDice > 0 {
		healAmount := RollDice(skill.Effect.HealCount, skill.Effect.HealDice) + skill.Effect.Heal
		result.Healed = character.Heal(healAmount)
		result.Description = fmt.Sprintf("%s uses %s, restoring %d HP!", character.Name, skill.Name, result.Healed)
		return result
	}

	// Special resurrection skills (heal to % HP)
	if skill.ID == "cleric_resurrection" {
		healTo := character.MaxHP / 2
		if character.HP < healTo {
			result.Healed = character.Heal(healTo - character.HP)
		}
		result.Description = fmt.Sprintf("%s uses %s, restoring to %d HP!", character.Name, skill.Name, character.HP)
		return result
	}
	if skill.ID == "necromancer_dark_resurrection" {
		healTo := character.MaxHP * 30 / 100
		if character.HP < healTo {
			result.Healed = character.Heal(healTo - character.HP)
		}
		result.Description = fmt.Sprintf("%s uses %s, dark energy restoring to %d HP!", character.Name, skill.Name, character.HP)
		return result
	}

	// Buff skills (buffs last until combat ends — reset via RecalcDerived)
	if skill.Effect.BuffStat != "" {
		result.BuffApplied = fmt.Sprintf("+%d %s for %d turns", skill.Effect.BuffAmount, skill.Effect.BuffStat, skill.Effect.BuffDuration)

		// Apply immediate stat effect, capped to prevent stacking
		base := StartingStats(character.Class)
		maxBuff := skill.Effect.BuffAmount * 2 // allow at most 2x the buff amount above base+level
		switch skill.Effect.BuffStat {
		case "str":
			cap := base.STR + character.Level + maxBuff
			if character.Stats.STR < cap {
				character.Stats.STR += skill.Effect.BuffAmount
				if character.Stats.STR > cap {
					character.Stats.STR = cap
				}
			}
		case "dex":
			cap := base.DEX + character.Level + maxBuff
			if character.Stats.DEX < cap {
				character.Stats.DEX += skill.Effect.BuffAmount
				if character.Stats.DEX > cap {
					character.Stats.DEX = cap
				}
			}
			character.RecalcDerived()
		case "ac":
			baseAC := 10 + Modifier(character.Stats.DEX) + character.Equipment.TotalACBonus()
			cap := baseAC + maxBuff
			if character.AC < cap {
				character.AC += skill.Effect.BuffAmount
				if character.AC > cap {
					character.AC = cap
				}
			}
		}

		result.Description = fmt.Sprintf("%s uses %s! %s", character.Name, skill.Name, result.BuffApplied)
		return result
	}

	result.Description = fmt.Sprintf("%s uses %s!", character.Name, skill.Name)
	return result
}

func buildSkillDescription(skill *Skill, result SkillResult, targetName string) string {
	desc := fmt.Sprintf("%s deals %d %s damage to %s", skill.Name, result.Damage, skill.Effect.DamageType, targetName)
	if result.Healed > 0 {
		desc += fmt.Sprintf(", healing for %d HP", result.Healed)
	}
	if result.DebuffApplied != "" {
		desc += fmt.Sprintf(", applying %s", result.DebuffApplied)
	}
	if result.BuffApplied != "" {
		desc += fmt.Sprintf(" (%s)", result.BuffApplied)
	}
	return desc + "!"
}
