<script lang="ts">
	import { gameState } from '$lib/stores/game';
	import { writable } from 'svelte/store';
	import type { ActionResult, GameMessage } from '$lib/types';

	interface LogEntry {
		tick: number;
		type: 'move' | 'claim' | 'message' | 'wait' | 'info' | 'error';
		agentId: string;
		agentName: string;
		text: string;
		isPlayer: boolean;
	}

	// Store for log entries
	export const logEntries = writable<LogEntry[]>([]);

	$: agents = $gameState.agents;
	$: playerAgentId = $gameState.playerAgentId;

	function getAgentName(agentId: string): string {
		const agent = agents.find(a => a.id === agentId);
		return agent?.name || 'Unknown';
	}

	function isPlayerAgent(agentId: string): boolean {
		return agentId === playerAgentId;
	}

	// Add entries from action results
	export function addActionResults(tick: number, results: ActionResult[]) {
		const newEntries: LogEntry[] = [];

		for (const result of results) {
			const agentName = getAgentName(result.agent_id);
			const isPlayer = isPlayerAgent(result.agent_id);
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
					// Messages handled separately
					continue;
				default:
					text = result.message || 'unknown action';
			}

			newEntries.push({ tick, type, agentId: result.agent_id, agentName, text, isPlayer });
		}

		logEntries.update(entries => [...entries, ...newEntries].slice(-100));
	}

	// Add message entries
	export function addMessages(messages: GameMessage[]) {
		const newEntries: LogEntry[] = messages.map(msg => ({
			tick: msg.tick,
			type: 'message' as const,
			agentId: msg.from_agent_id,
			agentName: getAgentName(msg.from_agent_id),
			text: msg.to_agent_id ? `says: "${msg.content}"` : `broadcasts: "${msg.content}"`,
			isPlayer: isPlayerAgent(msg.from_agent_id)
		}));

		logEntries.update(entries => [...entries, ...newEntries].slice(-100));
	}

	// Auto-scroll
	let logContainer: HTMLDivElement;
	$: if ($logEntries.length && logContainer) {
		setTimeout(() => {
			logContainer.scrollTop = logContainer.scrollHeight;
		}, 0);
	}
</script>

<div class="game-log">
	<h3>Game Log</h3>
	<div class="log-container" bind:this={logContainer}>
		{#each $logEntries as entry}
			<div class="log-entry {entry.type}" class:player={entry.isPlayer}>
				<span class="tick">T{entry.tick}</span>
				<span class="agent" class:player-name={entry.isPlayer}>{entry.agentName}</span>
				<span class="text">{entry.text}</span>
			</div>
		{/each}
		{#if $logEntries.length === 0}
			<div class="empty">Waiting for game actions...</div>
		{/if}
	</div>
</div>

<style>
	.game-log {
		background: #0f0f1a;
		border-radius: 8px;
		padding: 12px;
		display: flex;
		flex-direction: column;
		height: 280px;
		border: 1px solid #2d2d4a;
	}

	h3 {
		margin-bottom: 8px;
		font-size: 13px;
		color: #888;
		text-transform: uppercase;
		letter-spacing: 1px;
	}

	.log-container {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		font-size: 12px;
	}

	.log-entry {
		padding: 4px 8px;
		border-radius: 4px;
		display: flex;
		gap: 8px;
		align-items: baseline;
		background: #1a1a2e;
	}

	.log-entry.player {
		background: #1a2a1a;
		border-left: 2px solid #ffd700;
	}

	.tick {
		color: #555;
		font-size: 10px;
		min-width: 28px;
	}

	.agent {
		font-weight: bold;
		min-width: 80px;
	}

	.player-name {
		color: #ffd700 !important;
	}

	/* Action type colors */
	.log-entry.move .agent { color: #60a5fa; }
	.log-entry.move .text { color: #93c5fd; }

	.log-entry.claim .agent { color: #34d399; }
	.log-entry.claim .text { color: #6ee7b7; }

	.log-entry.message .agent { color: #a78bfa; }
	.log-entry.message .text { color: #c4b5fd; font-style: italic; }

	.log-entry.wait .agent { color: #9ca3af; }
	.log-entry.wait .text { color: #6b7280; }

	.log-entry.error .agent { color: #f87171; }
	.log-entry.error .text { color: #fca5a5; }

	.log-entry.info .text { color: #9ca3af; }

	.text {
		flex: 1;
	}

	.empty {
		color: #4a4a6a;
		font-style: italic;
		text-align: center;
		padding: 24px;
	}
</style>
