package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/storage"
	"realmweaver/world"
)

// TravelRequest is the input for the travel RPC.
type TravelRequest struct {
	CharacterID string `json:"character_id"`
	RegionX     int    `json:"region_x"`
	RegionY     int    `json:"region_y"`
}

// TravelResponse is the output of the travel RPC.
type TravelResponse struct {
	Region     *world.Region `json:"region"`
	TravelTime int           `json:"travel_time"`
	Narrative  string        `json:"narrative"`
	NPCs       []world.NPC   `json:"npcs"`
}

// RPCTravel handles player travel to an adjacent hex region.
func RPCTravel(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req TravelRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	// Load character
	character, err := storage.LoadCharacter(ctx, nk, userID, req.CharacterID)
	if err != nil || character == nil {
		return "", runtime.NewError("character not found", 5)
	}

	// Validate destination is adjacent (within 1 hex)
	if !world.IsAdjacent(character.RegionX, character.RegionY, req.RegionX, req.RegionY) {
		return "", runtime.NewError("destination is not adjacent to current position", 3)
	}

	// Use singleton world DB connection
	wdb := getSocialWorldDB()
	if wdb == nil {
		return "", runtime.NewError("world database not configured", 13)
	}

	// Generate the region if it doesn't exist
	region, err := world.GenerateRegion(ctx, wdb, req.RegionX, req.RegionY)
	if err != nil {
		logger.Error("Failed to generate region (%d,%d): %v", req.RegionX, req.RegionY, err)
		return "", runtime.NewError("failed to generate region", 13)
	}

	// Calculate travel time based on destination biome
	travelTime := world.TravelTimeForBiome(region.Biome)

	// Update character position
	character.RegionX = req.RegionX
	character.RegionY = req.RegionY

	// Save character
	if err := storage.SaveCharacter(ctx, nk, userID, character); err != nil {
		logger.Error("Failed to save character after travel: %v", err)
		return "", runtime.NewError("failed to save character", 13)
	}

	// Update player location in world DB
	wdb.UpdatePlayerLocation(ctx, &world.PlayerLocation{
		UserID:         userID,
		CharacterID:    character.ID,
		CharacterName:  character.Name,
		CharacterClass: string(character.Class),
		CharacterLevel: character.Level,
		RegionX:        character.RegionX,
		RegionY:        character.RegionY,
		Karma:          character.Karma,
	})

	// Load NPCs for the new region
	npcs, _ := wdb.GetNPCsInRegion(ctx, region.ID)

	// Generate travel narrative
	narrative := generateTravelNarrative(region)

	resp := TravelResponse{
		Region:     region,
		TravelTime: travelTime,
		Narrative:  narrative,
		NPCs:       npcs,
	}

	data, _ := json.Marshal(resp)
	return string(data), nil
}

// generateTravelNarrative creates a short narrative for the journey.
func generateTravelNarrative(region *world.Region) string {
	biomeNarratives := map[string]string{
		"plains":    "You travel across open grasslands, the wind at your back. The terrain is easy and the path clear.",
		"farmlands": "You walk along dirt roads between tended fields. The smell of fresh earth fills the air.",
		"forest":    "You push through dense trees, following animal trails deeper into the woods. Shafts of light pierce the canopy.",
		"hills":     "You climb steadily upward over rolling hills. The view from the ridgeline is breathtaking.",
		"foothills": "You navigate rocky paths that wind between larger and larger boulders. The mountains loom closer.",
		"mountains": "The ascent is brutal. Thin air and loose scree make every step a challenge. You finally crest a ridge.",
		"swamp":     "You slog through murky water and tangled roots. Insects buzz constantly and the air reeks of decay.",
		"marsh":     "The ground grows soft and treacherous. Reeds tower overhead as you pick your way through the wetlands.",
		"desert":    "Sand shifts beneath your feet as you cross the barren expanse. The sun beats down mercilessly.",
		"wastes":    "You cross cracked, lifeless earth. The wind carries dust and the faint smell of something long dead.",
		"snow":      "Snow crunches underfoot as you trudge through the frozen landscape. The cold seeps into your bones.",
		"ice":       "You carefully cross sheets of ice, your breath crystallizing in the frigid air. The silence is absolute.",
		"coast":     "You follow the coastline, waves crashing against rocks below. Salt spray stings your face.",
	}

	narrative := biomeNarratives[region.Biome]
	if narrative == "" {
		narrative = "You travel across unfamiliar terrain, watching for danger."
	}

	return fmt.Sprintf("%s\n\nYou arrive at %s.", narrative, region.Name)
}
