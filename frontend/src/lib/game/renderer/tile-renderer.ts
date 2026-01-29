import { Graphics } from 'pixi.js';
import type { Tile, Agent } from '$lib/types';
import type { RenderContext } from './render-context';
import type { ViewportManager } from './viewport-manager';
import type { FogOfWarRenderer } from './fog-renderer';
import type { ViewportBounds } from './types';
import {
	BIOME_COLORS,
	TERRAIN_COLORS,
	COLORS,
	getAgentColor,
	perfLog,
	DEBUG_PERF
} from './types';

export class TileRenderer {
	private ctx: RenderContext;
	private viewport: ViewportManager;
	private fog: FogOfWarRenderer;
	private viewportMask: Graphics | null = null;

	constructor(ctx: RenderContext, viewport: ViewportManager, fog: FogOfWarRenderer) {
		this.ctx = ctx;
		this.viewport = viewport;
		this.fog = fog;
	}

	initMask() {
		this.viewportMask = new Graphics();
		this.ctx.worldContainer.addChild(this.viewportMask);
	}

	drawGrid() {
		this.ctx.gridGraphics.clear();

		if (this.ctx.worldSize >= 512) return;

		const size = this.ctx.worldSize * this.ctx.tileSize;
		this.ctx.gridGraphics.setStrokeStyle({ width: 1, color: COLORS.gridLine, alpha: 0.3 });

		for (let i = 0; i <= this.ctx.worldSize; i++) {
			this.ctx.gridGraphics.moveTo(i * this.ctx.tileSize, 0);
			this.ctx.gridGraphics.lineTo(i * this.ctx.tileSize, size);
			this.ctx.gridGraphics.moveTo(0, i * this.ctx.tileSize);
			this.ctx.gridGraphics.lineTo(size, i * this.ctx.tileSize);
		}
		this.ctx.gridGraphics.stroke();
	}

	setTileIndex(tileIndex: Map<string, Tile>) {
		const startTime = performance.now();
		this.ctx.tileIndex = tileIndex;
		perfLog(`setTileIndex (${tileIndex.size} tiles)`, startTime);
	}

	getTileColor(tile: Tile): number {
		if (tile.biome && BIOME_COLORS[tile.biome]) {
			return BIOME_COLORS[tile.biome];
		}
		switch (tile.terrain) {
			case 'forest':
				return TERRAIN_COLORS.forest;
			case 'mountain':
				return TERRAIN_COLORS.mountain;
			case 'water':
				return TERRAIN_COLORS.water;
			default:
				return TERRAIN_COLORS.plains;
		}
	}

	/**
	 * Render only tiles within the viewport (for large maps).
	 */
	renderVisibleTiles() {
		const startTime = performance.now();
		const bounds = this.viewport.getViewportBounds();
		this.viewport.setLastViewport(bounds);

		const tilesInView = (bounds.maxX - bounds.minX + 1) * (bounds.maxY - bounds.minY + 1);

		this.ctx.tileGraphics.clear();
		this.fog.clear();

		const gap = this.ctx.isLargeMap ? 0 : 1;
		const tileDrawSize = this.ctx.tileSize - gap * 2;

		for (let y = bounds.minY; y <= bounds.maxY; y++) {
			for (let x = bounds.minX; x <= bounds.maxX; x++) {
				const key = `${x},${y}`;
				const tile = this.ctx.tileIndex.get(key);
				if (!tile) continue;

				const px = tile.x * this.ctx.tileSize;
				const py = tile.y * this.ctx.tileSize;

				const isVisible = !this.ctx.fogOfWarEnabled || this.ctx.visibleTiles.has(key);
				const isExplored = this.ctx.exploredTiles.has(key);

				const terrainColor = this.getTileColor(tile);
				const terrainAlpha = this.ctx.isLargeMap ? 1.0 : 0.6;
				this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
				this.ctx.tileGraphics.fill({ color: terrainColor, alpha: terrainAlpha });

				if (isVisible) {
					if (tile.owner_id) {
						const ownerColor = this.ctx.agentColors.get(tile.owner_id) || COLORS.neutral;
						const ownerAlpha = this.ctx.isLargeMap ? 0.6 : 0.5;
						this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
						this.ctx.tileGraphics.fill({ color: ownerColor, alpha: ownerAlpha });
					}
				} else if (isExplored) {
					if (tile.owner_id) {
						const ownerColor = this.ctx.agentColors.get(tile.owner_id) || COLORS.neutral;
						const ownerAlpha = (this.ctx.isLargeMap ? 0.6 : 0.5) * 0.5;
						this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
						this.ctx.tileGraphics.fill({ color: ownerColor, alpha: ownerAlpha });
					}
				}

				// Delegate fog rendering
				this.fog.renderFogForTile(key, px, py, isVisible, isExplored);
			}
		}

		this.updateViewportMask(bounds);
		perfLog(`renderVisibleTiles.total (viewport: ${tilesInView})`, startTime);
	}

	/**
	 * Full tile render for small maps.
	 */
	updateTiles(tiles: Tile[], agents: Agent[]) {
		const startTime = performance.now();
		if (DEBUG_PERF)
			console.log(
				`[Renderer] updateTiles called with ${tiles.length} tiles, ${agents.length} agents`
			);

		this.ctx.tiles = tiles;

		// Only rebuild index for small maps
		if (!this.ctx.isLargeMap) {
			const indexStart = performance.now();
			this.ctx.tileIndex.clear();
			for (const tile of tiles) {
				this.ctx.tileIndex.set(`${tile.x},${tile.y}`, tile);
			}
			perfLog(`updateTiles.buildIndex (${tiles.length} tiles)`, indexStart);
		}

		// Build agent color map
		const colorStart = performance.now();
		this.ctx.agentColors.clear();
		for (const agent of agents) {
			const isPlayer = agent.id === this.ctx.playerAgentId;
			this.ctx.agentColors.set(agent.id, getAgentColor(agent.id, isPlayer));
		}
		perfLog(`updateTiles.buildColors`, colorStart);

		if (this.ctx.isLargeMap) {
			this.renderVisibleTiles();
			perfLog(`updateTiles.total (large map)`, startTime);
			return;
		}

		// Small map: render all tiles
		this.ctx.tileGraphics.clear();
		this.fog.clear();
		const gap = 1;
		const tileDrawSize = this.ctx.tileSize - gap * 2;

		for (const tile of tiles) {
			const px = tile.x * this.ctx.tileSize;
			const py = tile.y * this.ctx.tileSize;
			const key = `${tile.x},${tile.y}`;

			const isVisible = !this.ctx.fogOfWarEnabled || this.ctx.visibleTiles.has(key);
			const isExplored = this.ctx.exploredTiles.has(key);

			const terrainColor = this.getTileColor(tile);
			this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
			this.ctx.tileGraphics.fill({ color: terrainColor, alpha: 0.6 });

			if (isVisible) {
				if (tile.owner_id) {
					const ownerColor = this.ctx.agentColors.get(tile.owner_id) || COLORS.neutral;
					this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
					this.ctx.tileGraphics.fill({ color: ownerColor, alpha: 0.5 });
				}
			} else if (isExplored) {
				if (tile.owner_id) {
					const ownerColor = this.ctx.agentColors.get(tile.owner_id) || COLORS.neutral;
					this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
					this.ctx.tileGraphics.fill({ color: ownerColor, alpha: 0.25 });
				}
			}

			this.fog.renderFogForTile(key, px, py, isVisible, isExplored);
		}

		// No mask needed for small maps
		this.ctx.objectContainer.mask = null;
		this.ctx.agentContainer.mask = null;

		perfLog(`updateTiles.total (small map)`, startTime);
	}

	updateTileOwnership(changes: Array<{ x: number; y: number; owner_id: string | null }>) {
		for (const change of changes) {
			const key = `${change.x},${change.y}`;
			const tile = this.ctx.tileIndex.get(key);
			if (tile) {
				tile.owner_id = change.owner_id;
			}
		}
		if (this.ctx.isLargeMap) {
			this.renderVisibleTiles();
		} else {
			this.updateTiles(this.ctx.tiles, []);
		}
	}

	refreshVisibleTiles() {
		if (this.ctx.isLargeMap) {
			this.renderVisibleTiles();
		}
	}

	private updateViewportMask(bounds: ViewportBounds) {
		if (!this.viewportMask) return;

		this.viewportMask.clear();
		const maskX = bounds.minX * this.ctx.tileSize;
		const maskY = bounds.minY * this.ctx.tileSize;
		const maskW = (bounds.maxX - bounds.minX + 1) * this.ctx.tileSize;
		const maskH = (bounds.maxY - bounds.minY + 1) * this.ctx.tileSize;

		this.viewportMask.rect(maskX, maskY, maskW, maskH);
		this.viewportMask.fill({ color: 0xffffff });

		this.ctx.objectContainer.mask = this.viewportMask;
		this.ctx.agentContainer.mask = this.viewportMask;
	}
}
