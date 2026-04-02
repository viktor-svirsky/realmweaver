package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/heroiclabs/nakama-common/runtime"

	dbsql "database/sql"
	_ "github.com/lib/pq"
	"realmweaver/storage"
	"realmweaver/world"
)

var (
	socialWorldDB     *world.WorldDB
	socialWorldDBOnce sync.Once
)

func getSocialWorldDB() *world.WorldDB {
	socialWorldDBOnce.Do(func() {
		dsn := os.Getenv("WORLD_DB_DSN")
		if dsn == "" {
			return
		}
		db, err := dbsql.Open("postgres", dsn)
		if err != nil {
			return
		}
		socialWorldDB = world.NewWorldDB(db)
	})
	return socialWorldDB
}

// RPCGetNearbyPlayers returns players in the same and adjacent regions.
func RPCGetNearbyPlayers(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		RegionX int `json:"region_x"`
		RegionY int `json:"region_y"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	wdb := getSocialWorldDB()
	if wdb == nil {
		return "", runtime.NewError("world DB unavailable", 13)
	}

	players, err := wdb.GetNearbyPlayers(ctx, req.RegionX, req.RegionY, userID)
	if err != nil {
		logger.Error("Failed to get nearby players: %v", err)
		return "", runtime.NewError("failed to get players", 13)
	}

	data, _ := json.Marshal(map[string]interface{}{"players": players})
	return string(data), nil
}

// RPCPostChat sends a chat message in the player's current region.
func RPCPostChat(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		RegionX       int    `json:"region_x"`
		RegionY       int    `json:"region_y"`
		CharacterName string `json:"character_name"`
		Message       string `json:"message"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	if req.Message == "" || len(req.Message) > 500 {
		return "", runtime.NewError("message must be 1-500 characters", 3)
	}

	// Sanitize message — strip control characters (except newline)
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return -1
		}
		return r
	}, req.Message)
	req.Message = cleaned

	// Validate character name belongs to this user
	characters, err := storage.ListCharacters(ctx, nk, userID)
	if err == nil {
		nameValid := false
		for _, c := range characters {
			if c.Name == req.CharacterName {
				nameValid = true
				break
			}
		}
		if !nameValid && len(characters) > 0 {
			req.CharacterName = characters[0].Name // fallback to first character
		}
	}

	wdb := getSocialWorldDB()
	if wdb == nil {
		return "", runtime.NewError("world DB unavailable", 13)
	}

	if err := wdb.PostChatMessage(ctx, req.RegionX, req.RegionY, userID, req.CharacterName, req.Message); err != nil {
		logger.Error("Failed to post chat: %v", err)
		return "", runtime.NewError("failed to post message", 13)
	}

	// Notify all players in the region about the new chat message
	players, _ := wdb.GetPlayersInRegion(ctx, req.RegionX, req.RegionY, userID)
	for _, p := range players {
		content := map[string]interface{}{
			"type":     "chat",
			"sender":   req.CharacterName,
			"message":  req.Message,
			"region_x": req.RegionX,
			"region_y": req.RegionY,
		}
		nk.NotificationSend(ctx, p.UserID, "chat_message", content, 1, "", false)
	}

	return `{"status":"sent"}`, nil
}

// RPCGetChat returns recent chat messages for a region.
func RPCGetChat(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		RegionX int `json:"region_x"`
		RegionY int `json:"region_y"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	wdb := getSocialWorldDB()
	if wdb == nil {
		return "", runtime.NewError("world DB unavailable", 13)
	}

	messages, err := wdb.GetChatMessages(ctx, req.RegionX, req.RegionY, 50)
	if err != nil {
		logger.Error("Failed to get chat: %v", err)
		return "", runtime.NewError("failed to get messages", 13)
	}

	data, _ := json.Marshal(map[string]interface{}{"messages": messages})
	return string(data), nil
}

// RPCUpdateLocation updates the player's location in the world (called by match handler).
func RPCUpdateLocation(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var loc world.PlayerLocation
	if err := json.Unmarshal([]byte(payload), &loc); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}
	loc.UserID = userID

	wdb := getSocialWorldDB()
	if wdb == nil {
		return "", runtime.NewError("world DB unavailable", 13)
	}

	if err := wdb.UpdatePlayerLocation(ctx, &loc); err != nil {
		logger.Error("Failed to update location: %v", err)
		return "", runtime.NewError("failed to update location", 13)
	}

	return `{"status":"updated"}`, nil
}
