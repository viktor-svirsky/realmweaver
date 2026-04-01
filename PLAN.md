# Realmweaver — MVP Implementation Plan

**Spec**: `../docs/superpowers/specs/2026-03-31-realmweaver-design.md`
**Scope**: Phase 1 — Solo Core (character creation, starting region, combat, inventory, Claude narration, save/load, mobile UI)

## Project Structure

```
realmweaver/
├── docker-compose.yml          # Nakama + CockroachDB + PostgreSQL
├── nakama/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                 # Plugin entrypoint, registers all hooks
│   ├── Dockerfile              # Builds Go plugin .so
│   ├── claude/
│   │   ├── client.go           # HTTP client for Claude API (streaming)
│   │   ├── context.go          # Builds context objects for Claude prompts
│   │   └── prompts.go          # System prompts and prompt templates
│   ├── engine/
│   │   ├── character.go        # Character creation, stats, leveling
│   │   ├── combat.go           # Combat resolution, initiative, turns
│   │   ├── dice.go             # Dice rolling, modifiers
│   │   ├── inventory.go        # Items, equipment, encumbrance
│   │   ├── skills.go           # Skill checks (stat + proficiency vs DC)
│   │   └── types.go            # Shared types: Character, Item, Stats, etc.
│   ├── world/
│   │   ├── region.go           # Region generation and loading
│   │   ├── starter.go          # Starting region + town + dungeon (hardcoded for MVP)
│   │   └── db.go               # PostgreSQL queries for world data
│   ├── handlers/
│   │   ├── gameplay.go         # Real-time match handler (WebSocket gameplay loop)
│   │   ├── character_rpc.go    # RPC: create/load/save character
│   │   └── action_rpc.go       # RPC: fallback REST endpoints for actions
│   └── storage/
│       └── nakama_storage.go   # Helpers for reading/writing Nakama storage
├── mobile/
│   ├── app.json                # Expo config
│   ├── package.json
│   ├── tsconfig.json
│   ├── App.tsx                 # Root component, navigation
│   ├── src/
│   │   ├── api/
│   │   │   ├── nakama.ts       # Nakama JS client setup + auth
│   │   │   └── socket.ts       # WebSocket connection, message handlers
│   │   ├── screens/
│   │   │   ├── HomeScreen.tsx          # Title screen, load/new game
│   │   │   ├── CharacterCreateScreen.tsx # Name, class, stat allocation
│   │   │   └── GameScreen.tsx          # Main gameplay (narrative + actions)
│   │   ├── components/
│   │   │   ├── NarrativeView.tsx       # Scrolling chat-like narrative
│   │   │   ├── ActionBar.tsx           # Context-sensitive quick actions
│   │   │   ├── CharacterSheet.tsx      # Stats, level, equipment (swipe panel)
│   │   │   ├── Inventory.tsx           # Item grid (swipe panel)
│   │   │   ├── TextInput.tsx           # Player free-text input
│   │   │   └── StreamingText.tsx       # Typewriter animation for Claude output
│   │   ├── state/
│   │   │   └── gameStore.ts    # Zustand store for client-side game state
│   │   └── types/
│   │       └── game.ts         # TypeScript types matching server types
│   └── assets/
├── migrations/
│   └── 001_world_tables.sql    # PostgreSQL schema for world data
├── Makefile
├── CLAUDE.md
└── README.md
```

## Build Order (16 steps)

### Step 1: Project Scaffolding
**Files**: `docker-compose.yml`, `Makefile`, `CLAUDE.md`, `README.md`
- Docker Compose: Nakama 3.x + CockroachDB + PostgreSQL 16
- Makefile targets: `build`, `up`, `down`, `logs`, `test`, `migrate`
- CLAUDE.md with project conventions

### Step 2: Go Module + Plugin Entrypoint
**Files**: `nakama/go.mod`, `nakama/main.go`, `nakama/Dockerfile`
- Initialize Go module
- Plugin entrypoint that registers with Nakama runtime
- Dockerfile that builds the Go plugin `.so` for Nakama
- Verify: `make build && make up` — Nakama starts with plugin loaded

### Step 3: PostgreSQL Migrations
**Files**: `migrations/001_world_tables.sql`
- Create world tables: `regions`, `npcs`, `factions`, `faction_reputation`, `quests`, `quest_progress`, `world_events`, `player_messages`, `combat_log`
- Run via Makefile target: `make migrate`

### Step 4: Game Types
**Files**: `nakama/engine/types.go`
- Define core types: `Character`, `Stats`, `Item`, `Equipment`, `EquipmentSlots`, `DiceRoll`, `CombatResult`, `SkillCheckResult`
- Character classes enum: Warrior, Mage, Rogue, Cleric
- Starting stat arrays per class
- Item rarity, damage types, equipment slots

### Step 5: Dice & Skill Checks
**Files**: `nakama/engine/dice.go`, `nakama/engine/skills.go`
- `RollD20()`, `RollDice(count, sides)`, `RollWithModifier(sides, modifier)`
- `SkillCheck(character, stat, dc)` → `SkillCheckResult`
- Deterministic seed option for testing

### Step 6: Character System
**Files**: `nakama/engine/character.go`
- `NewCharacter(name, class)` — applies class starting stats
- `AllocateStats(character, allocations)` — point-buy system
- `GainXP(character, amount)` — XP tracking + level-up logic
- `LevelUp(character)` — stat increases, HP/mana recalculation
- Derived stat calculation: HP, Mana, AC, Initiative

### Step 7: Inventory & Equipment
**Files**: `nakama/engine/inventory.go`
- `AddItem(character, item)`, `RemoveItem(character, itemID)`
- `EquipItem(character, itemID, slot)`, `UnequipItem(character, slot)`
- Weight/encumbrance check
- Starter equipment per class
- `UseItem(character, itemID)` — consumables (health potions, mana potions)

### Step 8: Combat Engine
**Files**: `nakama/engine/combat.go`
- `InitCombat(character, enemies)` — roll initiative, set up turn order
- `ResolveMeleeAttack(attacker, target)` — d20 + STR vs AC, damage roll
- `ResolveSpellAttack(attacker, target, spell)` — d20 + INT vs AC, spell damage
- `ResolveDefend(character)` — temporary AC bonus
- `ResolveFlee(character, enemies)` — DEX check
- `CheckCombatEnd(combat)` — victory/defeat/ongoing
- Loot table: random item drops on enemy defeat
- XP award calculation

### Step 9: World Data Layer
**Files**: `nakama/world/db.go`, `nakama/world/region.go`, `nakama/world/starter.go`
- PostgreSQL connection pool (via `database/sql` + `lib/pq`)
- CRUD for regions, NPCs
- `starter.go`: hardcoded starting region — town of Millhaven with:
  - Tavern (quest giver NPC), Blacksmith (shop NPC), Town square
  - One dungeon: Goblin Cave (3 rooms, increasing difficulty)
  - 4-5 NPCs with personality prompts
  - Starter quests: "Clear the Goblin Cave", "Find the Blacksmith's Lost Shipment"
- Seed starting region on first server boot

### Step 10: Nakama Storage Helpers
**Files**: `nakama/storage/nakama_storage.go`
- `SaveCharacter(ctx, nk, userID, character)`
- `LoadCharacter(ctx, nk, userID)`
- `SaveGame(ctx, nk, userID, slot, snapshot)`
- `LoadGame(ctx, nk, userID, slot)`
- JSON serialization/deserialization

### Step 11: Claude Client
**Files**: `nakama/claude/client.go`, `nakama/claude/prompts.go`, `nakama/claude/context.go`
- HTTP client calling `ai-proxy.9635783.xyz/v1/messages`
- Streaming support: read SSE events, forward tokens
- `prompts.go`: system prompt for DM personality, structured output instructions
- `context.go`: `BuildContext(character, location, mechanicalResult)` → assembles the context object
- JSON response parsing with fallback to raw text
- Rate limiting: max N calls per user per minute

### Step 12: Character RPCs
**Files**: `nakama/handlers/character_rpc.go`
- Register Nakama RPCs: `create_character`, `load_character`, `save_game`, `load_game`
- Wire up to engine + storage layers
- Verify: call RPCs via Nakama console

### Step 13: Gameplay Match Handler
**Files**: `nakama/handlers/gameplay.go`
- Nakama match handler (real-time): `MatchInit`, `MatchJoinAttempt`, `MatchJoin`, `MatchLeave`, `MatchLoop`, `MatchTerminate`
- Player sends action (free text or quick action) → handler routes to:
  - Combat: resolve via engine, narrate via Claude
  - Exploration: skill checks, describe location
  - NPC interaction: load NPC context, Claude dialogue
  - Inventory: direct engine response (no Claude needed)
- Stream Claude responses back via match data messages
- Game state machine: `exploring`, `in_combat`, `in_dialogue`, `in_shop`

### Step 14: Mobile — Expo Setup + Nakama Client
**Files**: `mobile/package.json`, `mobile/app.json`, `mobile/tsconfig.json`, `mobile/App.tsx`, `mobile/src/api/nakama.ts`, `mobile/src/api/socket.ts`, `mobile/src/types/game.ts`, `mobile/src/state/gameStore.ts`
- Expo init with TypeScript
- `@heroiclabs/nakama-js` client
- Device auth (auto, anonymous)
- WebSocket connection to Nakama match
- Zustand store for local game state (character, narrative history, current location)
- TypeScript types matching Go server types

### Step 15: Mobile — Screens
**Files**: `mobile/src/screens/HomeScreen.tsx`, `mobile/src/screens/CharacterCreateScreen.tsx`, `mobile/src/screens/GameScreen.tsx`
- **HomeScreen**: title, "New Game" / "Continue" buttons
- **CharacterCreateScreen**: name input, class selection (4 cards), stat allocation (point-buy sliders), confirm
- **GameScreen**: main gameplay — hosts NarrativeView, ActionBar, TextInput, swipe panels

### Step 16: Mobile — Game Components
**Files**: `mobile/src/components/NarrativeView.tsx`, `mobile/src/components/ActionBar.tsx`, `mobile/src/components/CharacterSheet.tsx`, `mobile/src/components/Inventory.tsx`, `mobile/src/components/TextInput.tsx`, `mobile/src/components/StreamingText.tsx`
- **NarrativeView**: scrolling FlatList of narrative entries (Claude text + mechanical summaries)
- **StreamingText**: typewriter animation, renders tokens as they arrive
- **ActionBar**: context-sensitive buttons (combat/explore/town/dialogue)
- **CharacterSheet**: swipeable drawer — stats, HP bar, mana bar, equipment slots, level/XP
- **Inventory**: swipeable drawer — grid of items, tap for details, equip/use/drop actions
- **TextInput**: free-text input with send button

## Dependencies

### Go (nakama/)
- `github.com/heroiclabs/nakama-common` — Nakama runtime API
- `github.com/lib/pq` — PostgreSQL driver
- Standard library for HTTP (Claude client), JSON, crypto/rand (dice)

### Mobile (mobile/)
- `expo` — framework
- `@heroiclabs/nakama-js` — Nakama JavaScript client
- `zustand` — state management
- `react-native-gesture-handler` — swipe panels
- `react-native-reanimated` — animations (streaming text, transitions)
- `@react-navigation/native` + `@react-navigation/native-stack` — navigation

## Testing Strategy

### Go
- Unit tests for all engine functions (dice, combat, character, inventory, skills)
- Integration tests for Nakama handlers (using Nakama test runtime)
- Tests run via `make test`

### Mobile
- Component tests with React Native Testing Library
- E2E smoke test: create character → enter world → take action → see narration

## Verification Checkpoints

After each group of steps, verify:

- **Steps 1-3**: `make up` starts all containers, `make migrate` creates tables
- **Steps 4-8**: `make test` — all engine unit tests pass
- **Steps 9-10**: starter region loads, character save/load works
- **Steps 11-13**: create character → join match → send action → receive Claude narration (via Nakama console)
- **Steps 14-16**: mobile app connects to Nakama, create character, play through starter region
