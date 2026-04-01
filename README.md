# Realmweaver

An AI-powered text adventure RPG where Claude is the Dungeon Master. Explore a persistent shared world, fight with real RPG mechanics, and experience dynamic AI-narrated storytelling.

## Features

- **Real RPG mechanics** — d20 combat, stats, skill checks, inventory, leveling
- **AI Dungeon Master** — Claude narrates your adventure, voices NPCs, generates quests
- **Persistent world** — your choices permanently change the world
- **Mobile-first** — React Native app for iOS and Android

## Tech Stack

- **Server**: [Nakama](https://heroiclabs.com/nakama/) game server with Go runtime plugin
- **AI**: Anthropic Claude API (streaming)
- **Database**: PostgreSQL (world data) + Nakama storage (player data)
- **Client**: React Native / Expo

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 20+
- An Anthropic API key

### Server

```bash
# Start all services
make up

# Check logs
make logs

# Nakama console at http://localhost:7351
```

### Mobile

```bash
cd mobile
npm install
npx expo start
```

## Configuration

Set these environment variables in `docker-compose.yml`:

| Variable | Description | Default |
|----------|-------------|---------|
| `CLAUDE_API_URL` | Claude API endpoint | `https://ai-proxy.9635783.xyz/v1/messages` |
| `CLAUDE_API_KEY` | Anthropic API key | `sk-local-proxy` |
| `CLAUDE_MODEL` | Claude model to use | `claude-sonnet-4-20250514` |

## License

MIT
