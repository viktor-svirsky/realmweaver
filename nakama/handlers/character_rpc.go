package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/engine"
	"realmweaver/storage"
)

// CreateCharacterRequest is the input for character creation.
type CreateCharacterRequest struct {
	Name  string       `json:"name"`
	Class engine.Class `json:"class"`
}

// CreateCharacterResponse is the output for character creation.
type CreateCharacterResponse struct {
	Character *engine.Character `json:"character"`
}

// RPCCreateCharacter handles the create_character RPC.
func RPCCreateCharacter(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16) // UNAUTHENTICATED
	}

	var req CreateCharacterRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request: "+err.Error(), 3) // INVALID_ARGUMENT
	}

	if req.Name == "" {
		return "", runtime.NewError("name is required", 3)
	}
	validClasses := map[engine.Class]bool{
		engine.ClassWarrior: true, engine.ClassMage: true, engine.ClassRogue: true, engine.ClassCleric: true,
		engine.ClassRanger: true, engine.ClassPaladin: true, engine.ClassNecromancer: true, engine.ClassBerserker: true,
	}
	if !validClasses[req.Class] {
		return "", runtime.NewError("invalid class", 3)
	}

	// Generate a unique character ID
	uidPrefix := userID
	if len(uidPrefix) > 8 {
		uidPrefix = uidPrefix[:8]
	}
	charID := fmt.Sprintf("char_%s_%s_%d", uidPrefix, req.Name, engine.Roll(99999))

	character := engine.NewCharacter(charID, req.Name, req.Class)

	if err := storage.SaveCharacter(ctx, nk, userID, character); err != nil {
		logger.Error("Failed to save character: %v", err)
		return "", runtime.NewError("failed to save character", 13) // INTERNAL
	}

	resp := CreateCharacterResponse{Character: character}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

// RPCLoadCharacter handles the load_character RPC.
func RPCLoadCharacter(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		CharacterID string `json:"character_id"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	character, err := storage.LoadCharacter(ctx, nk, userID, req.CharacterID)
	if err != nil {
		logger.Error("Failed to load character: %v", err)
		return "", runtime.NewError("failed to load character", 13)
	}
	if character == nil {
		return "", runtime.NewError("character not found", 5) // NOT_FOUND
	}

	data, _ := json.Marshal(map[string]interface{}{"character": character})
	return string(data), nil
}

// RPCListCharacters handles the list_characters RPC.
func RPCListCharacters(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	characters, err := storage.ListCharacters(ctx, nk, userID)
	if err != nil {
		logger.Error("Failed to list characters: %v", err)
		return "", runtime.NewError("failed to list characters", 13)
	}

	data, _ := json.Marshal(map[string]interface{}{"characters": characters})
	return string(data), nil
}

// RPCStartGame creates a server-authoritative match and returns the match ID.
func RPCStartGame(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		CharacterID string `json:"character_id"`
		Language    string `json:"language"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	lang := req.Language
	if lang == "" {
		lang = "en"
	}

	params := map[string]interface{}{
		"character_id": req.CharacterID,
		"user_id":      userID,
		"language":     lang,
	}
	matchID, err := nk.MatchCreate(ctx, "gameplay", params)
	if err != nil {
		logger.Error("Failed to create match: %v", err)
		return "", runtime.NewError("failed to create match", 13)
	}

	data, _ := json.Marshal(map[string]string{"match_id": matchID})
	return string(data), nil
}

// RPCSaveGame handles the save_game RPC.
func RPCSaveGame(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		CharacterID string `json:"character_id"`
		Slot        int    `json:"slot"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	character, err := storage.LoadCharacter(ctx, nk, userID, req.CharacterID)
	if err != nil || character == nil {
		return "", runtime.NewError("character not found", 5)
	}

	if err := storage.SaveGame(ctx, nk, userID, req.Slot, character); err != nil {
		logger.Error("Failed to save game: %v", err)
		return "", runtime.NewError("failed to save game", 13)
	}

	return `{"status": "saved"}`, nil
}

// RPCLoadGame handles the load_game RPC.
func RPCLoadGame(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		Slot int `json:"slot"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	character, err := storage.LoadGame(ctx, nk, userID, req.Slot)
	if err != nil {
		logger.Error("Failed to load game: %v", err)
		return "", runtime.NewError("failed to load game", 13)
	}
	if character == nil {
		return "", runtime.NewError("no save in that slot", 5)
	}

	// Also update the active character
	if err := storage.SaveCharacter(ctx, nk, userID, character); err != nil {
		logger.Error("Failed to update active character: %v", err)
	}

	data, _ := json.Marshal(map[string]interface{}{"character": character})
	return string(data), nil
}
