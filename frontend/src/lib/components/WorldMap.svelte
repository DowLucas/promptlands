<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { GameRenderer } from '$lib/game/renderer';
	import { gameState, addExploredTiles, playerAgent, exploredTiles, visibleTiles, centerOnRequest, objectsAtPosition, agents, latestResults } from '$lib/stores/game';
	import { wsState } from '$lib/stores/ws';
	import { get } from 'svelte/store';
	import { inspectTile, clearInspection, inspectedTile, hudVisible } from '$lib/stores/overlay';
	import { playerInventory } from '$lib/stores/inventory';
	import TileInspector from './TileInspector.svelte';
	import ResourceOverlay from './ResourceOverlay.svelte';

	const DEBUG_PERF = import.meta.env.DEV;

	let canvas: HTMLCanvasElement;
	let container: HTMLDivElement;
	let renderer: GameRenderer | null = null;
	let rendererReady = false;
	let lastWorldSize = 0;
	let isInitialized = false;
	let isLoading = true;
	let loadingMessage = 'Connecting...';

	onMount(() => {
		let resizeObserver: ResizeObserver;

		(async () => {
			renderer = new GameRenderer();
			await renderer.init(canvas);
			rendererReady = true;

			// Wire tile hover for inspection tooltip (only if tile has content)
			renderer.setOnTileHover((tileX, tileY, sx, sy) => {
				const key = `${tileX},${tileY}`;
				const tile = get(gameState).tileIndex.get(key);
				const objs = get(objectsAtPosition).get(key);
				const tileAgents = get(agents).filter(
					a => a.position.x === tileX && a.position.y === tileY && !a.is_dead
				);
				const hasContent = (tileAgents.length > 0) || (objs && objs.length > 0) || !!tile?.owner_id;

				if (!hasContent) {
					clearInspection();
					return;
				}

				const current = get(inspectedTile);
				if (!current || current.position.x !== tileX || current.position.y !== tileY) {
					inspectTile({ x: tileX, y: tileY }, sx, sy);
				}
			});

			renderer.setOnTileHoverEnd(() => {
				clearInspection();
			});

			// Close inspector on camera move (drag)
			renderer.setOnCameraMove(() => {
				clearInspection();
			});

			// Set up resize observer
			resizeObserver = new ResizeObserver((entries) => {
				for (const entry of entries) {
					const { width, height } = entry.contentRect;
					renderer?.resize(width, height);
				}
			});
			resizeObserver.observe(container);

			// Trigger initial render if state is already available
			if ($gameState.world) {
				initializeRenderer();
			}
		})();

		return () => {
			resizeObserver?.disconnect();
		};
	});

	onDestroy(() => {
		renderer?.destroy();
	});

	// Full initialization - only on first load or world size change
	function initializeRenderer() {
		if (!renderer || !$gameState.world) return;

		loadingMessage = `Rendering ${$gameState.world.tiles.length.toLocaleString()} tiles...`;

		const startTime = DEBUG_PERF ? performance.now() : 0;
		if (DEBUG_PERF) console.log(`[WorldMap] Initializing renderer with ${$gameState.world.tiles.length} tiles`);

		renderer.setWorldSize($gameState.world.size);
		renderer.setPlayerAgentId($gameState.playerAgentId);

		// Pass the store's tile index directly to avoid rebuilding
		renderer.setTileIndex($gameState.tileIndex);

		// Set up explored tiles from store
		renderer.setExploredTiles($exploredTiles);

		// Calculate initial visibility and update explored tiles
		updateFogOfWar();

		renderer.updateTiles($gameState.world.tiles, $gameState.agents);
		renderer.updateAgents($gameState.agents);
		renderer.updateWorldObjects($gameState.worldObjects);

		// Center camera on player's position
		if ($playerAgent) {
			renderer.centerOn($playerAgent.position.x, $playerAgent.position.y);
		}

		lastWorldSize = $gameState.world.size;
		isInitialized = true;
		isLoading = false;

		if (DEBUG_PERF) console.log(`[WorldMap] Initialize complete: ${(performance.now() - startTime).toFixed(2)}ms`);
	}

	// Update fog of war using server-provided visibility
	function updateFogOfWar() {
		if (!renderer) return;

		// Use server-provided visible tiles (already calculated by backend)
		const visible = $visibleTiles;

		if (DEBUG_PERF) {
			console.log(`[WorldMap] updateFogOfWar: visibleTiles.size = ${visible.size}, exploredTiles.size = ${$exploredTiles.size}`);
		}

		// Update renderer with visibility (also updates explored tiles)
		const newlyExplored = renderer.updateVisibility(visible, $exploredTiles);

		// Persist newly explored tiles to the store
		if (newlyExplored.size > 0) {
			addExploredTiles(newlyExplored);
		}
	}

	// Lightweight update for tick changes
	function updateForTick() {
		if (!renderer || !$gameState.world) return;

		const startTime = DEBUG_PERF ? performance.now() : 0;

		// Update fog of war visibility (player may have moved or beacons changed)
		updateFogOfWar();

		// Update agents (they move) - filtered by visibility in renderer
		renderer.updateAgents($gameState.agents);

		// Update reasoning thought bubbles
		if ($latestResults.length > 0) {
			renderer.updateReasonings($latestResults, $gameState.tick);
		}

		// Refresh visible tiles (ownership changes, no index rebuild)
		renderer.refreshVisibleTiles();

		// Update world objects - filtered by visibility in renderer
		renderer.updateWorldObjects($gameState.worldObjects);

		if (DEBUG_PERF) {
			const elapsed = performance.now() - startTime;
			if (elapsed > 5) {
				console.log(`[WorldMap] Tick update: ${elapsed.toFixed(2)}ms`);
			}
		}
	}

	// Track connection state for loading message
	$: if (!$wsState.connected && !isInitialized) {
		loadingMessage = 'Connecting...';
	} else if ($wsState.connected && !$gameState.world) {
		loadingMessage = 'Loading world data...';
	}

	// React to game state changes (including visibility updates)
	// Note: we include $visibleTiles in the condition to ensure reactivity when visibility changes
	$: if (rendererReady && renderer && $gameState.world && $visibleTiles !== undefined) {
		if (DEBUG_PERF) {
			console.log(`[WorldMap] Reactive update: tick=${$gameState.tick}, visibleTiles.size=${$visibleTiles.size}, isInitialized=${isInitialized}`);
		}
		// Check if we need full initialization or just a tick update
		if (!isInitialized || $gameState.world.size !== lastWorldSize) {
			initializeRenderer();
		} else {
			updateForTick();
		}
	}

	// React to center-on requests (e.g., clicking an agent in the sidebar)
	$: if (renderer && $centerOnRequest) {
		renderer.centerOn($centerOnRequest.x, $centerOnRequest.y);
		centerOnRequest.set(null); // Clear after handling
	}
</script>

<div class="world-container" bind:this={container}>
	<canvas bind:this={canvas}></canvas>
	{#if isLoading}
		<div class="loading-overlay">
			<div class="loading-content">
				<div class="spinner"></div>
				<p>{loadingMessage}</p>
			</div>
		</div>
	{/if}

	<!-- Resource Overlay HUD -->
	{#if $hudVisible && $playerAgent}
		<ResourceOverlay />
	{/if}

	<!-- Tile Inspector tooltip -->
	{#if $inspectedTile}
		<TileInspector
			position={$inspectedTile.position}
			screenX={$inspectedTile.screenX}
			screenY={$inspectedTile.screenY}
		/>
	{/if}


</div>

<style>
	.world-container {
		width: 100%;
		height: 100%;
		min-height: 400px;
		background: #1a1a2e;
		border-radius: 8px;
		overflow: hidden;
		position: relative;
	}

	canvas {
		width: 100%;
		height: 100%;
		display: block;
	}

	.loading-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(26, 26, 46, 0.95);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10;
	}

	.loading-content {
		text-align: center;
		color: #e2e8f0;
	}

	.loading-content p {
		margin-top: 1rem;
		font-size: 0.9rem;
		color: #a0aec0;
	}

	.spinner {
		width: 48px;
		height: 48px;
		border: 3px solid #4a5568;
		border-top-color: #6366f1;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

</style>
