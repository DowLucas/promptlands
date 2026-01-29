import { writable } from 'svelte/store';
import { setFullState, applyTickUpdate } from './game';
import type { WebSocketMessage } from '$lib/types';

// Debug timing utility - enabled in dev mode
const DEBUG_PERF = import.meta.env.DEV;

interface WSState {
	connected: boolean;
	error: string | null;
}

export const wsState = writable<WSState>({
	connected: false,
	error: null
});

let socket: WebSocket | null = null;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let currentGameId: string | null = null;
let currentPlayerAgentId: string | null = null;

export function connectToGame(gameId: string, playerAgentId?: string) {
	// Disconnect from previous game
	if (socket) {
		disconnect();
	}

	currentGameId = gameId;
	currentPlayerAgentId = playerAgentId || null;
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	let wsUrl = `${protocol}//${window.location.host}/ws/game/${gameId}`;
	// Add player_agent_id for fog of war calculation
	if (playerAgentId) {
		wsUrl += `?player_agent_id=${playerAgentId}`;
	}
	console.log(`[WS] Connecting to: ${wsUrl} (playerAgentId: ${playerAgentId || 'none'})`);

	try {
		socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			console.log('WebSocket connected');
			wsState.set({ connected: true, error: null });
		};

		socket.onmessage = (event) => {
			try {
				const parseStart = performance.now();
				const data: WebSocketMessage = JSON.parse(event.data);
				if (DEBUG_PERF) {
					const parseTime = performance.now() - parseStart;
					const msgSize = event.data.length;
					if (parseTime > 1 || msgSize > 100000) {
						console.log(`[WS] JSON.parse: ${parseTime.toFixed(2)}ms, size: ${(msgSize / 1024).toFixed(1)}KB, type: ${data.type}`);
					}
				}
				handleMessage(data);
			} catch (e) {
				console.error('Failed to parse WebSocket message:', e);
			}
		};

		socket.onerror = (error) => {
			console.error('WebSocket error:', error);
			wsState.update(s => ({ ...s, error: 'Connection error' }));
		};

		socket.onclose = (event) => {
			console.log('WebSocket closed:', event.code, event.reason);
			wsState.set({ connected: false, error: null });
			socket = null;

			// Attempt to reconnect after 3 seconds
			if (currentGameId) {
				reconnectTimeout = setTimeout(() => {
					if (currentGameId) {
						console.log('Attempting to reconnect...');
						connectToGame(currentGameId, currentPlayerAgentId || undefined);
					}
				}, 3000);
			}
		};
	} catch (e) {
		console.error('Failed to create WebSocket:', e);
		wsState.set({ connected: false, error: 'Failed to connect' });
	}
}

export function disconnect() {
	if (reconnectTimeout) {
		clearTimeout(reconnectTimeout);
		reconnectTimeout = null;
	}
	currentGameId = null;
	currentPlayerAgentId = null;
	if (socket) {
		socket.close();
		socket = null;
	}
	wsState.set({ connected: false, error: null });
}

function handleMessage(data: WebSocketMessage) {
	const startTime = DEBUG_PERF ? performance.now() : 0;
	switch (data.type) {
		case 'full_state':
			setFullState(data);
			break;
		case 'tick':
			applyTickUpdate(data);
			break;
		case 'game_over':
			console.log('Game over! Winner:', data.winner);
			// Could emit an event or update store
			break;
		case 'pong':
			// Heartbeat response
			break;
		default:
			if (DEBUG_PERF) console.log('Unknown message type:', data);
	}
	if (DEBUG_PERF) {
		const elapsed = performance.now() - startTime;
		if (elapsed > 5) {
			console.log(`[WS] handleMessage(${data.type}): ${elapsed.toFixed(2)}ms`);
		}
	}
}

export function sendPing() {
	if (socket && socket.readyState === WebSocket.OPEN) {
		socket.send(JSON.stringify({ type: 'ping' }));
	}
}
