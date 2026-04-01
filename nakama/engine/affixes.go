package engine

import (
	"fmt"
	"math/rand"
)

// ItemModType is the type of stat a modifier affects.
type ItemModType string

const (
	ModDamageFlat     ItemModType = "damage_flat"      // +X damage
	ModDamagePercent  ItemModType = "damage_percent"    // +X% damage
	ModACFlat         ItemModType = "ac_flat"           // +X armor
	ModHP             ItemModType = "hp"                // +X max HP
	ModMana           ItemModType = "mana"              // +X max Mana
	ModSTR            ItemModType = "str"               // +X STR
	ModDEX            ItemModType = "dex"               // +X DEX
	ModCON            ItemModType = "con"               // +X CON
	ModINT            ItemModType = "int"               // +X INT
	ModWIS            ItemModType = "wis"               // +X WIS
	ModCHA            ItemModType = "cha"               // +X CHA
	ModLifeSteal      ItemModType = "life_steal"        // X% life steal
	ModCritChance     ItemModType = "crit_chance"       // +X% crit chance
	ModCritDamage     ItemModType = "crit_damage"       // +X% crit damage
	ModResistFire     ItemModType = "resist_fire"       // +X% fire resist
	ModResistIce      ItemModType = "resist_ice"        // +X% ice resist
	ModResistLightning ItemModType = "resist_lightning"  // +X% lightning resist
	ModResistAll      ItemModType = "resist_all"        // +X% all resist
	ModGoldFind       ItemModType = "gold_find"         // +X% gold find
	ModXPBonus        ItemModType = "xp_bonus"          // +X% XP bonus
	ModAttackSpeed    ItemModType = "attack_speed"      // +X% attack speed
	ModMoveSpeed      ItemModType = "move_speed"        // +X% movement speed
)

// ItemMod is a single modifier on an item.
type ItemMod struct {
	Type  ItemModType `json:"type"`
	Value int         `json:"value"`
	Label string      `json:"label"` // e.g. "+5 Strength"
}

// Affix is a prefix or suffix that can roll on an item.
type Affix struct {
	Name     string      // Display name (e.g. "Cruel", "of the Bear")
	Type     ItemModType
	MinValue int
	MaxValue int
	MinLevel int         // Minimum item level to roll this affix
	IsPrefix bool
}

// --- Affix pools ---

var weaponPrefixes = []Affix{
	// Damage
	{Name: "Sharp", Type: ModDamageFlat, MinValue: 1, MaxValue: 3, MinLevel: 1, IsPrefix: true},
	{Name: "Fine", Type: ModDamageFlat, MinValue: 2, MaxValue: 5, MinLevel: 3, IsPrefix: true},
	{Name: "Cruel", Type: ModDamageFlat, MinValue: 4, MaxValue: 8, MinLevel: 5, IsPrefix: true},
	{Name: "Merciless", Type: ModDamageFlat, MinValue: 6, MaxValue: 12, MinLevel: 8, IsPrefix: true},
	// Crit
	{Name: "Keen", Type: ModCritChance, MinValue: 5, MaxValue: 10, MinLevel: 2, IsPrefix: true},
	{Name: "Deadly", Type: ModCritChance, MinValue: 10, MaxValue: 20, MinLevel: 5, IsPrefix: true},
	// Elemental (add flat damage, not resistance)
	{Name: "Fiery", Type: ModDamageFlat, MinValue: 3, MaxValue: 8, MinLevel: 3, IsPrefix: true},
	{Name: "Frozen", Type: ModDamageFlat, MinValue: 3, MaxValue: 8, MinLevel: 3, IsPrefix: true},
	{Name: "Shocking", Type: ModDamageFlat, MinValue: 3, MaxValue: 8, MinLevel: 3, IsPrefix: true},
	// Life steal
	{Name: "Leeching", Type: ModLifeSteal, MinValue: 3, MaxValue: 6, MinLevel: 4, IsPrefix: true},
	{Name: "Vampiric", Type: ModLifeSteal, MinValue: 5, MaxValue: 10, MinLevel: 7, IsPrefix: true},
}

var weaponSuffixes = []Affix{
	// Stats
	{Name: "of Strength", Type: ModSTR, MinValue: 1, MaxValue: 3, MinLevel: 1},
	{Name: "of the Bear", Type: ModSTR, MinValue: 3, MaxValue: 6, MinLevel: 5},
	{Name: "of the Titan", Type: ModSTR, MinValue: 5, MaxValue: 10, MinLevel: 8},
	{Name: "of Dexterity", Type: ModDEX, MinValue: 1, MaxValue: 3, MinLevel: 1},
	{Name: "of the Fox", Type: ModDEX, MinValue: 3, MaxValue: 6, MinLevel: 5},
	{Name: "of the Wind", Type: ModDEX, MinValue: 5, MaxValue: 10, MinLevel: 8},
	// HP/Mana
	{Name: "of Life", Type: ModHP, MinValue: 5, MaxValue: 15, MinLevel: 1},
	{Name: "of Vitality", Type: ModHP, MinValue: 10, MaxValue: 30, MinLevel: 4},
	{Name: "of the Whale", Type: ModHP, MinValue: 20, MaxValue: 50, MinLevel: 7},
	{Name: "of Energy", Type: ModMana, MinValue: 5, MaxValue: 15, MinLevel: 1},
	{Name: "of Brilliance", Type: ModMana, MinValue: 10, MaxValue: 30, MinLevel: 4},
	// Speed
	{Name: "of Speed", Type: ModAttackSpeed, MinValue: 10, MaxValue: 20, MinLevel: 3},
	{Name: "of Quickness", Type: ModAttackSpeed, MinValue: 20, MaxValue: 40, MinLevel: 6},
	// XP/Gold
	{Name: "of Wealth", Type: ModGoldFind, MinValue: 10, MaxValue: 30, MinLevel: 2},
	{Name: "of Knowledge", Type: ModXPBonus, MinValue: 5, MaxValue: 15, MinLevel: 3},
}

var armorPrefixes = []Affix{
	{Name: "Sturdy", Type: ModACFlat, MinValue: 1, MaxValue: 2, MinLevel: 1, IsPrefix: true},
	{Name: "Reinforced", Type: ModACFlat, MinValue: 2, MaxValue: 4, MinLevel: 3, IsPrefix: true},
	{Name: "Fortified", Type: ModACFlat, MinValue: 3, MaxValue: 6, MinLevel: 5, IsPrefix: true},
	{Name: "Indestructible", Type: ModACFlat, MinValue: 5, MaxValue: 8, MinLevel: 8, IsPrefix: true},
	{Name: "Blessed", Type: ModResistAll, MinValue: 5, MaxValue: 10, MinLevel: 4, IsPrefix: true},
	{Name: "Warding", Type: ModResistAll, MinValue: 10, MaxValue: 20, MinLevel: 7, IsPrefix: true},
}

var armorSuffixes = []Affix{
	{Name: "of Health", Type: ModHP, MinValue: 5, MaxValue: 20, MinLevel: 1},
	{Name: "of the Mammoth", Type: ModHP, MinValue: 20, MaxValue: 50, MinLevel: 5},
	{Name: "of the Colossus", Type: ModHP, MinValue: 40, MaxValue: 80, MinLevel: 8},
	{Name: "of Endurance", Type: ModCON, MinValue: 1, MaxValue: 3, MinLevel: 1},
	{Name: "of the Oak", Type: ModCON, MinValue: 3, MaxValue: 6, MinLevel: 4},
	{Name: "of Wisdom", Type: ModWIS, MinValue: 1, MaxValue: 3, MinLevel: 1},
	{Name: "of the Sage", Type: ModWIS, MinValue: 3, MaxValue: 6, MinLevel: 5},
	{Name: "of Deflection", Type: ModCritDamage, MinValue: -10, MaxValue: -5, MinLevel: 3}, // reduces crit damage taken
	{Name: "of Swiftness", Type: ModMoveSpeed, MinValue: 10, MaxValue: 20, MinLevel: 2},
}

// --- Item quality tiers (D2 style) ---

type ItemQuality string

const (
	QualityNormal ItemQuality = "normal"  // White — no affixes
	QualityMagic  ItemQuality = "magic"   // Blue — 1 prefix and/or 1 suffix
	QualityRare   ItemQuality = "rare"    // Yellow — 2-3 prefixes, 1-2 suffixes
	QualityUnique ItemQuality = "unique"  // Gold — fixed special properties
)

// QualityColor returns display color for the quality.
func QualityColor(q ItemQuality) string {
	switch q {
	case QualityNormal:
		return "#aaaaaa"
	case QualityMagic:
		return "#4169e1"
	case QualityRare:
		return "#ffd700"
	case QualityUnique:
		return "#8b6914"
	}
	return "#aaaaaa"
}

// RollItemQuality determines item quality based on level and luck.
func RollItemQuality(level int) ItemQuality {
	roll := Roll(100)
	uniqueChance := 2 + level/3  // 2% at lv1, ~5% at lv10
	rareChance := 8 + level*2    // 10% at lv1, ~28% at lv10
	magicChance := 30 + level*3  // 33% at lv1, ~60% at lv10

	if roll <= uniqueChance {
		return QualityUnique
	}
	if roll <= uniqueChance+rareChance {
		return QualityRare
	}
	if roll <= uniqueChance+rareChance+magicChance {
		return QualityMagic
	}
	return QualityNormal
}

// ApplyAffixes generates and applies random affixes to an item based on quality.
func ApplyAffixes(item *Item, quality ItemQuality, level int) {
	item.Quality = quality

	if quality == QualityNormal {
		return
	}

	isWeapon := item.Type == ItemTypeWeapon
	var prefixes, suffixes []Affix
	if isWeapon {
		prefixes = filterByLevel(weaponPrefixes, level)
		suffixes = filterByLevel(weaponSuffixes, level)
	} else {
		prefixes = filterByLevel(armorPrefixes, level)
		suffixes = filterByLevel(armorSuffixes, level)
	}

	var numPrefixes, numSuffixes int
	switch quality {
	case QualityMagic:
		// 1 prefix and/or 1 suffix (at least 1)
		if Roll(2) == 1 {
			numPrefixes = 1
		}
		if numPrefixes == 0 || Roll(2) == 1 {
			numSuffixes = 1
		}
	case QualityRare:
		numPrefixes = 1 + Roll(2) // 1-2
		numSuffixes = 1 + Roll(2) // 1-2
	case QualityUnique:
		numPrefixes = 2
		numSuffixes = 2
	}

	// Roll prefixes
	usedTypes := map[ItemModType]bool{}
	for i := 0; i < numPrefixes && len(prefixes) > 0; i++ {
		affix := pickAffix(prefixes, usedTypes)
		if affix == nil {
			break
		}
		value := affix.MinValue + rand.Intn(affix.MaxValue-affix.MinValue+1)
		mod := ItemMod{
			Type:  affix.Type,
			Value: value,
			Label: formatModLabel(affix.Type, value),
		}
		item.Mods = append(item.Mods, mod)
		usedTypes[affix.Type] = true

		// Prepend name
		if i == 0 {
			item.Name = affix.Name + " " + item.Name
		}
	}

	// Roll suffixes
	for i := 0; i < numSuffixes && len(suffixes) > 0; i++ {
		affix := pickAffix(suffixes, usedTypes)
		if affix == nil {
			break
		}
		value := affix.MinValue + rand.Intn(affix.MaxValue-affix.MinValue+1)
		mod := ItemMod{
			Type:  affix.Type,
			Value: value,
			Label: formatModLabel(affix.Type, value),
		}
		item.Mods = append(item.Mods, mod)
		usedTypes[affix.Type] = true

		// Append suffix name
		if i == 0 {
			item.Name = item.Name + " " + affix.Name
		}
	}

	// Update description with mods
	item.Description = buildModDescription(item)
	// Scale value by quality
	switch quality {
	case QualityMagic:
		item.Value = item.Value * 3
	case QualityRare:
		item.Value = item.Value * 8
	case QualityUnique:
		item.Value = item.Value * 15
	}
}

func filterByLevel(affixes []Affix, level int) []Affix {
	var result []Affix
	for _, a := range affixes {
		if level >= a.MinLevel {
			result = append(result, a)
		}
	}
	return result
}

func pickAffix(pool []Affix, usedTypes map[ItemModType]bool) *Affix {
	// Shuffle and pick first unused type
	indices := rand.Perm(len(pool))
	for _, i := range indices {
		if !usedTypes[pool[i].Type] {
			return &pool[i]
		}
	}
	return nil
}

func formatModLabel(modType ItemModType, value int) string {
	sign := "+"
	if value < 0 {
		sign = ""
	}
	switch modType {
	case ModDamageFlat:
		return fmt.Sprintf("%s%d Damage", sign, value)
	case ModDamagePercent:
		return fmt.Sprintf("%s%d%% Damage", sign, value)
	case ModACFlat:
		return fmt.Sprintf("%s%d Armor", sign, value)
	case ModHP:
		return fmt.Sprintf("%s%d Max HP", sign, value)
	case ModMana:
		return fmt.Sprintf("%s%d Max Mana", sign, value)
	case ModSTR:
		return fmt.Sprintf("%s%d Strength", sign, value)
	case ModDEX:
		return fmt.Sprintf("%s%d Dexterity", sign, value)
	case ModCON:
		return fmt.Sprintf("%s%d Constitution", sign, value)
	case ModINT:
		return fmt.Sprintf("%s%d Intelligence", sign, value)
	case ModWIS:
		return fmt.Sprintf("%s%d Wisdom", sign, value)
	case ModCHA:
		return fmt.Sprintf("%s%d Charisma", sign, value)
	case ModLifeSteal:
		return fmt.Sprintf("%s%d%% Life Steal", sign, value)
	case ModCritChance:
		return fmt.Sprintf("%s%d%% Critical Chance", sign, value)
	case ModCritDamage:
		return fmt.Sprintf("%s%d%% Critical Damage", sign, value)
	case ModResistFire:
		return fmt.Sprintf("%s%d%% Fire Resist", sign, value)
	case ModResistIce:
		return fmt.Sprintf("%s%d%% Cold Resist", sign, value)
	case ModResistLightning:
		return fmt.Sprintf("%s%d%% Lightning Resist", sign, value)
	case ModResistAll:
		return fmt.Sprintf("%s%d%% All Resist", sign, value)
	case ModGoldFind:
		return fmt.Sprintf("%s%d%% Gold Find", sign, value)
	case ModXPBonus:
		return fmt.Sprintf("%s%d%% XP Bonus", sign, value)
	case ModAttackSpeed:
		return fmt.Sprintf("%s%d%% Attack Speed", sign, value)
	case ModMoveSpeed:
		return fmt.Sprintf("%s%d%% Movement Speed", sign, value)
	}
	return fmt.Sprintf("%s%d %s", sign, value, modType)
}

func buildModDescription(item *Item) string {
	desc := ""
	if item.DamageDice > 0 {
		desc = fmt.Sprintf("%dd%d damage", item.DamageCount, item.DamageDice)
	}
	if item.ACBonus > 0 {
		if desc != "" {
			desc += ", "
		}
		desc += fmt.Sprintf("+%d AC", item.ACBonus)
	}
	for _, mod := range item.Mods {
		desc += "\n" + mod.Label
	}
	return desc
}
