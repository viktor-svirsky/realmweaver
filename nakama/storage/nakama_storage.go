package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/engine"
)

const (
	CollectionCharacters = "characters"
	CollectionSaves      = "saves"
)

// SaveCharacter persists a character to Nakama storage.
func SaveCharacter(ctx context.Context, nk runtime.NakamaModule, userID string, character *engine.Character) error {
	data, err := json.Marshal(character)
	if err != nil {
		return fmt.Errorf("marshal character: %w", err)
	}

	ops := []*runtime.StorageWrite{
		{
			Collection:      CollectionCharacters,
			Key:             character.ID,
			UserID:          userID,
			Value:           string(data),
			PermissionRead:  1, // owner only
			PermissionWrite: 1, // owner only
		},
	}

	_, err = nk.StorageWrite(ctx, ops)
	if err != nil {
		return fmt.Errorf("write character: %w", err)
	}
	return nil
}

// LoadCharacter loads a character from Nakama storage.
func LoadCharacter(ctx context.Context, nk runtime.NakamaModule, userID, characterID string) (*engine.Character, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: CollectionCharacters,
			Key:        characterID,
			UserID:     userID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("read character: %w", err)
	}
	if len(objects) == 0 {
		return nil, nil
	}

	var character engine.Character
	if err := json.Unmarshal([]byte(objects[0].Value), &character); err != nil {
		return nil, fmt.Errorf("unmarshal character: %w", err)
	}
	return &character, nil
}

// ListCharacters returns all characters for a user.
func ListCharacters(ctx context.Context, nk runtime.NakamaModule, userID string) ([]*engine.Character, error) {
	objects, _, err := nk.StorageList(ctx, "", userID, CollectionCharacters, 10, "")
	if err != nil {
		return nil, fmt.Errorf("list characters: %w", err)
	}

	characters := make([]*engine.Character, 0, len(objects))
	for _, obj := range objects {
		var c engine.Character
		if err := json.Unmarshal([]byte(obj.Value), &c); err != nil {
			continue
		}
		characters = append(characters, &c)
	}
	return characters, nil
}

// SaveGame saves a full game snapshot.
func SaveGame(ctx context.Context, nk runtime.NakamaModule, userID string, slot int, character *engine.Character) error {
	data, err := json.Marshal(character)
	if err != nil {
		return fmt.Errorf("marshal save: %w", err)
	}

	ops := []*runtime.StorageWrite{
		{
			Collection:      CollectionSaves,
			Key:             fmt.Sprintf("slot_%d", slot),
			UserID:          userID,
			Value:           string(data),
			PermissionRead:  1,
			PermissionWrite: 1,
		},
	}

	_, err = nk.StorageWrite(ctx, ops)
	if err != nil {
		return fmt.Errorf("write save: %w", err)
	}
	return nil
}

// LoadGame loads a game save from a slot.
func LoadGame(ctx context.Context, nk runtime.NakamaModule, userID string, slot int) (*engine.Character, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: CollectionSaves,
			Key:        fmt.Sprintf("slot_%d", slot),
			UserID:     userID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("read save: %w", err)
	}
	if len(objects) == 0 {
		return nil, nil
	}

	var character engine.Character
	if err := json.Unmarshal([]byte(objects[0].Value), &character); err != nil {
		return nil, fmt.Errorf("unmarshal save: %w", err)
	}
	return &character, nil
}

// ListSaves returns metadata about all save slots for a user.
func ListSaves(ctx context.Context, nk runtime.NakamaModule, userID string) ([]*api.StorageObject, error) {
	objects, _, err := nk.StorageList(ctx, "", userID, CollectionSaves, 10, "")
	if err != nil {
		return nil, fmt.Errorf("list saves: %w", err)
	}
	return objects, nil
}
