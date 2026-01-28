import { writable } from 'svelte/store';
import { setFullState, applyTickUpdate } from './game';
import type { WebSocketMessage } from '$lib/types';

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

export function connectToGame(gameId: string) {
	// Disconnect from previous game
	if (socket) {
		disconnect();
	}

	currentGameId = gameId;
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsUrl = `${protocol}//${window.location.host}/ws/game/${gameId}`;

	try {
		socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			console.log('WebSocket connected');
			wsState.set({ connected: true, error: null });
		};

		socket.onmessage = (event) => {
			try {
				const data: WebSocketMessage = JSON.parse(event.data);
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
						connectToGame(currentGameId);
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
	if (socket) {
		socket.close();
		socket = null;
	}
	wsState.set({ connected: false, error: null });
}

function handleMessage(data: WebSocketMessage) {
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
			console.log('Unknown message type:', data);
	}
}

export function sendPing() {
	if (socket && socket.readyState === WebSocket.OPEN) {
		socket.send(JSON.stringify({ type: 'ping' }));
	}
}
