package engine

import (
	"fmt"
	"sort"
)

// InitCombat sets up a combat encounter with initiative order.
func InitCombat(character *Character, enemies []Enemy) *CombatState {
	type initiative struct {
		ID   string
		Roll int
	}

	rolls := []initiative{
		{ID: character.ID, Roll: RollInitiative(Modifier(character.Stats.DEX))},
	}
	for _, e := range enemies {
		rolls = append(rolls, initiative{ID: e.ID, Roll: RollInitiative(Modifier(e.DEX))})
	}

	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].Roll > rolls[j].Roll
	})

	order := make([]string, len(rolls))
	for i, r := range rolls {
		order[i] = r.ID
	}

	return &CombatState{
		Enemies:     enemies,
		TurnOrder:   order,
		CurrentTurn: 0,
		Round:       1,
	}
}

// CurrentTurnID returns whose turn it is.
func (cs *CombatState) CurrentTurnID() string {
	if len(cs.TurnOrder) == 0 {
		return ""
	}
	return cs.TurnOrder[cs.CurrentTurn%len(cs.TurnOrder)]
}

// AdvanceTurn moves to the next turn, advancing round if needed.
func (cs *CombatState) AdvanceTurn() {
	cs.CurrentTurn++
	if cs.CurrentTurn >= len(cs.TurnOrder) {
		cs.CurrentTurn = 0
		cs.Round++
	}
	cs.PlayerDefending = false
}

// FindEnemy returns a pointer to an enemy by ID.
func (cs *CombatState) FindEnemy(id string) *Enemy {
	for i := range cs.Enemies {
		if cs.Enemies[i].ID == id {
			return &cs.Enemies[i]
		}
	}
	return nil
}

// AliveEnemies returns enemies with HP > 0.
func (cs *CombatState) AliveEnemies() []Enemy {
	alive := []Enemy{}
	for _, e := range cs.Enemies {
		if e.HP > 0 {
			alive = append(alive, e)
		}
	}
	return alive
}

// IsOver returns true if all enemies are dead or combat was ended.
func (cs *CombatState) IsOver() bool {
	return len(cs.AliveEnemies()) == 0
}

// ResolveMeleeAttack resolves a melee attack from character against enemy.
func ResolveMeleeAttack(character *Character, enemy *Enemy) ActionResult {
	weapon := character.Equipment.Weapon
	strMod := Modifier(character.Stats.STR)

	roll := RollD20()
	attackTotal := roll + strMod

	result := ActionResult{
		Action:   "melee_attack",
		Actor:    character.Name,
		Target:   enemy.Name,
		Roll:     roll,
		Modifier: strMod,
		DC:       enemy.AC,
	}

	if attackTotal >= enemy.AC {
		// Hit
		var damage int
		if weapon != nil {
			damage = RollDice(weapon.DamageCount, weapon.DamageDice) + strMod
		} else {
			damage = 1 + strMod // unarmed
		}
		if damage < 1 {
			damage = 1
		}

		enemy.HP -= damage
		result.Hit = true
		result.Damage = damage
		result.DamageType = DamagePhysical
		if weapon != nil {
			result.DamageType = weapon.DamageType
		}
		result.Success = true
		result.HPRemaining = enemy.HP

		if enemy.HP <= 0 {
			enemy.HP = 0
			result.EnemyDefeated = true
			result.HPRemaining = 0
		}

		result.Details = fmt.Sprintf("%s attacks %s with %s: %d damage",
			character.Name, enemy.Name, weaponName(weapon), damage)
	} else {
		result.Hit = false
		result.Success = false
		result.Details = fmt.Sprintf("%s swings at %s but misses", character.Name, enemy.Name)
	}

	return result
}

// ResolveEnemyAttack resolves an enemy's attack on the player character.
func ResolveEnemyAttack(enemy *Enemy, character *Character, defending bool) ActionResult {
	strMod := Modifier(enemy.STR)
	roll := RollD20()
	attackTotal := roll + strMod

	targetAC := character.AC
	if defending {
		targetAC += 2 // defend action bonus
	}

	result := ActionResult{
		Action:   "enemy_attack",
		Actor:    enemy.Name,
		Target:   character.Name,
		Roll:     roll,
		Modifier: strMod,
		DC:       targetAC,
	}

	if attackTotal >= targetAC {
		// Parse enemy damage (simplified: use STR mod + d6 baseline)
		damage := Roll(6) + strMod
		if damage < 1 {
			damage = 1
		}
		character.TakeDamage(damage)

		result.Hit = true
		result.Damage = damage
		result.DamageType = DamagePhysical
		result.Success = true
		result.HPRemaining = character.HP
		result.Details = fmt.Sprintf("%s attacks %s for %d damage", enemy.Name, character.Name, damage)
	} else {
		result.Hit = false
		result.Success = false
		result.Details = fmt.Sprintf("%s attacks %s but misses", enemy.Name, character.Name)
	}

	return result
}

// ResolveDefend sets the defending flag for AC bonus on the next enemy turn.
func ResolveDefend(character *Character, combat *CombatState) ActionResult {
	combat.PlayerDefending = true
	return ActionResult{
		Action:  "defend",
		Actor:   character.Name,
		Success: true,
		Details: fmt.Sprintf("%s takes a defensive stance (+2 AC until next turn)", character.Name),
	}
}

// ResolveFlee attempts to flee combat. DEX check DC 12.
func ResolveFlee(character *Character) ActionResult {
	check := SkillCheck(character, "DEX", 12)
	result := ActionResult{
		Action:   "flee",
		Actor:    character.Name,
		Roll:     check.Roll,
		Modifier: check.Modifier,
		DC:       check.DC,
		Success:  check.Success,
	}
	if check.Success {
		result.Details = fmt.Sprintf("%s successfully flees from combat!", character.Name)
		result.CombatEnded = true
	} else {
		result.Details = fmt.Sprintf("%s tries to flee but is blocked!", character.Name)
	}
	return result
}

// CalculateXP returns total XP for defeating all enemies.
func CalculateXP(enemies []Enemy) int {
	total := 0
	for _, e := range enemies {
		if e.HP <= 0 {
			total += e.XP
		}
	}
	return total
}

// GenerateLoot creates random loot from defeated enemies.
func GenerateLoot(enemies []Enemy) []Item {
	var loot []Item
	for _, e := range enemies {
		if e.HP > 0 {
			continue
		}
		// 40% chance to drop gold pouch
		if Roll(100) <= 40 {
			loot = append(loot, Item{
				ID:    fmt.Sprintf("gold_pouch_%d", Roll(10000)),
				Name:  "Gold Pouch",
				Type:  ItemTypeMisc,
				Value: Roll(10) + 5,
				Weight: 0.1,
			})
		}
		// 20% chance to drop health potion
		if Roll(100) <= 20 {
			loot = append(loot, HealthPotion())
		}
		// 10% chance to drop a weapon
		if Roll(100) <= 10 {
			loot = append(loot, randomWeaponDrop())
		}
	}
	return loot
}

func weaponName(weapon *Item) string {
	if weapon == nil {
		return "fists"
	}
	return weapon.Name
}

func randomWeaponDrop() Item {
	weapons := []Item{
		{ID: fmt.Sprintf("drop_sword_%d", Roll(10000)), Name: "Rusty Sword", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 6, DamageType: DamagePhysical, Weight: 3, Value: 8},
		{ID: fmt.Sprintf("drop_axe_%d", Roll(10000)), Name: "Hand Axe", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 6, DamageType: DamagePhysical, Weight: 2, Value: 7},
		{ID: fmt.Sprintf("drop_dagger_%d", Roll(10000)), Name: "Sharp Dagger", Type: ItemTypeWeapon, Slot: SlotWeapon, DamageCount: 1, DamageDice: 4, DamageType: DamagePhysical, Weight: 1, Value: 5},
	}
	return weapons[Roll(len(weapons))-1]
}

// Starter enemies for the goblin cave.
func GoblinCaveEnemies(room int) []Enemy {
	switch room {
	case 1:
		return []Enemy{
			{ID: "goblin_1", Name: "Goblin Scout", HP: 8, MaxHP: 8, AC: 12, STR: 8, DEX: 14, Damage: "1d4+1", XP: 25},
		}
	case 2:
		return []Enemy{
			{ID: "goblin_2", Name: "Goblin Warrior", HP: 15, MaxHP: 15, AC: 13, STR: 12, DEX: 12, Damage: "1d6+2", XP: 50},
			{ID: "goblin_3", Name: "Goblin Archer", HP: 10, MaxHP: 10, AC: 11, STR: 8, DEX: 14, Damage: "1d6+1", XP: 35},
		}
	case 3:
		return []Enemy{
			{ID: "goblin_boss", Name: "Goblin Chieftain", HP: 30, MaxHP: 30, AC: 15, STR: 14, DEX: 12, Damage: "1d8+3", XP: 100},
			{ID: "goblin_shaman", Name: "Goblin Shaman", HP: 12, MaxHP: 12, AC: 11, STR: 8, DEX: 10, Damage: "1d6+1", XP: 60},
		}
	default:
		return []Enemy{
			{ID: "goblin_random", Name: "Goblin", HP: 8, MaxHP: 8, AC: 12, STR: 10, DEX: 12, Damage: "1d4+1", XP: 25},
		}
	}
}
