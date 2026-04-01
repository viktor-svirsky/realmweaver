package engine

import (
	"fmt"
	"math/rand"
	"time"
)

// NewCharacter creates a new character with class-appropriate starting stats and equipment.
func NewCharacter(id, name string, class Class) *Character {
	stats := StartingStats(class)
	c := &Character{
		ID:        id,
		Name:      name,
		Class:     class,
		Level:     1,
		XP:        0,
		Stats:     stats,
		Gold:      50,
		Karma:     0,
		Equipment: Equipment{},
		Inventory: []Item{},
		RegionX:   0,
		RegionY:   0,
	}
	c.RecalcDerived()
	c.HP = c.MaxHP
	c.Mana = c.MaxMana

	// Add starter equipment
	starterItems := StarterEquipment(class)
	for i := range starterItems {
		c.Inventory = append(c.Inventory, starterItems[i])
	}
	// Auto-equip weapon
	if len(starterItems) > 0 {
		for i := range c.Inventory {
			if c.Inventory[i].Slot == SlotWeapon {
				item := c.Inventory[i]
				c.Equipment.Weapon = &item
				c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
				break
			}
		}
	}
	// Auto-equip armor
	for i := range c.Inventory {
		if c.Inventory[i].Slot == SlotArmor {
			item := c.Inventory[i]
			c.Equipment.Armor = &item
			c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			break
		}
	}

	// Add health potions
	c.Inventory = append(c.Inventory, HealthPotion())
	c.Inventory = append(c.Inventory, HealthPotion())

	c.RecalcDerived()
	return c
}

// RecalcDerived recalculates MaxHP, MaxMana, and AC from stats and equipment.
func (c *Character) RecalcDerived() {
	// MaxHP: base 10 + CON modifier per level
	conMod := Modifier(c.Stats.CON)
	c.MaxHP = 10 + conMod + (c.Level-1)*(6+conMod)
	if c.MaxHP < 1 {
		c.MaxHP = 1
	}

	// MaxMana: base 5 + INT modifier per level (0 minimum)
	intMod := Modifier(c.Stats.INT)
	c.MaxMana = 5 + intMod + (c.Level-1)*(4+intMod)
	if c.MaxMana < 0 {
		c.MaxMana = 0
	}

	// AC: 10 + DEX modifier + equipment bonus
	dexMod := Modifier(c.Stats.DEX)
	c.AC = 10 + dexMod + c.Equipment.TotalACBonus()
}

// GainXP adds XP and triggers level-up if threshold is reached.
// Returns true if character leveled up.
func (c *Character) GainXP(amount int) bool {
	c.XP += amount
	if c.XP >= XPForLevel(c.Level+1) {
		c.LevelUp()
		return true
	}
	return false
}

// LevelUp advances the character by one level.
func (c *Character) LevelUp() {
	c.Level++

	// Stat boost: +1 to primary stat every level, +1 to secondary every 2 levels
	switch c.Class {
	case ClassWarrior:
		c.Stats.STR++
		if c.Level%2 == 0 {
			c.Stats.CON++
		}
	case ClassMage:
		c.Stats.INT++
		if c.Level%2 == 0 {
			c.Stats.WIS++
		}
	case ClassRogue:
		c.Stats.DEX++
		if c.Level%2 == 0 {
			c.Stats.INT++
		}
	case ClassCleric:
		c.Stats.WIS++
		if c.Level%2 == 0 {
			c.Stats.CON++
		}
	case ClassRanger:
		c.Stats.DEX++
		if c.Level%2 == 0 {
			c.Stats.WIS++
		}
	case ClassPaladin:
		c.Stats.CHA++
		if c.Level%2 == 0 {
			c.Stats.STR++
		}
	case ClassNecromancer:
		c.Stats.INT++
		if c.Level%2 == 0 {
			c.Stats.CON++
		}
	case ClassBerserker:
		c.Stats.STR++
		if c.Level%2 == 0 {
			c.Stats.CON++
		}
	}

	c.RecalcDerived()
	// Heal to full on level-up
	c.HP = c.MaxHP
	c.Mana = c.MaxMana
}

// Heal restores HP up to MaxHP. Returns actual amount healed.
func (c *Character) Heal(amount int) int {
	before := c.HP
	c.HP += amount
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
	return c.HP - before
}

// TakeDamage reduces HP. Returns actual damage taken.
func (c *Character) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}
	before := c.HP
	c.HP -= amount
	if c.HP < 0 {
		c.HP = 0
	}
	return before - c.HP
}

// IsAlive returns true if HP > 0.
func (c *Character) IsAlive() bool {
	return c.HP > 0
}

// Rest fully restores HP and Mana.
func (c *Character) Rest() {
	c.HP = c.MaxHP
	c.Mana = c.MaxMana
}

// Summary returns a concise text summary for Claude context.
func (c *Character) Summary() string {
	weaponName := "unarmed"
	if c.Equipment.Weapon != nil {
		weaponName = c.Equipment.Weapon.Name
	}
	return fmt.Sprintf("%s (Lv%d %s) HP:%d/%d Mana:%d/%d AC:%d Weapon:%s Gold:%d",
		c.Name, c.Level, c.Class, c.HP, c.MaxHP, c.Mana, c.MaxMana, c.AC, weaponName, c.Gold)
}

// StarterEquipment returns starting items for a class.
func StarterEquipment(class Class) []Item {
	switch class {
	case ClassWarrior:
		return []Item{
			{ID: "starter_sword", Name: "Iron Longsword", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 8, DamageType: DamagePhysical, Weight: 3, Value: 15},
			{ID: "starter_chainmail", Name: "Chainmail", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 4, Weight: 8, Value: 30},
		}
	case ClassMage:
		return []Item{
			{ID: "starter_staff", Name: "Wooden Staff", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 6, DamageType: DamagePhysical, Weight: 2, Value: 10},
			{ID: "starter_robes", Name: "Mage Robes", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 1, Weight: 2, Value: 15},
		}
	case ClassRogue:
		return []Item{
			{ID: "starter_daggers", Name: "Twin Daggers", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 2, DamageDice: 4, DamageType: DamagePhysical, Weight: 1, Value: 12},
			{ID: "starter_leather", Name: "Leather Armor", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 2, Weight: 4, Value: 20},
		}
	case ClassCleric:
		return []Item{
			{ID: "starter_mace", Name: "Iron Mace", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 6, DamageType: DamagePhysical, Weight: 3, Value: 12},
			{ID: "starter_scalemail", Name: "Scale Mail", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 3, Weight: 6, Value: 25},
		}
	case ClassRanger:
		return []Item{
			{ID: "starter_longbow", Name: "Longbow", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 8, DamageType: DamagePhysical, Weight: 2, Value: 15},
			{ID: "starter_leather_ranger", Name: "Leather Armor", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 2, Weight: 4, Value: 20},
		}
	case ClassPaladin:
		return []Item{
			{ID: "starter_warhammer", Name: "Warhammer", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 8, DamageType: DamagePhysical, Weight: 4, Value: 18},
			{ID: "starter_plate", Name: "Plate Armor", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 5, Weight: 12, Value: 50},
		}
	case ClassNecromancer:
		return []Item{
			{ID: "starter_bone_staff", Name: "Bone Staff", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 6, DamageType: DamageNecrotic, Weight: 2, Value: 14},
			{ID: "starter_dark_robes", Name: "Dark Robes", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 1, Weight: 2, Value: 15},
		}
	case ClassBerserker:
		return []Item{
			{ID: "starter_great_axe", Name: "Great Axe", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 2, DamageDice: 6, DamageType: DamagePhysical, Weight: 5, Value: 20},
			{ID: "starter_hide", Name: "Hide Armor", Type: ItemTypeArmor, Slot: SlotArmor, ACBonus: 2, Weight: 5, Value: 15},
		}
	}
	return nil
}

// HealthPotion returns a standard health potion item.
func HealthPotion() Item {
	return Item{
		ID:          fmt.Sprintf("health_potion_%d", Roll(10000)),
		Name:        "Health Potion",
		Description: "Restores 2d4+2 HP",
		Type:        ItemTypeConsumable,
		HealAmount:  0, // Calculated on use: 2d4+2
		Weight:      0.5,
		Value:       25,
	}
}

// AddPKKarma adds karma for killing an innocent player and increments PK count.
func (c *Character) AddPKKarma(amount int) {
	c.Karma += amount
	c.PKCount++
}

// AddPvPKill increments the consensual PvP kill count (no karma change).
func (c *Character) AddPvPKill() {
	c.PvPCount++
}

// ReduceKarma reduces karma (e.g. from mob kills), minimum 0.
func (c *Character) ReduceKarma(amount int) {
	c.Karma -= amount
	if c.Karma < 0 {
		c.Karma = 0
	}
}

// SetFlagged marks the character as purple (attacked someone) for 5 minutes.
func (c *Character) SetFlagged() {
	c.Flagged = true
	c.FlaggedUntil = time.Now().Unix() + 300
}

// CheckFlagExpiry clears the flagged status if the timer has expired.
func (c *Character) CheckFlagExpiry() {
	if c.Flagged && time.Now().Unix() > c.FlaggedUntil {
		c.Flagged = false
	}
}

// IsRed returns true if the character has PK karma.
func (c *Character) IsRed() bool {
	return c.Karma > 0
}

// IsPurple returns true if the character is currently flagged (attacked someone).
func (c *Character) IsPurple() bool {
	c.CheckFlagExpiry()
	return c.Flagged
}

// GetNameColor returns the hex color for a character's name based on karma and flag state.
func GetNameColor(karma int, flagged bool) string {
	if flagged {
		return "#9b59b6" // purple
	}
	if karma > 0 {
		return "#c0392b" // red
	}
	return "#ffffff" // white
}

// GetPKTitle returns a display title based on karma and PK count.
func GetPKTitle(karma int, pkCount int) string {
	switch {
	case karma >= 5000:
		return "Serial Killer"
	case karma >= 1000:
		return "Wanted"
	case karma > 0:
		return "Outlaw"
	default:
		return "Innocent"
	}
}

// RollItemDrop determines which items a player drops on PvP death based on karma.
// Directly removes items from the character's inventory and clears weapon slot.
// Returns the list of dropped items.
func RollItemDrop(c *Character) []Item {
	var dropChance int
	var dropWeapon bool

	switch {
	case c.Karma >= 5000:
		dropChance = 30
		dropWeapon = true
	case c.Karma >= 1000:
		dropChance = 15
	case c.Karma > 0:
		dropChance = 5
	default:
		return nil // no drop for clean players
	}

	var dropped []Item

	// Roll for random inventory item drop
	if len(c.Inventory) > 0 && rand.Intn(100) < dropChance {
		idx := rand.Intn(len(c.Inventory))
		dropped = append(dropped, c.Inventory[idx])
		c.Inventory = append(c.Inventory[:idx], c.Inventory[idx+1:]...)
	}

	// Roll for weapon drop at 5000+ karma
	if dropWeapon && c.Equipment.Weapon != nil && rand.Intn(100) < dropChance {
		dropped = append(dropped, *c.Equipment.Weapon)
		c.Equipment.Weapon = nil
	}

	return dropped
}

// ManaPotion returns a standard mana potion item.
func ManaPotion() Item {
	return Item{
		ID:          fmt.Sprintf("mana_potion_%d", Roll(10000)),
		Name:        "Mana Potion",
		Description: "Restores 2d4+2 Mana",
		Type:        ItemTypeConsumable,
		ManaRestore: 0, // Calculated on use: 2d4+2
		Weight:      0.5,
		Value:       30,
	}
}
