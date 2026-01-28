<script lang="ts">
	import { gameState, ownershipCounts } from '$lib/stores/game';
	import { wsState } from '$lib/stores/ws';

	$: status = $gameState.status;
	$: tick = $gameState.tick;
	$: agents = $gameState.agents;
	$: connected = $wsState.connected;

	function getAgentScore(agentId: string): number {
		return $ownershipCounts.get(agentId) || 0;
	}
</script>

<div class="controls">
	<div class="status-bar">
		<div class="status">
			<span class="label">Status:</span>
			<span class="value {status}">{status || 'disconnected'}</span>
		</div>
		<div class="tick">
			<span class="label">Tick:</span>
			<span class="value">{tick}</span>
		</div>
		<div class="connection" class:connected>
			{connected ? 'Connected' : 'Disconnected'}
		</div>
	</div>

	<div class="scoreboard">
		<h3>Agents</h3>
		<div class="agent-list">
			{#each agents as agent}
				<div class="agent-row" class:player={agent.id === $gameState.playerAgentId}>
					<span class="name">
						{agent.name}
						{#if agent.is_adversary}
							<span class="badge">{agent.adversary_type}</span>
						{/if}
					</span>
					<span class="score">{getAgentScore(agent.id)} tiles</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.controls {
		background: #16213e;
		border-radius: 8px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.status-bar {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
	}

	.status, .tick {
		display: flex;
		gap: 8px;
	}

	.label {
		color: #a0aec0;
	}

	.value {
		font-weight: bold;
	}

	.value.waiting { color: #f6e05e; }
	.value.running { color: #48bb78; }
	.value.finished { color: #fc8181; }

	.connection {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		background: #fc8181;
		color: #1a1a2e;
	}

	.connection.connected {
		background: #48bb78;
	}

	.scoreboard h3 {
		margin-bottom: 8px;
		font-size: 14px;
		color: #a0aec0;
	}

	.agent-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.agent-row {
		display: flex;
		justify-content: space-between;
		padding: 8px;
		background: #1a1a2e;
		border-radius: 4px;
	}

	.agent-row.player {
		border-left: 3px solid #f6e05e;
	}

	.name {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.badge {
		font-size: 10px;
		padding: 2px 6px;
		background: #4a5568;
		border-radius: 4px;
	}

	.score {
		color: #a0aec0;
	}
</style>
