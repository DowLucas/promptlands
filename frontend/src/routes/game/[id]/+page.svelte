<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { connectToGame, disconnect } from '$lib/stores/ws';
	import { setPlayerAgentId, resetGame, logEntries, playerAgent } from '$lib/stores/game';
	import { playerInventory } from '$lib/stores/inventory';
	import { hudVisible } from '$lib/stores/overlay';
	import WorldMap from '$lib/components/WorldMap.svelte';
	import GameControls from '$lib/components/GameControls.svelte';
	import AgentInfo from '$lib/components/AgentInfo.svelte';
	import InventoryPanel from '$lib/components/InventoryPanel.svelte';

	$: gameId = $page.params.id ?? '';
	$: playerAgentId = $page.url.searchParams.get('agent');

	onMount(() => {
		if (playerAgentId) {
			setPlayerAgentId(playerAgentId);
		}
		if (gameId) {
			// Pass playerAgentId for fog of war calculation
			connectToGame(gameId, playerAgentId || undefined);
		}
	});

	onDestroy(() => {
		disconnect();
		resetGame();
	});
</script>

<div class="game-page">
	<header>
		<a href="/" class="back">Exit Game</a>
		<h1>Game: {gameId ? gameId.slice(0, 8) : '...'}...</h1>
	</header>

	<div class="game-layout" class:has-left-sidebar={$hudVisible && ($playerAgent || $playerInventory)}>
		{#if $hudVisible && ($playerAgent || $playerInventory)}
			<div class="left-sidebar">
				{#if $playerAgent}
					<AgentInfo agent={$playerAgent} isPlayer={true} />
				{/if}
				{#if $playerInventory}
					<InventoryPanel inventory={$playerInventory} />
				{/if}
			</div>
		{/if}
		<div class="main-area">
			<WorldMap />
		</div>
		<div class="sidebar">
			<GameControls />
			<div class="game-log">
				<h3>Game Log</h3>
				<div class="log-container">
					{#each $logEntries as entry}
						<div class="log-entry {entry.type}" class:player={entry.isPlayer}>
							<span class="tick">T{entry.tick}</span>
							<span class="agent" class:player-name={entry.isPlayer}>{entry.agentName}</span>
							<span class="text">{entry.text}</span>
						</div>
						{#if entry.reasoning}
							<div class="reasoning-line">
								<span class="reasoning-text">{entry.reasoning}</span>
							</div>
						{/if}
					{/each}
					{#if $logEntries.length === 0}
						<div class="empty">Waiting for game actions...</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.game-page {
		height: 100vh;
		display: flex;
		flex-direction: column;
		padding: 16px;
		gap: 16px;
	}

	header {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.back {
		color: #fc8181;
		text-decoration: none;
	}

	.back:hover {
		text-decoration: underline;
	}

	h1 {
		font-size: 18px;
		color: #a0aec0;
	}

	.game-layout {
		flex: 1;
		display: grid;
		grid-template-columns: 1fr 300px;
		gap: 12px;
		min-height: 0;
	}

	.game-layout.has-left-sidebar {
		grid-template-columns: 260px 1fr 300px;
	}

	.left-sidebar {
		display: flex;
		flex-direction: column;
		gap: 12px;
		overflow-y: auto;
		min-height: 0;
	}

	.main-area {
		min-height: 0;
	}

	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 12px;
		overflow-y: auto;
	}

	@media (max-width: 1100px) {
		.game-layout.has-left-sidebar {
			grid-template-columns: 220px 1fr 260px;
		}
	}

	@media (max-width: 900px) {
		.game-layout,
		.game-layout.has-left-sidebar {
			grid-template-columns: 1fr;
			grid-template-rows: 1fr auto;
		}
		.left-sidebar {
			display: none;
		}
	}

	/* Game Log Styles */
	.game-log {
		background: #0f0f1a;
		border-radius: 8px;
		padding: 12px;
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 200px;
		border: 1px solid #2d2d4a;
	}

	.game-log h3 {
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

	.reasoning-line {
		padding: 2px 8px 4px 44px;
		font-style: italic;
		color: #6b7280;
		font-size: 11px;
	}

	.reasoning-text {
		color: #8b8fa3;
	}

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
