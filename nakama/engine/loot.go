package engine

import "fmt"

// Rarity tiers
type Rarity string

const (
	RarityCommon   Rarity = "common"
	RarityUncommon Rarity = "uncommon"
	RarityRare     Rarity = "rare"
	RarityEpic     Rarity = "epic"
)

// RarityColor returns a display color hint for the UI.
func RarityColor(r Rarity) string {
	switch r {
	case RarityCommon:
		return "#aaaaaa"
	case RarityUncommon:
		return "#2ecc71"
	case RarityRare:
		return "#3498db"
	case RarityEpic:
		return "#9b59b6"
	}
	return "#aaaaaa"
}

// LootTable defines possible drops for an enemy type.
type LootTable struct {
	GoldMin    int
	GoldMax    int
	GoldChance int // percentage
	Drops      []LootDrop
}

// LootDrop is a single possible item drop.
type LootDrop struct {
	Chance int // percentage (1-100)
	Item   func(level int) Item
}

// RollLoot generates loot from a loot table based on player level.
func RollLoot(table *LootTable, playerLevel int) (int, []Item) {
	gold := 0
	var items []Item

	// Gold
	if Roll(100) <= table.GoldChance {
		gold = table.GoldMin + Roll(table.GoldMax-table.GoldMin+1) - 1
		// Scale gold with level
		gold = gold * (100 + playerLevel*15) / 100
	}

	// Items
	for _, drop := range table.Drops {
		if Roll(100) <= drop.Chance {
			item := drop.Item(playerLevel)
			// Apply D2-style affixes to equipment (not consumables)
			if item.Type == ItemTypeWeapon || item.Type == ItemTypeArmor {
				quality := RollItemQuality(playerLevel)
				ApplyAffixes(&item, quality, playerLevel)
			}
			items = append(items, item)
		}
	}

	return gold, items
}

// GenerateLootFromEnemies generates loot for all defeated enemies using proper loot tables.
func GenerateLootFromEnemies(enemies []Enemy, playerLevel int) (int, []Item) {
	totalGold := 0
	var allItems []Item

	for _, e := range enemies {
		if e.HP > 0 {
			continue
		}
		table := GetLootTable(e.ID)
		gold, items := RollLoot(table, playerLevel)
		totalGold += gold
		allItems = append(allItems, items...)
	}

	return totalGold, allItems
}

// --- Loot Tables per enemy ---

func GetLootTable(enemyID string) *LootTable {
	switch enemyID {
	case "goblin_1":
		return &goblinScoutLoot
	case "goblin_2":
		return &goblinWarriorLoot
	case "goblin_3":
		return &goblinArcherLoot
	case "goblin_boss":
		return &goblinChieftainLoot
	case "goblin_shaman":
		return &goblinShamanLoot
	default:
		return &defaultLoot
	}
}

var defaultLoot = LootTable{
	GoldMin: 3, GoldMax: 10, GoldChance: 50,
	Drops: []LootDrop{
		{Chance: 20, Item: genHealthPotion},
	},
}

var goblinScoutLoot = LootTable{
	GoldMin: 5, GoldMax: 15, GoldChance: 60,
	Drops: []LootDrop{
		{Chance: 30, Item: genHealthPotion},
		{Chance: 10, Item: genCommonWeapon},
	},
}

var goblinWarriorLoot = LootTable{
	GoldMin: 10, GoldMax: 25, GoldChance: 70,
	Drops: []LootDrop{
		{Chance: 25, Item: genHealthPotion},
		{Chance: 20, Item: genUncommonWeapon},
		{Chance: 10, Item: genCommonArmor},
	},
}

var goblinArcherLoot = LootTable{
	GoldMin: 8, GoldMax: 20, GoldChance: 65,
	Drops: []LootDrop{
		{Chance: 25, Item: genHealthPotion},
		{Chance: 15, Item: genCommonWeapon},
		{Chance: 10, Item: genManaPotion},
	},
}

var goblinChieftainLoot = LootTable{
	GoldMin: 30, GoldMax: 75, GoldChance: 100, // Always drops gold
	Drops: []LootDrop{
		{Chance: 100, Item: genRareWeapon},    // Guaranteed rare weapon
		{Chance: 50, Item: genUncommonArmor},   // 50% uncommon armor
		{Chance: 40, Item: genHealthPotion},
		{Chance: 30, Item: genManaPotion},
		{Chance: 10, Item: genEpicAccessory},   // 10% epic ring/amulet
	},
}

var goblinShamanLoot = LootTable{
	GoldMin: 15, GoldMax: 40, GoldChance: 80,
	Drops: []LootDrop{
		{Chance: 50, Item: genManaPotion},
		{Chance: 30, Item: genRareStaff},
		{Chance: 20, Item: genUncommonArmor},
	},
}

// --- Item generators (level-scaled) ---

func genHealthPotion(level int) Item {
	return HealthPotion()
}

func genManaPotion(level int) Item {
	return ManaPotion()
}

func genCommonWeapon(level int) Item {
	weapons := []struct {
		name  string
		dice  int
		count int
		slot  EquipSlot
	}{
		{"Rusty Sword", 6, 1, SlotWeapon},
		{"Worn Dagger", 4, 1, SlotWeapon},
		{"Cracked Mace", 6, 1, SlotWeapon},
		{"Bent Short Bow", 4, 1, SlotWeapon},
	}
	w := weapons[Roll(len(weapons))-1]
	baseDmg := w.dice + level/3
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "cw", Roll(99999)),
		Name:        w.name,
		Description: fmt.Sprintf("A common weapon. %dd%d damage.", w.count, baseDmg),
		Type:        ItemTypeWeapon,
		Slot:        w.slot,
		DamageCount: w.count,
		DamageDice:  baseDmg,
		DamageType:  DamagePhysical,
		Weight:      2,
		Value:       5 + level*2,
	}
}

func genUncommonWeapon(level int) Item {
	weapons := []struct {
		name string
		dice int
	}{
		{"Fine Steel Sword", 8},
		{"Balanced War Axe", 8},
		{"Elven Short Blade", 6},
		{"Tempered Mace", 8},
	}
	w := weapons[Roll(len(weapons))-1]
	baseDmg := w.dice + level/2
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "uw", Roll(99999)),
		Name:        fmt.Sprintf("%s +%d", w.name, 1+level/4),
		Description: fmt.Sprintf("An uncommon weapon. 1d%d damage.", baseDmg),
		Type:        ItemTypeWeapon,
		Slot:        SlotWeapon,
		DamageCount: 1,
		DamageDice:  baseDmg,
		DamageType:  DamagePhysical,
		Weight:      3,
		Value:       15 + level*5,
	}
}

func genRareWeapon(level int) Item {
	weapons := []struct {
		name   string
		dice   int
		dmgType DamageType
	}{
		{"Goblin Slayer", 10, DamagePhysical},
		{"Flamebrand", 8, DamageFire},
		{"Frostbite Blade", 8, DamageIce},
		{"Thunderstrike Hammer", 10, DamageLightning},
	}
	w := weapons[Roll(len(weapons))-1]
	baseDmg := w.dice + level/2
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "rw", Roll(99999)),
		Name:        w.name,
		Description: fmt.Sprintf("A rare weapon! 1d%d %s damage.", baseDmg, w.dmgType),
		Type:        ItemTypeWeapon,
		Slot:        SlotWeapon,
		DamageCount: 1,
		DamageDice:  baseDmg,
		DamageType:  w.dmgType,
		Weight:      3,
		Value:       40 + level*10,
	}
}

func genRareStaff(level int) Item {
	staves := []struct {
		name    string
		dmgType DamageType
	}{
		{"Staff of Embers", DamageFire},
		{"Frostweave Staff", DamageIce},
		{"Thundercaller", DamageLightning},
		{"Holy Rod", DamageHoly},
	}
	s := staves[Roll(len(staves))-1]
	baseDmg := 6 + level/2
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "rs", Roll(99999)),
		Name:        s.name,
		Description: fmt.Sprintf("A rare staff! 1d%d %s damage.", baseDmg, s.dmgType),
		Type:        ItemTypeWeapon,
		Slot:        SlotWeapon,
		DamageCount: 1,
		DamageDice:  baseDmg,
		DamageType:  s.dmgType,
		Weight:      2,
		Value:       35 + level*8,
	}
}

func genCommonArmor(level int) Item {
	armors := []struct {
		name string
		ac   int
		slot EquipSlot
	}{
		{"Patched Leather Vest", 1, SlotArmor},
		{"Rusty Iron Helm", 1, SlotHelmet},
		{"Worn Boots", 1, SlotBoots},
		{"Wooden Shield", 1, SlotOffhand},
	}
	a := armors[Roll(len(armors))-1]
	bonus := a.ac + level/4
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "ca", Roll(99999)),
		Name:        a.name,
		Description: fmt.Sprintf("Common armor. +%d AC.", bonus),
		Type:        ItemTypeArmor,
		Slot:        a.slot,
		ACBonus:     bonus,
		Weight:      4,
		Value:       8 + level*3,
	}
}

func genUncommonArmor(level int) Item {
	armors := []struct {
		name string
		ac   int
		slot EquipSlot
	}{
		{"Reinforced Chain Shirt", 3, SlotArmor},
		{"Steel Helm", 2, SlotHelmet},
		{"Iron-shod Boots", 2, SlotBoots},
		{"Kite Shield", 2, SlotOffhand},
	}
	a := armors[Roll(len(armors))-1]
	bonus := a.ac + level/3
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "ua", Roll(99999)),
		Name:        fmt.Sprintf("%s +%d", a.name, 1+level/5),
		Description: fmt.Sprintf("Uncommon armor. +%d AC.", bonus),
		Type:        ItemTypeArmor,
		Slot:        a.slot,
		ACBonus:     bonus,
		Weight:      5,
		Value:       20 + level*6,
	}
}

func genEpicAccessory(level int) Item {
	accessories := []struct {
		name string
		slot EquipSlot
		desc string
	}{
		{"Ring of Vitality", SlotRing1, "Grants vigor to the wearer"},
		{"Amulet of Shadows", SlotAmulet, "Shrouds the wearer in darkness"},
		{"Band of the Chieftain", SlotRing2, "Taken from a goblin king"},
	}
	a := accessories[Roll(len(accessories))-1]
	acBonus := 1 + level/3
	return Item{
		ID:          fmt.Sprintf("loot_%s_%d", "ea", Roll(99999)),
		Name:        a.name,
		Description: fmt.Sprintf("Epic! %s +%d AC.", a.desc, acBonus),
		Type:        ItemTypeArmor,
		Slot:        a.slot,
		ACBonus:     acBonus,
		Weight:      0.5,
		Value:       50 + level*15,
	}
}

// FormatLootNotification creates a narrative-friendly loot summary.
func FormatLootNotification(gold int, items []Item) string {
	if gold == 0 && len(items) == 0 {
		return "No loot found."
	}

	text := "Loot found:\n"
	if gold > 0 {
		text += fmt.Sprintf("  +%d gold\n", gold)
	}
	for _, item := range items {
		quality := item.Quality
		if quality == "" {
			quality = QualityNormal
		}
		qualityLabel := string(quality)
		text += fmt.Sprintf("  [%s] %s", qualityLabel, item.Name)
		if item.DamageDice > 0 {
			text += fmt.Sprintf(" (%dd%d dmg)", item.DamageCount, item.DamageDice)
		}
		if item.ACBonus > 0 {
			text += fmt.Sprintf(" (+%d AC)", item.ACBonus)
		}
		text += "\n"
		for _, mod := range item.Mods {
			text += fmt.Sprintf("    %s\n", mod.Label)
		}
	}
	return text
}

// ItemRarity guesses rarity from item value.
func ItemRarity(item Item) Rarity {
	if item.Type == ItemTypeConsumable {
		return RarityCommon
	}
	if item.Value >= 50 {
		return RarityEpic
	}
	if item.Value >= 30 {
		return RarityRare
	}
	if item.Value >= 15 {
		return RarityUncommon
	}
	return RarityCommon
}
