import { writable, derived, type Writable } from 'svelte/store';
import type { Agent, FullGameState, GameMessage, Tile, TickUpdate, WorldSnapshot, ActionResult, WorldObject } from '$lib/types';

// Log entries store
export interface LogEntry {
	tick: number;
	type: 'move' | 'claim' | 'message' | 'wait' | 'info' | 'error' | 'combat' | 'item' | 'craft' | 'harvest' | 'interact' | 'upgrade' | 'death' | 'respawn';
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
	worldObjects: WorldObject[];
}

const initialState: GameState = {
	gameId: null,
	status: null,
	tick: 0,
	world: null,
	agents: [],
	messages: [],
	playerAgentId: null,
	worldObjects: []
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
		messages: state.messages,
		worldObjects: state.world_objects || []
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

		// Update world objects
		let newWorldObjects = [...s.worldObjects];

		// Remove objects
		if (update.changes.objects_removed) {
			const removedIds = new Set(update.changes.objects_removed);
			newWorldObjects = newWorldObjects.filter(obj => !removedIds.has(obj.id));
		}

		// Add new objects
		if (update.changes.objects_added) {
			newWorldObjects = [...newWorldObjects, ...update.changes.objects_added];
		}

		return {
			...s,
			tick: update.tick,
			world: {
				...s.world,
				tiles: Array.from(tileMap.values())
			},
			agents: newAgents,
			messages: newMessages,
			worldObjects: newWorldObjects
		};
	});

	// Add to game log
	addToLog(update.tick, update.changes.results, update.changes.messages, update.changes.respawned || [], currentState!);
}

function addToLog(tick: number, results: ActionResult[], messages: GameMessage[], respawned: string[], state: GameState) {
	const getAgentName = (id: string) => state.agents.find(a => a.id === id)?.name || 'Unknown';
	const isPlayer = (id: string) => id === state.playerAgentId;

	const newEntries: LogEntry[] = [];

	// Process respawns
	for (const agentId of respawned) {
		newEntries.push({
			tick,
			type: 'respawn',
			agentId,
			agentName: getAgentName(agentId),
			text: 'respawned',
			isPlayer: isPlayer(agentId)
		});
	}

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
			case 'HOLD':
				type = 'wait';
				text = 'is waiting...';
				break;
			case 'FIGHT':
				type = 'combat';
				if (result.success) {
					text = result.message || `dealt ${result.damage_dealt} damage`;
				} else {
					type = 'error';
					text = `failed to attack: ${result.message}`;
				}
				break;
			case 'PICKUP':
				type = 'item';
				if (result.success) {
					text = `picked up ${result.item_quantity || 1} ${result.item_id}`;
				} else {
					type = 'error';
					text = `failed to pick up: ${result.message}`;
				}
				break;
			case 'DROP':
				type = 'item';
				if (result.success) {
					text = `dropped ${result.item_quantity || 1} ${result.item_id}`;
				} else {
					type = 'error';
					text = `failed to drop: ${result.message}`;
				}
				break;
			case 'USE':
				type = 'item';
				if (result.success) {
					text = `used ${result.item_id}: ${result.message}`;
				} else {
					type = 'error';
					text = `failed to use: ${result.message}`;
				}
				break;
			case 'PLACE':
				type = 'item';
				if (result.success) {
					text = `placed ${result.placed}`;
				} else {
					type = 'error';
					text = `failed to place: ${result.message}`;
				}
				break;
			case 'CRAFT':
				type = 'craft';
				if (result.success) {
					text = `crafted ${result.item_quantity || 1} ${result.crafted}`;
				} else {
					type = 'error';
					text = `failed to craft: ${result.message}`;
				}
				break;
			case 'HARVEST':
				type = 'harvest';
				if (result.success) {
					text = `harvested ${result.item_quantity || 1} ${result.harvested}`;
				} else {
					type = 'error';
					text = `failed to harvest: ${result.message}`;
				}
				break;
			case 'SCAN':
				type = 'info';
				if (result.success) {
					text = 'scanned the area';
				} else {
					type = 'error';
					text = `failed to scan: ${result.message}`;
				}
				break;
			case 'INTERACT':
				type = 'interact';
				if (result.success) {
					text = result.message || 'interacted with object';
				} else {
					type = 'error';
					text = `failed to interact: ${result.message}`;
				}
				break;
			case 'UPGRADE':
				type = 'upgrade';
				if (result.success) {
					text = `upgraded ${result.upgraded} to level ${result.new_level}`;
				} else {
					type = 'error';
					text = `failed to upgrade: ${result.message}`;
				}
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
export const worldObjects = derived(gameState, $s => $s.worldObjects);

// Get world objects at a specific position
export const objectsAtPosition = derived(gameState, $s => {
	const map = new Map<string, WorldObject[]>();
	for (const obj of $s.worldObjects) {
		const key = `${obj.position.x},${obj.position.y}`;
		const existing = map.get(key) || [];
		existing.push(obj);
		map.set(key, existing);
	}
	return map;
});

// Get player agent
export const playerAgent = derived(gameState, $s => {
	if (!$s.playerAgentId) return null;
	return $s.agents.find(a => a.id === $s.playerAgentId) || null;
});

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
