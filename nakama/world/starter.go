package world

import (
	"context"
	"encoding/json"
)

// SeedStarterRegion creates the starting town of Millhaven if it doesn't exist.
func SeedStarterRegion(ctx context.Context, wdb *WorldDB) error {
	exists, err := wdb.RegionExists(ctx, 0, 0)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	structures, _ := json.Marshal([]Structure{
		{Name: "The Wanderer's Rest", Type: "tavern", Description: "A warm tavern with a crackling fireplace. The innkeeper Marta serves ale and gossip in equal measure."},
		{Name: "Ironheart Forge", Type: "shop", Description: "The blacksmith's forge glows day and night. Weapons and armor hang from every wall."},
		{Name: "Town Square", Type: "square", Description: "The heart of Millhaven. A weathered stone fountain stands at the center, surrounded by market stalls."},
		{Name: "The Goblin Cave", Type: "dungeon", Description: "A dark cave mouth in the hillside north of town. Strange noises echo from within. The townfolk avoid it."},
		{Name: "Chapel of Dawn", Type: "house", Description: "A small chapel where the town cleric tends to the sick and offers blessings to travelers."},
	})

	region := &Region{
		X:         0,
		Y:         0,
		Biome:     "forest",
		Difficulty: 1,
		Name:      "Millhaven",
		Description: "A small but lively town nestled at the edge of an ancient forest. Cobblestone paths wind between timber-framed buildings. The air smells of pine and fresh bread. To the north, dark hills hide the entrance to a cave the locals speak of in whispers.",
		Lore:      "Millhaven was founded two centuries ago by settlers who discovered a natural spring in the forest. The town prospered as a trading post, but recent goblin raids from the northern caves have threatened its peace. The town guard is stretched thin, and the mayor has posted bounties for adventurers willing to clear the caves.",
		Structures: structures,
	}

	if err := wdb.InsertRegion(ctx, region); err != nil {
		return err
	}

	// Seed NPCs
	npcs := []NPC{
		{
			RegionID:        region.ID,
			Name:            "Marta",
			Race:            "human",
			Occupation:      "innkeeper",
			PersonalityPrompt: "Warm and motherly, but sharp-tongued. She knows everyone's business and loves sharing rumors. She worries about the goblin raids affecting trade. Speaks with a slight accent.",
			Alive:           true,
			LocationDetail:  "The Wanderer's Rest",
		},
		{
			RegionID:        region.ID,
			Name:            "Theron",
			Race:            "dwarf",
			Occupation:      "blacksmith",
			PersonalityPrompt: "Gruff but fair. Proud of his craft. He'll haggle but respects a good fighter. His shipment of rare ore went missing — he suspects the goblins took it.",
			Alive:           true,
			LocationDetail:  "Ironheart Forge",
		},
		{
			RegionID:        region.ID,
			Name:            "Elder Corin",
			Race:            "human",
			Occupation:      "mayor",
			PersonalityPrompt: "An elderly man with a calm demeanor but growing desperation. He offers the quest to clear the Goblin Cave. Formal speech, treats adventurers with respect.",
			Alive:           true,
			LocationDetail:  "Town Square",
		},
		{
			RegionID:        region.ID,
			Name:            "Sister Lina",
			Race:            "human",
			Occupation:      "cleric",
			PersonalityPrompt: "Gentle and kind. She can heal minor wounds and offers blessings. She senses a dark presence growing in the caves and urges caution.",
			Alive:           true,
			LocationDetail:  "Chapel of Dawn",
		},
		{
			RegionID:        region.ID,
			Name:            "Pip",
			Race:            "halfling",
			Occupation:      "merchant",
			PersonalityPrompt: "A fast-talking halfling merchant with a cart of curiosities. Sells potions, trinkets, and occasionally useful items. Always looking for a deal.",
			Alive:           true,
			LocationDetail:  "Town Square",
		},
	}

	for i := range npcs {
		if err := wdb.InsertNPC(ctx, &npcs[i]); err != nil {
			return err
		}
	}

	return nil
}
