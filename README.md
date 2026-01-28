# Promptlands

AI agents compete for territory. You write the rules.

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18+
- Docker (for PostgreSQL & Redis, optional in dev mode)

### One Command Start

```bash
./run.sh
```

This starts both backend and frontend in dev mode. Open http://localhost:5173

### Alternative: Using Make

```bash
make install  # First time only
make dev      # Start everything
```

### Manual Start

1. Start the backend:
```bash
cd backend
go run ./cmd/server --dev
```

2. In another terminal, start the frontend:
```bash
cd frontend
npm install
npm run dev
```

3. Open http://localhost:5173

### Full Setup (With Database)

1. Start databases:
```bash
docker-compose up -d
```

2. Start the backend:
```bash
cd backend
go run ./cmd/server
```

3. Start the frontend:
```bash
cd frontend
npm run dev
```

## Project Structure

```
promptlands/
├── backend/              # Go server
│   ├── cmd/server/       # Entry point
│   └── internal/
│       ├── api/          # HTTP handlers
│       ├── config/       # Configuration
│       ├── db/           # Database clients
│       ├── game/         # Game engine
│       ├── llm/          # LLM integration
│       └── ws/           # WebSocket hub
├── frontend/             # SvelteKit app
│   └── src/
│       ├── lib/
│       │   ├── components/   # Svelte components
│       │   ├── game/         # PixiJS renderer
│       │   └── stores/       # State management
│       └── routes/           # Pages
├── docker-compose.yml    # PostgreSQL + Redis
└── schema.sql           # Database schema
```

## Configuration

Set your Gemini API key:
```bash
export GEMINI_API_KEY=your_key_here
```

Edit `backend/config.yaml` for other settings.

## API Endpoints

- `GET /health` - Health check
- `GET /api/games` - List games
- `POST /api/games/singleplayer` - Create singleplayer game
- `GET /api/adversaries` - List AI adversary types
- `GET /ws/game/{id}` - WebSocket connection for game updates

## Game Mechanics

### Actions
- **MOVE** - Move one tile (north/south/east/west)
- **CLAIM** - Claim the tile you're standing on
- **MESSAGE** - Send messages to other agents
- **WAIT** - Do nothing this turn

### AI Adversaries
- **Warlord** (aggressive) - Relentless expansion
- **Turtle** (defensive) - Compact territory building
- **Ambassador** (diplomatic) - Peaceful negotiation
- **Wildcard** (chaotic) - Unpredictable behavior
- **Calculator** (methodical) - Optimal efficiency
- **Scout** (explorer) - Map discovery focused
# promptlands
