-- Promptlands Database Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Players table
CREATE TABLE players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Games table
CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status TEXT DEFAULT 'waiting' CHECK (status IN ('waiting', 'running', 'finished')),
    config JSONB DEFAULT '{}',
    current_tick INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);

-- Agents table
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    position_x INT NOT NULL,
    position_y INT NOT NULL,
    memory JSONB DEFAULT '[]',
    is_adversary BOOLEAN DEFAULT FALSE,
    adversary_type TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agents_game_id ON agents(game_id);

-- Tiles table (for persistence, live state is in memory/Redis)
CREATE TABLE tiles (
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    x INT NOT NULL,
    y INT NOT NULL,
    owner_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    terrain TEXT DEFAULT 'plains',
    PRIMARY KEY (game_id, x, y)
);

CREATE INDEX idx_tiles_game_owner ON tiles(game_id, owner_id);

-- Game events log (for replay functionality)
CREATE TABLE game_events (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    tick INT NOT NULL,
    event_type TEXT NOT NULL,
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_game_events_game_tick ON game_events(game_id, tick);

-- Messages between agents
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    tick INT NOT NULL,
    from_agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    to_agent_id UUID REFERENCES agents(id) ON DELETE CASCADE, -- NULL = broadcast
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_game_tick ON messages(game_id, tick);
