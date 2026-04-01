package world

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// Biome constants for region generation.
const (
	BiomePlains    = "plains"
	BiomeFarmlands = "farmlands"
	BiomeForest    = "forest"
	BiomeHills     = "hills"
	BiomeFoothills = "foothills"
	BiomeMountains = "mountains"
	BiomeSwamp     = "swamp"
	BiomeDesert    = "desert"
	BiomeWastes    = "wastes"
	BiomeSnow      = "snow"
	BiomeIce       = "ice"
	BiomeCoast     = "coast"
	BiomeMarsh     = "marsh"
)

// TravelTimeForBiome returns the travel time in seconds for a given biome.
func TravelTimeForBiome(biome string) int {
	switch biome {
	case BiomePlains, BiomeFarmlands:
		return 3
	case BiomeForest:
		return 5
	case BiomeHills, BiomeFoothills:
		return 7
	case BiomeMountains:
		return 10
	case BiomeSwamp, BiomeMarsh:
		return 8
	case BiomeDesert, BiomeWastes:
		return 8
	case BiomeSnow, BiomeIce:
		return 9
	case BiomeCoast:
		return 5
	default:
		return 6 // unknown/fog
	}
}

// GenerateRegion procedurally creates a region at the given coordinates.
// Determines biome from position, generates name, NPCs, and structures.
// Saves to PostgreSQL and returns the region.
func GenerateRegion(ctx context.Context, wdb *WorldDB, x, y int) (*Region, error) {
	// Check if already exists
	existing, err := wdb.GetRegion(ctx, x, y)
	if err != nil {
		return nil, fmt.Errorf("check existing region: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	biome := determineBiome(x, y)
	difficulty := hexDistance(x, y)
	if difficulty < 1 {
		difficulty = 1
	}

	name := generateRegionName(biome)
	description := generateRegionDescription(biome, name, difficulty)
	structures := generateStructures(biome, difficulty)
	structJSON, _ := json.Marshal(structures)

	region := &Region{
		X:           x,
		Y:           y,
		Biome:       biome,
		Difficulty:  difficulty,
		Name:        name,
		Description: description,
		Lore:        fmt.Sprintf("A %s region at the edge of the known world.", biome),
		Structures:  structJSON,
	}

	if err := wdb.InsertRegion(ctx, region); err != nil {
		return nil, fmt.Errorf("insert region: %w", err)
	}
	if region.ID == 0 {
		return nil, fmt.Errorf("region insert failed: ID is 0")
	}

	// Generate NPCs
	npcs := generateNPCs(biome, difficulty, region.ID)
	for i := range npcs {
		if err := wdb.InsertNPC(ctx, &npcs[i]); err != nil {
			// Non-fatal: region exists even if NPC insert fails
			continue
		}
	}

	return region, nil
}

// determineBiome picks a biome based on hex coordinates using direction and distance.
// North: mountains -> ice, East: forest -> desert, South: plains -> swamp, West: hills -> coast.
func determineBiome(x, y int) string {
	dist := hexDistance(x, y)
	if dist == 0 {
		return BiomeForest
	}

	// Use angle from center to determine direction
	angle := math.Atan2(float64(y), float64(x))
	degrees := angle * 180 / math.Pi

	// Add randomness based on coordinates
	jitter := float64((x*7+y*13)%20-10) * 0.5

	switch {
	// North: mountains -> ice (roughly -90 degrees in hex coords, but y is inverted for hex)
	case degrees >= -135 && degrees < -45:
		if dist+int(jitter) >= 3 {
			return BiomeIce
		}
		if dist >= 2 {
			return BiomeMountains
		}
		return BiomeHills

	// East: forest -> desert
	case degrees >= -45 && degrees < 45:
		if dist+int(jitter) >= 3 {
			return BiomeDesert
		}
		if dist >= 2 {
			return BiomeWastes
		}
		return BiomeForest

	// South: plains -> swamp
	case degrees >= 45 && degrees < 135:
		if dist+int(jitter) >= 3 {
			return BiomeSwamp
		}
		if dist >= 2 {
			return BiomeMarsh
		}
		return BiomePlains

	// West: hills -> coast
	default:
		if dist+int(jitter) >= 3 {
			return BiomeCoast
		}
		if dist >= 2 {
			return BiomeFoothills
		}
		return BiomeFarmlands
	}
}

// hexDistance returns the hex distance from center (ring number) using axial coordinates.
func hexDistance(x, y int) int {
	s := -x - y
	ax, ay, as := x, y, s
	if ax < 0 {
		ax = -ax
	}
	if ay < 0 {
		ay = -ay
	}
	if as < 0 {
		as = -as
	}
	return maxOf3(ax, ay, as)
}

func maxOf3(a, b, c int) int {
	if a >= b && a >= c {
		return a
	}
	if b >= c {
		return b
	}
	return c
}

// IsAdjacent returns true if (x1,y1) and (x2,y2) are within 1 hex of each other.
func IsAdjacent(x1, y1, x2, y2 int) bool {
	dx := x2 - x1
	dy := y2 - y1
	ds := (-x2 - y2) - (-x1 - y1)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if ds < 0 {
		ds = -ds
	}
	return maxOf3(dx, dy, ds) == 1
}

var biomeAdjectives = map[string][]string{
	BiomePlains:    {"Golden", "Windswept", "Endless", "Quiet", "Verdant"},
	BiomeFarmlands: {"Fertile", "Peaceful", "Sun-drenched", "Rolling", "Green"},
	BiomeForest:    {"Dark", "Ancient", "Whispering", "Shadowed", "Emerald"},
	BiomeHills:     {"Rolling", "Craggy", "Wind-blown", "Stony", "Rugged"},
	BiomeFoothills: {"Rocky", "Mist-covered", "Steep", "Weathered", "Wild"},
	BiomeMountains: {"Towering", "Frozen", "Jagged", "Storm-wracked", "Perilous"},
	BiomeSwamp:     {"Murky", "Festering", "Fog-bound", "Cursed", "Rotting"},
	BiomeMarsh:     {"Boggy", "Misty", "Treacherous", "Damp", "Gloomy"},
	BiomeDesert:    {"Scorching", "Barren", "Sun-bleached", "Shifting", "Desolate"},
	BiomeWastes:    {"Blighted", "Ashen", "Twisted", "Forsaken", "Cracked"},
	BiomeSnow:      {"Frozen", "Silent", "Bitter", "Crystal", "White"},
	BiomeIce:       {"Glacial", "Eternal", "Shimmering", "Deadly", "Frigid"},
	BiomeCoast:     {"Windswept", "Salt-crusted", "Misty", "Rocky", "Storm-battered"},
}

var biomeNouns = map[string][]string{
	BiomePlains:    {"Meadows", "Expanse", "Fields", "Steppes", "Grasslands"},
	BiomeFarmlands: {"Farmstead", "Pastures", "Homestead", "Acres", "Dell"},
	BiomeForest:    {"Woods", "Thicket", "Grove", "Timberland", "Glade"},
	BiomeHills:     {"Heights", "Ridges", "Knolls", "Bluffs", "Downs"},
	BiomeFoothills: {"Pass", "Slopes", "Terrace", "Ledge", "Outcrops"},
	BiomeMountains: {"Peaks", "Summit", "Crags", "Spire", "Pinnacle"},
	BiomeSwamp:     {"Mire", "Bog", "Fen", "Quagmire", "Hollow"},
	BiomeMarsh:     {"Wetlands", "Lowlands", "Fen", "Morass", "Bottoms"},
	BiomeDesert:    {"Sands", "Dunes", "Wasteland", "Flats", "Barrens"},
	BiomeWastes:    {"Ruins", "Desolation", "Badlands", "Reach", "Expanse"},
	BiomeSnow:      {"Tundra", "Drift", "Expanse", "Wastes", "Reach"},
	BiomeIce:       {"Glacier", "Shelf", "Floe", "Waste", "Reach"},
	BiomeCoast:     {"Shore", "Cliffs", "Cove", "Strand", "Headland"},
}

func generateRegionName(biome string) string {
	adjs := biomeAdjectives[biome]
	nouns := biomeNouns[biome]
	if adjs == nil {
		adjs = []string{"Unknown"}
	}
	if nouns == nil {
		nouns = []string{"Lands"}
	}
	return adjs[rand.Intn(len(adjs))] + " " + nouns[rand.Intn(len(nouns))]
}

func generateRegionDescription(biome, name string, difficulty int) string {
	templates := map[string][]string{
		BiomePlains:    {"%s stretches before you, tall grass swaying in the wind. The sky is vast and open.", "The %s are quiet save for the rustle of grass and distant birdsong."},
		BiomeFarmlands: {"Rows of crops line the %s, tended by unseen hands. A dirt road winds between fences.", "The %s smell of fresh earth and growing things. Smoke rises from a distant chimney."},
		BiomeForest:    {"Massive trees form a canopy over the %s, filtering the light into green shadows.", "The %s are dense and ancient. Strange sounds echo between the trunks."},
		BiomeHills:     {"The %s roll into the distance, dotted with boulders and scrub brush.", "Wind howls across the %s, carrying the scent of stone and wild herbs."},
		BiomeFoothills: {"The %s rise sharply ahead, a transition between flatlands and mountains.", "Rocky paths wind through the %s, offering glimpses of the peaks above."},
		BiomeMountains: {"The %s loom overhead, their peaks lost in clouds. The air is thin and cold.", "Sheer cliffs and narrow ledges define the treacherous %s."},
		BiomeSwamp:     {"Murky water pools between twisted trees in the %s. The air is thick with insects.", "The %s reek of decay. Every step risks sinking into the muck."},
		BiomeMarsh:     {"Reeds and cattails fill the %s. The ground squelches underfoot.", "Mist hangs low over the %s, obscuring what lurks in the shallow water."},
		BiomeDesert:    {"Sand stretches to the horizon in the %s. Heat shimmers rise from the ground.", "The %s are merciless. No shade, no water, only endless sand."},
		BiomeWastes:    {"The %s are barren and lifeless. Cracked earth stretches in every direction.", "A desolate wind blows through the %s, carrying dust and echoes of the past."},
		BiomeSnow:      {"Snow blankets the %s in every direction. The cold bites through armor.", "The %s are silent and white. Your breath hangs in frozen clouds."},
		BiomeIce:       {"Sheets of ice cover the %s, reflecting pale light. The cold is absolute.", "The %s creak and groan underfoot. One wrong step could be your last."},
		BiomeCoast:     {"Waves crash against the rocky %s. Salt spray fills the air.", "The %s overlook a churning sea. Gulls wheel overhead, crying."},
	}

	t := templates[biome]
	if t == nil {
		return fmt.Sprintf("You arrive at %s. The terrain is unfamiliar and challenging.", name)
	}

	desc := fmt.Sprintf(t[rand.Intn(len(t))], name)
	if difficulty >= 3 {
		desc += " Danger lurks here — only the strong survive."
	}
	return desc
}

func generateStructures(biome string, difficulty int) []Structure {
	var structures []Structure

	// Most regions get 1-2 structures
	switch biome {
	case BiomePlains, BiomeFarmlands:
		structures = append(structures, Structure{Name: "Roadside Camp", Type: "camp", Description: "A well-used campsite with a fire pit."})
		if rand.Intn(2) == 0 {
			structures = append(structures, Structure{Name: "Abandoned Farmhouse", Type: "house", Description: "A crumbling farmhouse with a intact cellar."})
		}
	case BiomeForest:
		structures = append(structures, Structure{Name: "Hunter's Lodge", Type: "camp", Description: "A small lodge used by local hunters."})
		if difficulty >= 2 {
			structures = append(structures, Structure{Name: "Ruined Shrine", Type: "dungeon", Description: "An overgrown shrine to a forgotten god. Dark passages lead underground."})
		}
	case BiomeHills, BiomeFoothills:
		structures = append(structures, Structure{Name: "Mountain Outpost", Type: "camp", Description: "A stone lookout post with a good view of the surrounding area."})
	case BiomeMountains:
		structures = append(structures, Structure{Name: "Cave Entrance", Type: "dungeon", Description: "A dark cave mouth yawns in the mountainside."})
		if rand.Intn(2) == 0 {
			structures = append(structures, Structure{Name: "Dwarven Ruins", Type: "dungeon", Description: "Ancient dwarven carvings frame a collapsed tunnel entrance."})
		}
	case BiomeSwamp, BiomeMarsh:
		structures = append(structures, Structure{Name: "Witch's Hut", Type: "house", Description: "A rickety hut on stilts above the murky water."})
	case BiomeDesert, BiomeWastes:
		structures = append(structures, Structure{Name: "Buried Temple", Type: "dungeon", Description: "Sand-scoured pillars mark the entrance to a buried temple."})
	case BiomeSnow, BiomeIce:
		structures = append(structures, Structure{Name: "Frozen Watchtower", Type: "camp", Description: "An ice-encrusted tower, abandoned long ago."})
	case BiomeCoast:
		structures = append(structures, Structure{Name: "Fisherman's Shack", Type: "house", Description: "A weathered shack smelling of salt and fish."})
		if rand.Intn(2) == 0 {
			structures = append(structures, Structure{Name: "Shipwreck", Type: "dungeon", Description: "The broken hull of a ship lies half-buried in the sand."})
		}
	}

	// Higher difficulty regions may have a tavern or shop
	if difficulty >= 2 && rand.Intn(3) == 0 {
		structures = append(structures, Structure{Name: "Traveling Merchant's Tent", Type: "shop", Description: "A colorful tent where a merchant hawks wares."})
	}

	return structures
}

var biomeNPCNames = map[string][]struct {
	Name       string
	Race       string
	Occupation string
	Personality string
	Location   string
}{
	BiomePlains: {
		{"Farmer Giles", "human", "farmer", "Simple and friendly. Offers directions and warns of bandits on the road.", "Roadside Camp"},
		{"Wandering Bard", "half-elf", "bard", "Cheerful and talkative. Knows many tales and songs of distant lands.", "The road"},
	},
	BiomeForest: {
		{"Ranger Elara", "elf", "ranger", "Quiet and watchful. Knows every path in the forest. Suspicious of strangers.", "Hunter's Lodge"},
		{"Old Hemlock", "human", "hermit", "Eccentric and cryptic. Speaks in riddles. Sells herbs and mushrooms.", "Deep in the woods"},
	},
	BiomeHills: {
		{"Prospector Dunn", "dwarf", "miner", "Grizzled and tough. Looking for ore veins. Will trade supplies for help.", "Mountain Outpost"},
	},
	BiomeMountains: {
		{"Goat Herder Sven", "human", "herder", "Stoic and laconic. Knows secret paths through the mountains.", "Mountain Pass"},
		{"Storm Priestess", "human", "cleric", "Mysterious and powerful. Offers blessings for a price.", "Cave Entrance"},
	},
	BiomeSwamp: {
		{"Bog Witch Morrigan", "human", "witch", "Unsettling but helpful. Brews potions from swamp ingredients.", "Witch's Hut"},
	},
	BiomeDesert: {
		{"Nomad Kadir", "human", "trader", "Shrewd and hospitable. Trades water and supplies at steep prices.", "Buried Temple"},
	},
	BiomeSnow: {
		{"Frost Hunter Bjorn", "human", "hunter", "Hardy and practical. Hunts frost wolves for their pelts.", "Frozen Watchtower"},
	},
	BiomeCoast: {
		{"Fisher Mae", "human", "fisher", "Tough and salty. Knows the tides and the dangers of the sea.", "Fisherman's Shack"},
	},
}

func generateNPCs(biome string, difficulty int, regionID int) []NPC {
	templates := biomeNPCNames[biome]
	if templates == nil {
		// Fallback: generic NPC
		return []NPC{
			{
				RegionID:          regionID,
				Name:              "Mysterious Traveler",
				Race:              "human",
				Occupation:        "wanderer",
				PersonalityPrompt: "Enigmatic and quiet. Seems to know more than they let on.",
				Alive:             true,
				LocationDetail:    "The area",
			},
		}
	}

	// Pick 2-3 NPCs (or all if fewer available)
	count := 2 + rand.Intn(2)
	if count > len(templates) {
		count = len(templates)
	}

	perm := rand.Perm(len(templates))
	var npcs []NPC
	for i := 0; i < count; i++ {
		t := templates[perm[i]]
		npcs = append(npcs, NPC{
			RegionID:          regionID,
			Name:              t.Name,
			Race:              t.Race,
			Occupation:        t.Occupation,
			PersonalityPrompt: t.Personality,
			Alive:             true,
			LocationDetail:    t.Location,
		})
	}

	return npcs
}
