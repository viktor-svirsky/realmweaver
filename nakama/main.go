package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	_ "github.com/lib/pq"

	"realmweaver/handlers"
	"realmweaver/world"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	// Connect to world database (PostgreSQL) and seed starter content
	worldDSN := os.Getenv("WORLD_DB_DSN")
	if worldDSN == "" {
		logger.Warn("WORLD_DB_DSN not set, world features will be unavailable")
	} else {
		worldDB, err := sql.Open("postgres", worldDSN)
		if err != nil {
			logger.Error("Failed to connect to world database: %v", err)
			return err
		}
		if err := worldDB.Ping(); err != nil {
			logger.Error("Failed to ping world database: %v", err)
			return err
		}
		logger.Info("Connected to world database")

		// Seed starter region
		wdb := world.NewWorldDB(worldDB)
		if err := world.SeedStarterRegion(ctx, wdb); err != nil {
			logger.Error("Failed to seed starter region: %v", err)
		} else {
			logger.Info("Starter region ready")
		}
	}

	// Register RPCs
	rpcs := map[string]func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error){
		"create_character": handlers.RPCCreateCharacter,
		"load_character":   handlers.RPCLoadCharacter,
		"list_characters":  handlers.RPCListCharacters,
		"start_game":       handlers.RPCStartGame,
		"get_nearby_players": handlers.RPCGetNearbyPlayers,
		"post_chat":         handlers.RPCPostChat,
		"get_chat":          handlers.RPCGetChat,
		"update_location":   handlers.RPCUpdateLocation,
		"trade_offer":       handlers.RPCTradeOffer,
		"trade_accept":      handlers.RPCAcceptTrade,
		"trade_list":        handlers.RPCListTrades,
		"pvp_challenge":     handlers.RPCPvPChallenge,
		"coop_help":         handlers.RPCCoopHelp,
		"travel":            handlers.RPCTravel,
		"save_game":         handlers.RPCSaveGame,
		"load_game":        handlers.RPCLoadGame,
	}
	for name, fn := range rpcs {
		if err := initializer.RegisterRpc(name, fn); err != nil {
			return err
		}
	}

	// Register match handler
	if err := initializer.RegisterMatch("gameplay", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &handlers.GameplayMatch{}, nil
	}); err != nil {
		return err
	}

	logger.Info("Realmweaver plugin loaded in %d ms", time.Since(initStart).Milliseconds())
	return nil
}
