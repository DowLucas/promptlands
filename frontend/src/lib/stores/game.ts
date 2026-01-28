import { writable, derived, type Writable } from 'svelte/store';
import type { Agent, FullGameState, GameMessage, Tile, TickUpdate, WorldSnapshot, ActionResult } from '$lib/types';

// Log entries store
export interface LogEntry {
	tick: number;
	type: 'move' | 'claim' | 'message' | 'wait' | 'info' | 'error';
	agentId: string;
	agentName: string;
	text: string;
	isPlayer: boolean;
}

export const logEntries = writable<LogEntry[]>([]);

// Game state store
interface GameState {
	gameId: string | null;
	status: 'waiting' | 'running' | 'finished' | null;
	tick: number;
	world: WorldSnapshot | null;
	agents: Agent[];
	messages: GameMessage[];
	playerAgentId: string | null;
}

const initialState: GameState = {
	gameId: null,
	status: null,
	tick: 0,
	world: null,
	agents: [],
	messages: [],
	playerAgentId: null
};

export const gameState: Writable<GameState> = writable(initialState);

// Actions
export function setFullState(state: FullGameState) {
	gameState.update(s => ({
		...s,
		gameId: state.game_id,
		status: state.status,
		tick: state.tick,
		world: state.world,
		agents: state.agents,
		messages: state.messages
	}));
}

export function applyTickUpdate(update: TickUpdate) {
	let currentState: GameState;
	gameState.subscribe(s => currentState = s)();

	gameState.update(s => {
		if (!s.world) return s;

		// Update tiles
		const tileMap = new Map<string, Tile>();
		for (const tile of s.world.tiles) {
			tileMap.set(`${tile.x},${tile.y}`, tile);
		}
		for (const change of update.changes.tiles) {
			const key = `${change.x},${change.y}`;
			const existing = tileMap.get(key);
			if (existing) {
				existing.owner_id = change.owner_id;
			}
		}

		// Update agents
		const newAgents = update.changes.agents;

		// Append new messages
		const newMessages = [...s.messages, ...update.changes.messages];

		return {
			...s,
			tick: update.tick,
			world: {
				...s.world,
				tiles: Array.from(tileMap.values())
			},
			agents: newAgents,
			messages: newMessages
		};
	});

	// Add to game log
	addToLog(update.tick, update.changes.results, update.changes.messages, currentState!);
}

function addToLog(tick: number, results: ActionResult[], messages: GameMessage[], state: GameState) {
	const getAgentName = (id: string) => state.agents.find(a => a.id === id)?.name || 'Unknown';
	const isPlayer = (id: string) => id === state.playerAgentId;

	const newEntries: LogEntry[] = [];

	// Process action results
	for (const result of results) {
		const agentName = getAgentName(result.agent_id);
		let type: LogEntry['type'] = 'info';
		let text = '';

		switch (result.action) {
			case 'MOVE':
				type = 'move';
				if (result.success && result.new_pos) {
					text = `moved to (${result.new_pos.x}, ${result.new_pos.y})`;
				} else {
					type = 'error';
					text = `failed to move: ${result.message}`;
				}
				break;
			case 'CLAIM':
				type = 'claim';
				if (result.success && result.claimed_at) {
					text = `claimed tile at (${result.claimed_at.x}, ${result.claimed_at.y})`;
				} else {
					type = 'error';
					text = `failed to claim: ${result.message}`;
				}
				break;
			case 'WAIT':
				type = 'wait';
				text = 'is waiting...';
				break;
			case 'MESSAGE':
				continue; // Handle below
			default:
				text = result.message || 'unknown action';
		}

		newEntries.push({
			tick,
			type,
			agentId: result.agent_id,
			agentName,
			text,
			isPlayer: isPlayer(result.agent_id)
		});
	}

	// Process messages
	for (const msg of messages) {
		newEntries.push({
			tick: msg.tick,
			type: 'message',
			agentId: msg.from_agent_id,
			agentName: getAgentName(msg.from_agent_id),
			text: msg.to_agent_id ? `says: "${msg.content}"` : `broadcasts: "${msg.content}"`,
			isPlayer: isPlayer(msg.from_agent_id)
		});
	}

	logEntries.update(entries => [...entries, ...newEntries].slice(-200));
}

export function setPlayerAgentId(agentId: string) {
	gameState.update(s => ({ ...s, playerAgentId: agentId }));
}

export function resetGame() {
	gameState.set(initialState);
	logEntries.set([]);
}

// Derived stores
export const currentTick = derived(gameState, $s => $s.tick);
export const gameStatus = derived(gameState, $s => $s.status);
export const agents = derived(gameState, $s => $s.agents);
export const messages = derived(gameState, $s => $s.messages);

// Get tile ownership map for rendering
export const tileOwnership = derived(gameState, $s => {
	const map = new Map<string, string | null>();
	if ($s.world) {
		for (const tile of $s.world.tiles) {
			map.set(`${tile.x},${tile.y}`, tile.owner_id);
		}
	}
	return map;
});

// Get ownership counts
export const ownershipCounts = derived(gameState, $s => {
	const counts = new Map<string, number>();
	if ($s.world) {
		for (const tile of $s.world.tiles) {
			if (tile.owner_id) {
				counts.set(tile.owner_id, (counts.get(tile.owner_id) || 0) + 1);
			}
		}
	}
	return counts;
});
