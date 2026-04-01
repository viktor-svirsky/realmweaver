package engine

import (
	"fmt"
	"strings"
)

const MaxInventorySize = 20
const MaxWeight = 50.0

// AddItem adds an item to the character's inventory.
func AddItem(c *Character, item Item) error {
	if len(c.Inventory) >= MaxInventorySize {
		return fmt.Errorf("inventory full (max %d items)", MaxInventorySize)
	}
	if CurrentWeight(c)+item.Weight > MaxWeight {
		return fmt.Errorf("too heavy (%.1f/%.1f)", CurrentWeight(c)+item.Weight, MaxWeight)
	}
	c.Inventory = append(c.Inventory, item)
	return nil
}

// RemoveItem removes an item by ID from inventory. Returns the removed item.
func RemoveItem(c *Character, itemID string) (*Item, error) {
	for i, item := range c.Inventory {
		if item.ID == itemID {
			removed := c.Inventory[i]
			c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			return &removed, nil
		}
	}
	return nil, fmt.Errorf("item %q not found in inventory", itemID)
}

// FindItem finds an item by ID in inventory.
func FindItem(c *Character, itemID string) *Item {
	for i := range c.Inventory {
		if c.Inventory[i].ID == itemID {
			return &c.Inventory[i]
		}
	}
	return nil
}

// EquipItem moves an item from inventory to an equipment slot.
// If the slot is occupied, the old item goes back to inventory.
func EquipItem(c *Character, itemID string) error {
	item, err := RemoveItem(c, itemID)
	if err != nil {
		return err
	}
	if item.Slot == "" {
		// Put it back, can't equip
		c.Inventory = append(c.Inventory, *item)
		return fmt.Errorf("%q cannot be equipped", item.Name)
	}

	prev := c.Equipment.SetSlot(item.Slot, item)
	if prev != nil {
		c.Inventory = append(c.Inventory, *prev)
	}
	c.RecalcDerived()
	return nil
}

// UnequipItem removes an item from an equipment slot and puts it in inventory.
func UnequipItem(c *Character, slot EquipSlot) error {
	item := c.Equipment.GetSlot(slot)
	if item == nil {
		return fmt.Errorf("nothing equipped in %s slot", slot)
	}
	if len(c.Inventory) >= MaxInventorySize {
		return fmt.Errorf("inventory full, cannot unequip")
	}
	c.Equipment.SetSlot(slot, nil)
	c.Inventory = append(c.Inventory, *item)
	c.RecalcDerived()
	return nil
}

// UseItem uses a consumable item (potions, etc). Returns a description of what happened.
func UseItem(c *Character, itemID string) (string, error) {
	item := FindItem(c, itemID)
	if item == nil {
		return "", fmt.Errorf("item %q not found", itemID)
	}
	if item.Type != ItemTypeConsumable {
		return "", fmt.Errorf("%q is not consumable", item.Name)
	}

	var result string

	// Match by HealAmount/ManaRestore fields — not by name (names change with affixes)
	switch {
	case item.HealAmount > 0:
		healed := c.Heal(item.HealAmount)
		result = fmt.Sprintf("Used %s, restored %d HP (HP: %d/%d)", item.Name, healed, c.HP, c.MaxHP)
	case item.ManaRestore > 0:
		c.Mana += item.ManaRestore
		if c.Mana > c.MaxMana {
			c.Mana = c.MaxMana
		}
		result = fmt.Sprintf("Used %s, restored %d Mana (Mana: %d/%d)", item.Name, item.ManaRestore, c.Mana, c.MaxMana)
	default:
		// Generic consumable with no heal/mana — roll 2d4+2 as HP restore (potions)
		if strings.Contains(strings.ToLower(item.Name), "health") {
			amount := RollDice(2, 4) + 2
			healed := c.Heal(amount)
			result = fmt.Sprintf("Drank %s, restored %d HP (HP: %d/%d)", item.Name, healed, c.HP, c.MaxHP)
		} else if strings.Contains(strings.ToLower(item.Name), "mana") {
			amount := RollDice(2, 4) + 2
			c.Mana += amount
			if c.Mana > c.MaxMana {
				c.Mana = c.MaxMana
			}
			result = fmt.Sprintf("Drank %s, restored %d Mana (Mana: %d/%d)", item.Name, amount, c.Mana, c.MaxMana)
		} else {
			result = fmt.Sprintf("Used %s", item.Name)
		}
	}

	// Remove consumed item
	RemoveItem(c, itemID)
	return result, nil
}

// CurrentWeight returns total weight of inventory + equipment.
func CurrentWeight(c *Character) float64 {
	total := 0.0
	for _, item := range c.Inventory {
		total += item.Weight
	}
	// Add equipped items
	equipped := []*Item{c.Equipment.Weapon, c.Equipment.Offhand, c.Equipment.Armor,
		c.Equipment.Helmet, c.Equipment.Boots, c.Equipment.Ring1, c.Equipment.Ring2, c.Equipment.Amulet}
	for _, item := range equipped {
		if item != nil {
			total += item.Weight
		}
	}
	return total
}

// DropItem removes an item from inventory permanently.
func DropItem(c *Character, itemID string) (*Item, error) {
	return RemoveItem(c, itemID)
}
