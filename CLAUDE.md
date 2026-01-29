# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Promptlands is a real-time AI territory-competition game. Players write system prompts for AI agents that compete for territory on a procedurally generated map. The backend runs a tick-based game loop, agents decide actions via LLM calls, and the frontend renders the world with PixiJS over WebSocket.

## Commands

### Development (no database required)

```bash
./run.sh                    # Start backend + frontend (reads backend/.env for GEMINI_API_KEY)
make dev                    # Alternative: parallel backend + frontend
```

- Backend: http://localhost:8080
- Frontend: http://localhost:5173

### Backend only

```bash
cd backend
go run ./cmd/server --dev      # Mock LLM, no database
go run ./cmd/server --no-db    # Real LLM (needs GEMINI_API_KEY), no database
go run ./cmd/server            # Full mode with PostgreSQL + Redis
```

### Frontend only

```bash
cd frontend
npm run dev                    # Vite dev server
npm run check                  # TypeScript + Svelte type checking
npm run check:watch            # Type checking in watch mode
```

### Tests

```bash
cd backend && go test ./...                                    # All backend tests
cd backend && go test ./internal/game/actions/                 # Action handler tests
cd backend && go test ./internal/game/worldgen/                # World generation tests
cd backend && go test -run TestMoveHandler ./internal/game/actions/  # Single test
```

### Build & Infrastructure

```bash
make install       # Install all dependencies (go mod tidy + npm install)
make build         # Production build (Go binary + Vite build)
make db-up         # Start PostgreSQL + Redis via Docker Compose
make db-down       # Stop databases
make run           # Start with database (production-like)
make clean         # Remove build artifacts
```

## Architecture

### Stack

- **Backend:** Go 1.23, standard library `net/http`, Gorilla WebSocket, pgx (PostgreSQL), go-redis
- **Frontend:** SvelteKit 2, Svelte 5, TypeScript, PixiJS 8 (WebGL rendering), Vite 6
- **Infrastructure:** PostgreSQL 16, Redis 7 (both via Docker Compose)
- **AI:** Google Gemini API (configurable in `backend/config.yaml`)

### Communication

REST API for game creation/listing (`/api/*`), WebSocket for real-time game state (`/ws/game/{id}?player_agent_id={uuid}`). The server broadcasts full game state to connected clients after each tick. Fog-of-war is server-enforced — the WebSocket payload includes a `visible_tiles` array scoped to the player.

### Backend Structure (`backend/internal/`)

- **`api/`** — HTTP handlers and route registration
- **`config/`** — YAML config loading from `backend/config.yaml`
- **`db/`** — PostgreSQL and Redis client setup
- **`game/engine.go`** — Core game loop: processes ticks, resolves conflicts, executes agent actions
- **`game/manager.go`** — Manages multiple concurrent game instances
- **`game/agent.go`** — Agent state, LLM prompt construction, action parsing
- **`game/conflict.go`** — Simultaneous action conflict resolution
- **`game/actions/`** — Action handler registry pattern. Each handler implements `ActionHandler` interface (defined in `game` package to avoid import cycles). Add new actions by creating a `*_handler.go` and registering in `all_handlers.go`
- **`game/worldgen/`** — Procedural map generation using OpenSimplex noise with biome system. Map presets in `map_presets.go`, biome config in `map_config.go`
- **`game/testutil/`** — Shared test helpers for game tests
- **`llm/`** — Gemini API client for agent decision-making
- **`ws/`** — WebSocket hub (broadcast) and per-connection handler

### Frontend Structure (`frontend/src/lib/`)

- **`stores/game.ts`** — Central game state store. Tiles indexed as `Map<string, Tile>` keyed by `"x,y"` for O(1) lookup. Tracks explored tiles and visibility state
- **`stores/ws.ts`** — WebSocket connection management, reconnection logic
- **`stores/inventory.ts`** — Player inventory state
- **`game/renderer.ts`** — PixiJS rendering pipeline: tile grid, agent sprites, fog-of-war overlay, smooth movement animation
- **`game/visibility.ts`** — Client-side fog-of-war calculation (3 states: unexplored, explored/dark, visible/bright)
- **`components/`** — Svelte 5 components (WorldMap, GameControls, InventoryPanel, AgentEditor, etc.)
- **`types.ts`** — Shared TypeScript interfaces for game state

### Key Patterns

- **Action Handler Registry:** Backend action handlers implement `game.ActionHandler` interface with `Handle(ctx ActionContext) ActionResult`. The `ActionContext` provides access to the game world, agent state, and balance config. Register new handlers in `actions/all_handlers.go`
- **Tick-based loop:** The engine runs on a configurable tick interval. Each tick: LLM decides agent actions → conflicts resolved → actions executed → state broadcast via WebSocket
- **Config-driven balance:** All game balance values (HP, damage, energy costs, vision radius, upgrade costs) live in `backend/config.yaml` under `balance`

## Configuration

- **`backend/config.yaml`** — Server port, game settings, balance values, LLM provider/model, database URLs, dev mode flags
- **`backend/config/loot_tables.json`** — Item drop rates per biome
- **`backend/.env`** — Environment variables (GEMINI_API_KEY); loaded by `run.sh`
- **`schema.sql`** — Database schema (players, games, agents, tiles, events, messages)
