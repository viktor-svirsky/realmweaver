package world

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// WorldDB wraps the PostgreSQL connection for world data operations.
type WorldDB struct {
	db *sql.DB
}

// NewWorldDB creates a new WorldDB instance.
func NewWorldDB(db *sql.DB) *WorldDB {
	return &WorldDB{db: db}
}

// Region represents a world region from the database.
type Region struct {
	ID          int             `json:"id"`
	X           int             `json:"x"`
	Y           int             `json:"y"`
	Biome       string          `json:"biome"`
	Difficulty  int             `json:"difficulty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Lore        string          `json:"lore"`
	Structures  json.RawMessage `json:"structures"`
}

// NPC represents a non-player character from the database.
type NPC struct {
	ID              int             `json:"id"`
	RegionID        int             `json:"region_id"`
	Name            string          `json:"name"`
	Race            string          `json:"race"`
	Occupation      string          `json:"occupation"`
	PersonalityPrompt string        `json:"personality_prompt"`
	Disposition     json.RawMessage `json:"disposition"`
	MemoryTags      json.RawMessage `json:"memory_tags"`
	FactionID       *int            `json:"faction_id,omitempty"`
	Alive           bool            `json:"alive"`
	LocationDetail  string          `json:"location_detail"`
}

// GetRegion loads a region by coordinates.
func (w *WorldDB) GetRegion(ctx context.Context, x, y int) (*Region, error) {
	r := &Region{}
	err := w.db.QueryRowContext(ctx,
		`SELECT id, x, y, biome, difficulty, name, description, lore, structures FROM regions WHERE x = $1 AND y = $2`,
		x, y,
	).Scan(&r.ID, &r.X, &r.Y, &r.Biome, &r.Difficulty, &r.Name, &r.Description, &r.Lore, &r.Structures)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query region (%d,%d): %w", x, y, err)
	}
	return r, nil
}

// InsertRegion creates a new region.
func (w *WorldDB) InsertRegion(ctx context.Context, r *Region) error {
	structures, _ := json.Marshal(r.Structures)
	if r.Structures == nil {
		structures = []byte("[]")
	}
	return w.db.QueryRowContext(ctx,
		`INSERT INTO regions (x, y, biome, difficulty, name, description, lore, structures)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		r.X, r.Y, r.Biome, r.Difficulty, r.Name, r.Description, r.Lore, structures,
	).Scan(&r.ID)
}

// GetNPCsInRegion loads all alive NPCs in a region.
func (w *WorldDB) GetNPCsInRegion(ctx context.Context, regionID int) ([]NPC, error) {
	rows, err := w.db.QueryContext(ctx,
		`SELECT id, region_id, name, race, occupation, personality_prompt, disposition, memory_tags, faction_id, alive, location_detail
		 FROM npcs WHERE region_id = $1 AND alive = true`,
		regionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query NPCs for region %d: %w", regionID, err)
	}
	defer rows.Close()

	var npcs []NPC
	for rows.Next() {
		n := NPC{}
		if err := rows.Scan(&n.ID, &n.RegionID, &n.Name, &n.Race, &n.Occupation, &n.PersonalityPrompt,
			&n.Disposition, &n.MemoryTags, &n.FactionID, &n.Alive, &n.LocationDetail); err != nil {
			return nil, err
		}
		npcs = append(npcs, n)
	}
	return npcs, rows.Err()
}

// GetNPC loads a single NPC by ID.
func (w *WorldDB) GetNPC(ctx context.Context, id int) (*NPC, error) {
	n := &NPC{}
	err := w.db.QueryRowContext(ctx,
		`SELECT id, region_id, name, race, occupation, personality_prompt, disposition, memory_tags, faction_id, alive, location_detail
		 FROM npcs WHERE id = $1`,
		id,
	).Scan(&n.ID, &n.RegionID, &n.Name, &n.Race, &n.Occupation, &n.PersonalityPrompt,
		&n.Disposition, &n.MemoryTags, &n.FactionID, &n.Alive, &n.LocationDetail)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query NPC %d: %w", id, err)
	}
	return n, nil
}

// InsertNPC creates a new NPC.
func (w *WorldDB) InsertNPC(ctx context.Context, n *NPC) error {
	disposition := n.Disposition
	if disposition == nil {
		disposition = json.RawMessage("{}")
	}
	memoryTags := n.MemoryTags
	if memoryTags == nil {
		memoryTags = json.RawMessage("[]")
	}
	return w.db.QueryRowContext(ctx,
		`INSERT INTO npcs (region_id, name, race, occupation, personality_prompt, disposition, memory_tags, faction_id, alive, location_detail)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		n.RegionID, n.Name, n.Race, n.Occupation, n.PersonalityPrompt, disposition, memoryTags, n.FactionID, n.Alive, n.LocationDetail,
	).Scan(&n.ID)
}

// PlayerLocation represents a player's position in the world.
type PlayerLocation struct {
	UserID         string `json:"user_id"`
	CharacterID    string `json:"character_id"`
	CharacterName  string `json:"character_name"`
	CharacterClass string `json:"character_class"`
	CharacterLevel int    `json:"character_level"`
	RegionX        int    `json:"region_x"`
	RegionY        int    `json:"region_y"`
	Karma          int    `json:"karma"`
}

// UpdatePlayerLocation upserts a player's current location.
func (w *WorldDB) UpdatePlayerLocation(ctx context.Context, loc *PlayerLocation) error {
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO player_locations (user_id, character_id, character_name, character_class, character_level, region_x, region_y, karma, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET
		   character_id = $2, character_name = $3, character_class = $4, character_level = $5,
		   region_x = $6, region_y = $7, karma = $8, updated_at = NOW()`,
		loc.UserID, loc.CharacterID, loc.CharacterName, loc.CharacterClass, loc.CharacterLevel, loc.RegionX, loc.RegionY, loc.Karma,
	)
	return err
}

// GetNearbyPlayers returns players in the same or adjacent regions, active in the last 5 minutes.
func (w *WorldDB) GetNearbyPlayers(ctx context.Context, x, y int, excludeUserID string) ([]PlayerLocation, error) {
	rows, err := w.db.QueryContext(ctx,
		`SELECT user_id, character_id, character_name, character_class, character_level, region_x, region_y, karma
		 FROM player_locations
		 WHERE ABS(region_x - $1) <= 1 AND ABS(region_y - $2) <= 1
		   AND user_id != $3
		   AND updated_at > NOW() - INTERVAL '5 minutes'
		 ORDER BY updated_at DESC
		 LIMIT 50`,
		x, y, excludeUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerLocation
	for rows.Next() {
		var p PlayerLocation
		if err := rows.Scan(&p.UserID, &p.CharacterID, &p.CharacterName, &p.CharacterClass, &p.CharacterLevel, &p.RegionX, &p.RegionY, &p.Karma); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// GetPlayersInRegion returns players in a specific region.
func (w *WorldDB) GetPlayersInRegion(ctx context.Context, x, y int, excludeUserID string) ([]PlayerLocation, error) {
	rows, err := w.db.QueryContext(ctx,
		`SELECT user_id, character_name, character_class, character_level, region_x, region_y, karma
		 FROM player_locations
		 WHERE region_x = $1 AND region_y = $2
		   AND user_id != $3
		   AND updated_at > NOW() - INTERVAL '5 minutes'
		 ORDER BY updated_at DESC
		 LIMIT 20`,
		x, y, excludeUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerLocation
	for rows.Next() {
		var p PlayerLocation
		if err := rows.Scan(&p.UserID, &p.CharacterName, &p.CharacterClass, &p.CharacterLevel, &p.RegionX, &p.RegionY, &p.Karma); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// ChatMessage represents a message in a region.
type ChatMessage struct {
	ID            int    `json:"id"`
	RegionX       int    `json:"region_x"`
	RegionY       int    `json:"region_y"`
	UserID        string `json:"user_id"`
	CharacterName string `json:"character_name"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
}

// PostChatMessage sends a message in a region.
func (w *WorldDB) PostChatMessage(ctx context.Context, regionX, regionY int, userID, characterName, content string) error {
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO player_messages (region_id, user_id, content, created_at)
		 VALUES ((SELECT id FROM regions WHERE x = $1 AND y = $2), $3, $4, NOW())`,
		regionX, regionY, userID, characterName+": "+content,
	)
	return err
}

// GetChatMessages returns recent messages in a region.
func (w *WorldDB) GetChatMessages(ctx context.Context, regionX, regionY int, limit int) ([]ChatMessage, error) {
	rows, err := w.db.QueryContext(ctx,
		`SELECT pm.id, $1::int as region_x, $2::int as region_y, pm.user_id, pm.content, pm.created_at::text
		 FROM player_messages pm
		 JOIN regions r ON r.id = pm.region_id
		 WHERE r.x = $1 AND r.y = $2
		 ORDER BY pm.created_at DESC
		 LIMIT $3`,
		regionX, regionY, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.RegionX, &m.RegionY, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	// Reverse to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, rows.Err()
}

// RemovePlayerLocation deletes a player's location (on disconnect).
func (w *WorldDB) RemovePlayerLocation(ctx context.Context, userID string) error {
	_, err := w.db.ExecContext(ctx, `DELETE FROM player_locations WHERE user_id = $1`, userID)
	return err
}

// RegionExists checks if a region at coordinates already exists.
func (w *WorldDB) RegionExists(ctx context.Context, x, y int) (bool, error) {
	var exists bool
	err := w.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM regions WHERE x = $1 AND y = $2)`, x, y,
	).Scan(&exists)
	return exists, err
}
