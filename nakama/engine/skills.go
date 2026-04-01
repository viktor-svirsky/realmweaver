package engine

// SkillCheckResult holds the outcome of a skill check.
type SkillCheckResult struct {
	Stat     string `json:"stat"`
	Roll     int    `json:"roll"`
	Modifier int    `json:"modifier"`
	DC       int    `json:"dc"`
	Total    int    `json:"total"`
	Success  bool   `json:"success"`
}

// SkillCheck performs a d20 + stat modifier check against a difficulty class.
func SkillCheck(character *Character, stat string, dc int) SkillCheckResult {
	mod := getStatModifier(character, stat)
	roll := RollD20()
	total := roll + mod
	return SkillCheckResult{
		Stat:     stat,
		Roll:     roll,
		Modifier: mod,
		DC:       dc,
		Total:    total,
		Success:  total >= dc,
	}
}

func getStatModifier(c *Character, stat string) int {
	switch stat {
	case "str", "STR":
		return Modifier(c.Stats.STR)
	case "dex", "DEX":
		return Modifier(c.Stats.DEX)
	case "con", "CON":
		return Modifier(c.Stats.CON)
	case "int", "INT":
		return Modifier(c.Stats.INT)
	case "wis", "WIS":
		return Modifier(c.Stats.WIS)
	case "cha", "CHA":
		return Modifier(c.Stats.CHA)
	}
	return 0
}

// GetStatValue returns the raw stat value for a given stat name.
func GetStatValue(c *Character, stat string) int {
	switch stat {
	case "str", "STR":
		return c.Stats.STR
	case "dex", "DEX":
		return c.Stats.DEX
	case "con", "CON":
		return c.Stats.CON
	case "int", "INT":
		return c.Stats.INT
	case "wis", "WIS":
		return c.Stats.WIS
	case "cha", "CHA":
		return c.Stats.CHA
	}
	return 10
}
