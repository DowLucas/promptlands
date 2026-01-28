<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { GameRenderer } from '$lib/game/renderer';
	import { gameState } from '$lib/stores/game';

	let canvas: HTMLCanvasElement;
	let container: HTMLDivElement;
	let renderer: GameRenderer | null = null;

	onMount(() => {
		let resizeObserver: ResizeObserver;

		(async () => {
			renderer = new GameRenderer();
			await renderer.init(canvas);

			// Set up resize observer
			resizeObserver = new ResizeObserver((entries) => {
				for (const entry of entries) {
					const { width, height } = entry.contentRect;
					renderer?.resize(width, height);
				}
			});
			resizeObserver.observe(container);
		})();

		return () => {
			resizeObserver?.disconnect();
		};
	});

	onDestroy(() => {
		renderer?.destroy();
	});

	// React to game state changes
	$: if (renderer && $gameState.world) {
		renderer.setWorldSize($gameState.world.size);
		renderer.setPlayerAgentId($gameState.playerAgentId);
		renderer.updateTiles($gameState.world.tiles, $gameState.agents);
		renderer.updateAgents($gameState.agents);
	}
</script>

<div class="world-container" bind:this={container}>
	<canvas bind:this={canvas}></canvas>
</div>

<style>
	.world-container {
		width: 100%;
		height: 100%;
		min-height: 400px;
		background: #1a1a2e;
		border-radius: 8px;
		overflow: hidden;
	}

	canvas {
		width: 100%;
		height: 100%;
		display: block;
	}
</style>
