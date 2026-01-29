import { writable, derived, type Writable } from 'svelte/store';
import type { Agent, FullGameState, GameMessage, Tile, TickUpdate, WorldSnapshot, ActionResult, WorldObject } from '$lib/types';
import { formatActionResult } from '$lib/game/log-formatter';
import { setPlayerInventory } from '$lib/stores/inventory';

// Debug timing utility - enabled in dev mode
const DEBUG_PERF = import.meta.env.DEV;
function perfLog(label: string, startTime: number) {
	if (DEBUG_PERF) {
		const elapsed = performance.now() - startTime;
		if (elapsed > 1) { // Only log if > 1ms
			console.log(`[GameStore] ${label}: ${elapsed.toFixed(2)}ms`);
		}
	}
}

// Log entries store
export interface LogEntry {
	tick: number;
	type: 'move' | 'claim' | 'message' | 'wait' | 'info' | 'error' | 'combat' | 'item' | 'craft' | 'harvest' | 'interact' | 'upgrade' | 'death' | 'respawn';
	agentId: string;
	agentName: string;
	text: string;
	reasoning?: string;
	isPlayer: boolean;
}

export const logEntries = writable<LogEntry[]>([]);

// Latest action results for rendering thought bubbles
export const latestResults = writable<ActionResult[]>([]);

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
	tileIndex: Map<string, Tile>; // O(1) tile lookups by "x,y" key
	exploredTiles: Set<string>; // "x,y" keys of tiles the player has seen
	visibleTiles: Set<string>; // Server-provided currently visible tiles
	// MULTIPLAYER: Will need Map<playerId, Set<string>> for per-player exploration
}

const initialState: GameState = {
	gameId: null,
	status: null,
	tick: 0,
	world: null,
	agents: [],
	messages: [],
	playerAgentId: null,
	worldObjects: [],
	tileIndex: new Map(),
	exploredTiles: new Set(),
	visibleTiles: new Set()
};

export const gameState: Writable<GameState> = writable(initialState);

// Build tile index from tiles array
function buildTileIndex(tiles: Tile[]): Map<string, Tile> {
	const index = new Map<string, Tile>();
	for (const tile of tiles) {
		index.set(`${tile.x},${tile.y}`, tile);
	}
	return index;
}

// Actions
export function setFullState(state: FullGameState) {
	const startTime = performance.now();
	const tileCount = state.world?.tiles?.length ?? 0;
	if (DEBUG_PERF) console.log(`[GameStore] setFullState called with ${tileCount} tiles`);

	const indexStart = performance.now();
	const tileIndex = state.world ? buildTileIndex(state.world.tiles) : new Map<string, Tile>();
	perfLog(`setFullState.buildTileIndex (${tileCount} tiles)`, indexStart);

	// Use server-provided visible tiles if available
	const visibleTiles = new Set(state.visible_tiles || []);
	if (DEBUG_PERF) {
		console.log(`[GameStore] setFullState: visible_tiles from server = ${state.visible_tiles?.length ?? 0} tiles`, state.visible_tiles?.slice(0, 5));
	}

	const updateStart = performance.now();
	gameState.update(s => ({
		...s,
		gameId: state.game_id,
		status: state.status,
		tick: state.tick,
		world: state.world,
		agents: state.agents,
		messages: state.messages,
		worldObjects: state.world_objects || [],
		tileIndex,
		visibleTiles
	}));
	perfLog(`setFullState.storeUpdate`, updateStart);

	// Update player inventory if provided
	if (state.player_inventory) {
		setPlayerInventory(state.player_inventory);
	}

	perfLog(`setFullState.total`, startTime);
}

export function applyTickUpdate(update: TickUpdate) {
	const startTime = performance.now();
	let currentState: GameState;
	gameState.subscribe(s => currentState = s)();

	gameState.update(s => {
		if (!s.world) return s;

		// Update tiles directly in the index (O(1) per change)
		const tileStart = performance.now();
		const tileIndex = s.tileIndex;
		for (const change of update.changes.tiles) {
			const key = `${change.x},${change.y}`;
			const existing = tileIndex.get(key);
			if (existing) {
				existing.owner_id = change.owner_id;
			}
		}
		perfLog(`applyTickUpdate.tiles (${update.changes.tiles.length} changes)`, tileStart);

		// Update agents
		const newAgents = update.changes.agents;

		// Append new messages
		const newMessages = [...s.messages, ...update.changes.messages];

		// Update world objects
		const objectStart = performance.now();
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
		perfLog(`applyTickUpdate.objects`, objectStart);

		// Use server-provided visible tiles if available
		const visibleTiles = update.changes.visible_tiles
			? new Set(update.changes.visible_tiles)
			: s.visibleTiles;
		if (DEBUG_PERF && update.changes.visible_tiles) {
			console.log(`[GameStore] applyTickUpdate: visible_tiles from server = ${update.changes.visible_tiles.length} tiles`);
		}

		return {
			...s,
			tick: update.tick,
			world: s.world, // Keep same world reference, tiles updated in place via index
			agents: newAgents,
			messages: newMessages,
			worldObjects: newWorldObjects,
			tileIndex, // Keep same index reference
			visibleTiles
		};
	});
	perfLog(`applyTickUpdate.total (tick ${update.tick})`, startTime);

	// Update player inventory if provided
	if (update.changes.player_inventory) {
		setPlayerInventory(update.changes.player_inventory);
	}

	// Store latest results for thought bubble rendering
	latestResults.set(update.changes.results);

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
		const formatted = formatActionResult(result, agentName, isPlayer(result.agent_id));
		if (!formatted) continue; // MESSAGE returns null — handled below

		newEntries.push({
			tick,
			type: formatted.type,
			agentId: result.agent_id,
			agentName,
			text: formatted.text,
			reasoning: result.reasoning,
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
	gameState.set({...initialState, tileIndex: new Map(), exploredTiles: new Set(), visibleTiles: new Set()});
	logEntries.set([]);
}

// Update explored tiles (called from renderer when visibility changes)
export function addExploredTiles(tiles: Set<string>) {
	gameState.update(s => {
		// Merge new tiles into existing explored set
		const newExplored = new Set(s.exploredTiles);
		for (const key of tiles) {
			newExplored.add(key);
		}
		return { ...s, exploredTiles: newExplored };
	});
}

// Derived stores
export const currentTick = derived(gameState, $s => $s.tick);
export const gameStatus = derived(gameState, $s => $s.status);
export const agents = derived(gameState, $s => $s.agents);
export const messages = derived(gameState, $s => $s.messages);
export const worldObjects = derived(gameState, $s => $s.worldObjects);
export const tileIndex = derived(gameState, $s => $s.tileIndex);

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

// Get tile ownership map for rendering (uses the existing tile index for efficiency)
export const tileOwnership = derived(gameState, $s => {
	// For large maps, return the tile index directly since it already has O(1) access
	// For small maps or when we need just ownership, build a simple map
	const map = new Map<string, string | null>();
	for (const [key, tile] of $s.tileIndex) {
		map.set(key, tile.owner_id);
	}
	return map;
});

// Get ownership counts
export const ownershipCounts = derived(gameState, $s => {
	const counts = new Map<string, number>();
	for (const tile of $s.tileIndex.values()) {
		if (tile.owner_id) {
			counts.set(tile.owner_id, (counts.get(tile.owner_id) || 0) + 1);
		}
	}
	return counts;
});

// Get explored tiles for fog of war
export const exploredTiles = derived(gameState, $s => $s.exploredTiles);

// Get server-provided visible tiles for fog of war
export const visibleTiles = derived(gameState, $s => $s.visibleTiles);

// Center-on-position request (for UI actions like clicking an agent in sidebar)
export const centerOnRequest = writable<{ x: number; y: number } | null>(null);

export function centerOnAgent(agentId: string) {
	let state: GameState;
	gameState.subscribe(s => state = s)();
	const agent = state!.agents.find(a => a.id === agentId);
	if (agent) {
		centerOnRequest.set({ x: agent.position.x, y: agent.position.y });
	}
}
