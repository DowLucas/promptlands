// Game types matching backend

export interface Position {
	x: number;
	y: number;
}

export interface Tile {
	x: number;
	y: number;
	owner_id: string | null;
	terrain: 'plains' | 'forest' | 'mountain' | 'water';
}

export interface Agent {
	id: string;
	name: string;
	position: Position;
	is_adversary: boolean;
	adversary_type?: string;
}

export interface GameMessage {
	tick: number;
	from_agent_id: string;
	to_agent_id: string | null;
	content: string;
}

export interface ActionResult {
	agent_id: string;
	action: string;
	success: boolean;
	message?: string;
	old_pos?: Position;
	new_pos?: Position;
	claimed_at?: Position;
}

export interface TileChange {
	x: number;
	y: number;
	owner_id: string | null;
}

export interface TickChanges {
	tiles: TileChange[];
	agents: Agent[];
	messages: GameMessage[];
	results: ActionResult[];
}

export interface TickUpdate {
	type: 'tick';
	tick: number;
	game_id: string;
	changes: TickChanges;
}

export interface WorldSnapshot {
	size: number;
	seed: number;
	tiles: Tile[];
}

export interface FullGameState {
	type: 'full_state';
	game_id: string;
	tick: number;
	status: 'waiting' | 'running' | 'finished';
	world: WorldSnapshot;
	agents: Agent[];
	messages: GameMessage[];
}

export interface GameInfo {
	id: string;
	status: 'waiting' | 'running' | 'finished';
	player_count: number;
	max_players: number;
	tick: number;
}

export interface AdversaryType {
	type: string;
	name: string;
}

export type WebSocketMessage = TickUpdate | FullGameState | { type: 'pong' } | { type: 'game_over'; winner: string; scores: Record<string, number> };
