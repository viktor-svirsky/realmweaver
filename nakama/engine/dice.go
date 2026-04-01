package engine

import (
	"math/rand"
	"sync"
	"time"
)

var (
	rng   *rand.Rand
	rngMu sync.Mutex
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Roll returns a random number between 1 and sides (inclusive).
func Roll(sides int) int {
	if sides <= 0 {
		return 0
	}
	rngMu.Lock()
	v := rng.Intn(sides) + 1
	rngMu.Unlock()
	return v
}

// RollD20 rolls a 20-sided die.
func RollD20() int {
	return Roll(20)
}

// RollDice rolls count dice of the given sides and returns the total.
func RollDice(count, sides int) int {
	total := 0
	for i := 0; i < count; i++ {
		total += Roll(sides)
	}
	return total
}

// RollWithModifier rolls a d20 and adds the modifier.
func RollWithModifier(modifier int) (roll int, total int) {
	roll = RollD20()
	return roll, roll + modifier
}

// RollInitiative rolls initiative for combat ordering.
func RollInitiative(dexMod int) int {
	_, total := RollWithModifier(dexMod)
	return total
}

// SeedRng allows setting a deterministic seed for testing.
func SeedRng(seed int64) {
	rngMu.Lock()
	rng = rand.New(rand.NewSource(seed))
	rngMu.Unlock()
}
