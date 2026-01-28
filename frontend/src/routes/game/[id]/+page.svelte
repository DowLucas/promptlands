<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { connectToGame, disconnect } from '$lib/stores/ws';
	import { setPlayerAgentId, resetGame, logEntries } from '$lib/stores/game';
	import WorldMap from '$lib/components/WorldMap.svelte';
	import GameControls from '$lib/components/GameControls.svelte';

	$: gameId = $page.params.id ?? '';
	$: playerAgentId = $page.url.searchParams.get('agent');

	onMount(() => {
		if (playerAgentId) {
			setPlayerAgentId(playerAgentId);
		}
		if (gameId) {
			connectToGame(gameId);
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

	<div class="game-layout">
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
		grid-template-columns: 1fr 320px;
		gap: 16px;
		min-height: 0;
	}

	.main-area {
		min-height: 0;
	}

	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 16px;
		overflow-y: auto;
	}

	@media (max-width: 900px) {
		.game-layout {
			grid-template-columns: 1fr;
			grid-template-rows: 1fr auto;
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
