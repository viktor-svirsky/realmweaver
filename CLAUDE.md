# Realmweaver

AI Dungeon Master mobile RPG. Nakama game server (Go) + React Native (Expo) client.

## Architecture

- **Nakama server** with Go runtime plugin — handles auth, WebSocket, storage, matchmaking
- **Go plugin** — game engine (combat, dice, stats, inventory), Claude integration, world management
- **PostgreSQL** — world data (regions, NPCs, factions, quests)
- **Nakama storage** — player data (characters, saves)
- **React Native / Expo** — mobile client

## Key Principle

Go game engine = rules and state. Claude = storytelling.
Claude never determines mechanical outcomes (HP, damage, success/failure). The engine resolves mechanics, then Claude narrates.

## Project Structure

```
nakama/           # Go plugin (game server)
  engine/         # Game mechanics (combat, dice, character, inventory)
  claude/         # Claude API client, context builder, prompts
  world/          # Region generation, world DB queries
  handlers/       # Nakama match handlers and RPCs
  storage/        # Nakama storage helpers
mobile/           # React Native / Expo app
  src/api/        # Nakama client + WebSocket
  src/screens/    # App screens
  src/components/ # UI components
  src/state/      # Zustand store
migrations/       # PostgreSQL world data schema
```

## Commands

```bash
make build    # Build Go plugin Docker image
make up       # Start Nakama + CockroachDB + PostgreSQL
make down     # Stop services
make logs     # Follow Nakama logs
make test     # Run Go tests
make migrate  # Apply PostgreSQL migrations
make clean    # Remove all data volumes
```

## Code Style

### Go (nakama/)
- Standard Go formatting (gofmt)
- Error handling: always check and return errors, never ignore
- Types in `engine/types.go`, shared across packages
- Tests next to source files: `foo_test.go`

### TypeScript (mobile/)
- Strict TypeScript, no `any`
- Functional components with hooks
- Zustand for state management
- Types in `src/types/game.ts` matching server types

## Environment

- Nakama console: http://localhost:7351
- Nakama HTTP API: http://localhost:7350
- Nakama gRPC: localhost:7349
- PostgreSQL (world): localhost:5433
- CockroachDB: localhost:26257

## Claude API

- Proxy: `https://ai-proxy.9635783.xyz/v1/messages`
- Model: `claude-sonnet-4-20250514` (configurable via env)
- Streaming enabled for narration
