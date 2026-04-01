package world

import "encoding/json"

// Structure represents a notable location within a region.
type Structure struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "tavern", "shop", "dungeon", "house", "square"
	Description string `json:"description"`
}

// RegionSummary is a concise representation for Claude context.
type RegionSummary struct {
	Name        string      `json:"name"`
	Biome       string      `json:"biome"`
	Description string      `json:"description"`
	Structures  []Structure `json:"structures"`
	NPCNames    []string    `json:"npc_names"`
}

// ParseStructures extracts structures from a region's JSON.
func ParseStructures(raw json.RawMessage) []Structure {
	var structures []Structure
	if raw == nil {
		return structures
	}
	json.Unmarshal(raw, &structures)
	return structures
}

// BuildRegionSummary creates a summary for Claude context.
func BuildRegionSummary(region *Region, npcs []NPC) RegionSummary {
	structures := ParseStructures(region.Structures)
	npcNames := make([]string, len(npcs))
	for i, n := range npcs {
		npcNames[i] = n.Name + " (" + n.Occupation + ")"
	}
	return RegionSummary{
		Name:        region.Name,
		Biome:       region.Biome,
		Description: region.Description,
		Structures:  structures,
		NPCNames:    npcNames,
	}
}
