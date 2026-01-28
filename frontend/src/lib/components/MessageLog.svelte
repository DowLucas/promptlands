<script lang="ts">
	import { messages, gameState } from '$lib/stores/game';

	$: agents = $gameState.agents;

	function getAgentName(agentId: string): string {
		const agent = agents.find(a => a.id === agentId);
		return agent?.name || 'Unknown';
	}

	// Auto-scroll to bottom on new messages
	let logContainer: HTMLDivElement;
	$: if ($messages.length && logContainer) {
		setTimeout(() => {
			logContainer.scrollTop = logContainer.scrollHeight;
		}, 0);
	}
</script>

<div class="message-log">
	<h3>Messages</h3>
	<div class="log-container" bind:this={logContainer}>
		{#each $messages as msg}
			<div class="message" class:broadcast={!msg.to_agent_id}>
				<span class="tick">T{msg.tick}</span>
				<span class="from">{getAgentName(msg.from_agent_id)}:</span>
				{#if !msg.to_agent_id}
					<span class="broadcast-tag">[broadcast]</span>
				{/if}
				<span class="content">{msg.content}</span>
			</div>
		{/each}
		{#if $messages.length === 0}
			<div class="empty">No messages yet</div>
		{/if}
	</div>
</div>

<style>
	.message-log {
		background: #16213e;
		border-radius: 8px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		height: 200px;
	}

	h3 {
		margin-bottom: 8px;
		font-size: 14px;
		color: #a0aec0;
	}

	.log-container {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.message {
		padding: 4px 8px;
		background: #1a1a2e;
		border-radius: 4px;
		font-size: 12px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}

	.message.broadcast {
		background: #2d3748;
	}

	.tick {
		color: #4a5568;
		font-family: monospace;
	}

	.from {
		color: #48bb78;
		font-weight: bold;
	}

	.broadcast-tag {
		color: #f6e05e;
		font-size: 10px;
	}

	.content {
		color: #e2e8f0;
	}

	.empty {
		color: #4a5568;
		font-style: italic;
		text-align: center;
		padding: 16px;
	}
</style>
