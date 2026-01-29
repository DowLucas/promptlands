import type { LogEntry } from '$lib/stores/game';
import type { ActionResult } from '$lib/types';

type LogFormatResult = { type: LogEntry['type']; text: string } | null;

type LogFormatHandler = (
	result: ActionResult,
	agentName: string,
	isPlayer: boolean
) => LogFormatResult;

const handlers: Record<string, LogFormatHandler> = {
	MOVE: (result) => {
		if (result.success && result.new_pos) {
			return { type: 'move', text: `moved to (${result.new_pos.x}, ${result.new_pos.y})` };
		}
		return { type: 'error', text: `failed to move: ${result.message}` };
	},
	CLAIM: (result) => {
		if (result.success && result.claimed_at) {
			return {
				type: 'claim',
				text: `claimed tile at (${result.claimed_at.x}, ${result.claimed_at.y})`
			};
		}
		return { type: 'error', text: `failed to claim: ${result.message}` };
	},
	WAIT: () => ({ type: 'wait', text: 'is waiting...' }),
	FIGHT: (result) => {
		if (result.success) {
			return { type: 'combat', text: result.message || `dealt ${result.damage_dealt} damage` };
		}
		return { type: 'error', text: `failed to attack: ${result.message}` };
	},
	PICKUP: (result) => {
		if (result.success) {
			return {
				type: 'item',
				text: `picked up ${result.item_quantity || 1} ${result.item_id}`
			};
		}
		return { type: 'error', text: `failed to pick up: ${result.message}` };
	},
	USE: (result) => {
		if (result.success) {
			if (result.placed) {
				return { type: 'item', text: `placed ${result.placed}` };
			}
			return { type: 'item', text: `used ${result.item_id}: ${result.message}` };
		}
		return { type: 'error', text: `failed to use: ${result.message}` };
	},
	HARVEST: (result) => {
		if (result.success) {
			return {
				type: 'harvest',
				text: `harvested ${result.item_quantity || 1} ${result.harvested}`
			};
		}
		return { type: 'error', text: `failed to harvest: ${result.message}` };
	},
	INTERACT: (result) => {
		if (result.success) {
			return { type: 'interact', text: result.message || 'interacted with object' };
		}
		return { type: 'error', text: `failed to interact: ${result.message}` };
	},
	UPGRADE: (result) => {
		if (result.success) {
			return {
				type: 'upgrade',
				text: `upgraded ${result.upgraded} to level ${result.new_level}`
			};
		}
		return { type: 'error', text: `failed to upgrade: ${result.message}` };
	},
	BUY: (result) => {
		if (result.success) {
			return { type: 'item', text: `bought ${result.item_id}` };
		}
		return { type: 'error', text: `failed to buy: ${result.message}` };
	},
	MESSAGE: () => null // Handled separately
};

export function formatActionResult(result: ActionResult, agentName: string, isPlayer: boolean): LogFormatResult {
	const handler = handlers[result.action];
	if (handler) {
		return handler(result, agentName, isPlayer);
	}
	return { type: 'info', text: result.message || 'unknown action' };
}
