import { Application, Container, Graphics, Sprite } from 'pixi.js';
import type { Agent, Tile, WorldObject, ActionResult } from '$lib/types';
import { SpriteAssetManager } from './sprites/index';

import { RenderContext } from './renderer/render-context';
import { ViewportManager } from './renderer/viewport-manager';
import { InputHandler } from './renderer/input-handler';
import { FogOfWarRenderer } from './renderer/fog-renderer';
import { MinimapRenderer } from './renderer/minimap-renderer';
import { TileRenderer } from './renderer/tile-renderer';
import { AgentRenderer } from './renderer/agent-renderer';
import { WorldObjectRenderer } from './renderer/object-renderer';
import {
	COLORS,
	getAgentColor,
	perfLog,
	DEBUG_PERF
} from './renderer/types';

// Re-export types for external consumers
export type { ViewportBounds, AgentSprite } from './renderer/types';

// Sprite pool for recycling Sprite objects during viewport panning
class SpritePool {
	private pool: Sprite[] = [];
	private active: Map<string, Sprite> = new Map();

	acquire(key: string): Sprite {
		let sprite = this.active.get(key);
		if (sprite) return sprite;

		sprite = this.pool.pop() ?? new Sprite();
		sprite.visible = true;
		this.active.set(key, sprite);
		return sprite;
	}

	releaseAll() {
		for (const [, sprite] of this.active) {
			sprite.visible = false;
			this.pool.push(sprite);
		}
		this.active.clear();
	}

	releaseExcept(activeKeys: Set<string>, container: Container) {
		for (const [key, sprite] of this.active) {
			if (!activeKeys.has(key)) {
				sprite.visible = false;
				container.removeChild(sprite);
				this.pool.push(sprite);
				this.active.delete(key);
			}
		}
	}

	destroy() {
		for (const sprite of this.active.values()) {
			sprite.destroy();
		}
		for (const sprite of this.pool) {
			sprite.destroy();
		}
		this.active.clear();
		this.pool.length = 0;
	}
}

export class GameRenderer {
	private ctx: RenderContext;
	private viewport: ViewportManager;
	private input: InputHandler;
	private fog: FogOfWarRenderer;
	private tiles_: TileRenderer;
	private agents_: AgentRenderer;
	private objects_: WorldObjectRenderer;
	private minimap: MinimapRenderer;

	private animationFrame: number | null = null;
	private spriteManager: SpriteAssetManager;
	private terrainSpritePool: SpritePool = new SpritePool();
	private tileSpriteContainer: Container;
	private agents: Agent[] = [];

	constructor() {
		this.ctx = new RenderContext();
		this.ctx.app = new Application();
		this.ctx.worldContainer = new Container();
		this.ctx.agentContainer = new Container();
		this.ctx.objectContainer = new Container();
		this.ctx.gridGraphics = new Graphics();
		this.ctx.tileGraphics = new Graphics();
		this.ctx.fogGraphics = new Graphics();
		this.tileSpriteContainer = new Container();
		this.spriteManager = new SpriteAssetManager();

		this.viewport = new ViewportManager(this.ctx);
		this.input = new InputHandler(this.ctx, this.viewport);
		this.fog = new FogOfWarRenderer(this.ctx);
		this.tiles_ = new TileRenderer(this.ctx, this.viewport, this.fog);
		this.agents_ = new AgentRenderer(this.ctx);
		this.objects_ = new WorldObjectRenderer(this.ctx);
		this.minimap = new MinimapRenderer(this.ctx, this.viewport, this.tiles_);
	}

	async init(canvas: HTMLCanvasElement) {
		await this.ctx.app.init({
			canvas,
			width: 800,
			height: 600,
			backgroundColor: 0x1a1a2e,
			antialias: false,
			resolution: window.devicePixelRatio || 1,
			autoDensity: true
		});

		// Layer order: terrain sprites -> ownership/flat tiles -> fog -> grid -> objects -> agents
		this.ctx.worldContainer.addChild(this.tileSpriteContainer);
		this.ctx.worldContainer.addChild(this.ctx.tileGraphics);
		this.ctx.worldContainer.addChild(this.ctx.fogGraphics);
		this.ctx.worldContainer.addChild(this.ctx.gridGraphics);
		this.ctx.worldContainer.addChild(this.ctx.objectContainer);
		this.ctx.worldContainer.addChild(this.ctx.agentContainer);
		this.ctx.app.stage.addChild(this.ctx.worldContainer);

		this.spriteManager.initialize();
		this.tiles_.initMask();
		this.minimap.init();

		this.viewport.setOnViewportChange(() => {
			const bounds = this.viewport.getViewportBounds();
			if (this.viewport.viewportChanged(bounds)) {
				this.renderVisibleTiles();
			}
			this.minimap.updateViewportIndicator();
		});

		this.viewport.centerWorld();
		this.input.setup();
		this.startAnimationLoop();
	}

	// --- Public API ---

	setOnTileClick(cb: (tileX: number, tileY: number, screenX: number, screenY: number) => void) {
		this.input.setOnTileClick(cb);
	}

	setOnTileHover(cb: (tileX: number, tileY: number, screenX: number, screenY: number) => void) {
		this.input.setOnTileHover(cb);
	}

	setOnTileHoverEnd(cb: () => void) {
		this.input.setOnTileHoverEnd(cb);
	}

	setOnCameraMove(cb: () => void) {
		this.input.setOnCameraMove(cb);
	}

	setPlayerAgentId(agentId: string | null) {
		this.ctx.playerAgentId = agentId;
	}

	setFogOfWarEnabled(enabled: boolean) {
		this.fog.setFogOfWarEnabled(enabled);
	}

	updateVisibility(visible: Set<string>, explored: Set<string>): Set<string> {
		const newlyExplored = this.fog.updateVisibility(visible, explored);

		if (this.ctx.isLargeMap) {
			this.renderVisibleTiles();
		} else {
			this.updateTiles(this.ctx.tiles, this.agents);
		}

		return newlyExplored;
	}

	setExploredTiles(explored: Set<string>) {
		this.fog.setExploredTiles(explored);
	}

	setWorldSize(size: number) {
		this.viewport.setWorldSize(size);
		this.tiles_.drawGrid();
		this.viewport.centerWorld();

		if (this.ctx.isLargeMap) {
			this.viewport.clampPan();
		}
	}

	setTileIndex(tileIndex: Map<string, Tile>) {
		this.tiles_.setTileIndex(tileIndex);
	}

	updateTiles(tiles: Tile[], agents: Agent[]) {
		const startTime = performance.now();
		if (DEBUG_PERF) console.log(`[Renderer] updateTiles called with ${tiles.length} tiles, ${agents.length} agents`);

		this.ctx.tiles = tiles;
		this.agents = agents;

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
			this.minimap.render();
			perfLog(`updateTiles.total (large map)`, startTime);
			return;
		}

		// Small map: render all tiles with sprite support
		this.ctx.tileGraphics.clear();
		this.ctx.fogGraphics.clear();
		const gap = 1;
		const tileDrawSize = this.ctx.tileSize - (gap * 2);
		const useSprites = this.spriteManager.shouldUseSprites(this.ctx.tileSize);
		const activeSpriteKeys = useSprites ? new Set<string>() : null;

		for (const tile of tiles) {
			const px = tile.x * this.ctx.tileSize;
			const py = tile.y * this.ctx.tileSize;
			const key = `${tile.x},${tile.y}`;

			const isVisible = !this.ctx.fogOfWarEnabled || this.ctx.visibleTiles.has(key);
			const isExplored = this.ctx.exploredTiles.has(key);

			if (useSprites && tile.biome) {
				const texture = this.spriteManager.getTerrainTexture(tile.biome, tile.x, tile.y);
				if (texture) {
					activeSpriteKeys!.add(key);
					const sprite = this.terrainSpritePool.acquire(key);
					sprite.texture = texture;
					sprite.x = px;
					sprite.y = py;
					sprite.width = this.ctx.tileSize;
					sprite.height = this.ctx.tileSize;
					if (!sprite.parent) {
						this.tileSpriteContainer.addChild(sprite);
					}
				} else {
					const terrainColor = this.tiles_.getTileColor(tile);
					this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
					this.ctx.tileGraphics.fill({ color: terrainColor, alpha: 0.6 });
				}
			} else {
				const terrainColor = this.tiles_.getTileColor(tile);
				this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
				this.ctx.tileGraphics.fill({ color: terrainColor, alpha: 0.6 });
			}

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
				this.fog.renderFogForTile(key, px, py, false, true);
			} else {
				this.fog.renderFogForTile(key, px, py, false, false);
			}
		}

		if (activeSpriteKeys) {
			this.terrainSpritePool.releaseExcept(activeSpriteKeys, this.tileSpriteContainer);
		}

		this.minimap.hideForSmallMap();
		this.tileSpriteContainer.mask = null;
		this.ctx.objectContainer.mask = null;
		this.ctx.agentContainer.mask = null;

		perfLog(`updateTiles.total (small map)`, startTime);
	}

	refreshVisibleTiles() {
		if (this.ctx.isLargeMap) {
			this.renderVisibleTiles();
		}
	}

	updateAgents(agents: Agent[]) {
		this.agents_.updateAgents(agents);
	}

	updateReasonings(results: ActionResult[], currentTick: number) {
		this.agents_.updateReasonings(results, currentTick);
	}

	updateWorldObjects(objects: WorldObject[]) {
		this.objects_.updateWorldObjects(objects);
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
			this.updateTiles(this.ctx.tiles, this.agents);
		}
	}

	resize(width: number, height: number) {
		this.viewport.resize(width, height);

		if (this.ctx.isLargeMap) {
			this.minimap.updatePosition();
			this.minimap.updateViewportIndicator();
		}
	}

	screenToWorld(screenX: number, screenY: number): { x: number; y: number } {
		return this.viewport.screenToWorld(screenX, screenY);
	}

	getZoom(): number {
		return this.viewport.getZoom();
	}

	setZoom(zoom: number) {
		this.viewport.setZoom(zoom);
	}

	centerOn(x: number, y: number) {
		this.viewport.centerOn(x, y);
	}

	destroy() {
		if (this.animationFrame) {
			cancelAnimationFrame(this.animationFrame);
		}
		this.terrainSpritePool.destroy();
		this.spriteManager.destroy();
		this.ctx.app.destroy(true);
	}

	// --- Private ---

	private startAnimationLoop() {
		const animate = () => {
			this.agents_.animateAgents();
			this.animationFrame = requestAnimationFrame(animate);
		};
		this.animationFrame = requestAnimationFrame(animate);
	}

	private renderVisibleTiles() {
		const startTime = performance.now();
		const bounds = this.viewport.getViewportBounds();
		this.viewport.setLastViewport(bounds);

		const tilesInView = (bounds.maxX - bounds.minX + 1) * (bounds.maxY - bounds.minY + 1);
		const useSprites = this.spriteManager.shouldUseSprites(this.ctx.tileSize);

		this.ctx.tileGraphics.clear();
		this.ctx.fogGraphics.clear();

		const gap = this.ctx.isLargeMap ? 0 : 1;
		const tileDrawSize = this.ctx.tileSize - (gap * 2);
		const activeSpriteKeys = useSprites ? new Set<string>() : null;

		for (let y = bounds.minY; y <= bounds.maxY; y++) {
			for (let x = bounds.minX; x <= bounds.maxX; x++) {
				const key = `${x},${y}`;
				const tile = this.ctx.tileIndex.get(key);
				if (!tile) continue;

				const px = tile.x * this.ctx.tileSize;
				const py = tile.y * this.ctx.tileSize;

				const isVisible = !this.ctx.fogOfWarEnabled || this.ctx.visibleTiles.has(key);
				const isExplored = this.ctx.exploredTiles.has(key);

				if (useSprites && tile.biome) {
					const texture = this.spriteManager.getTerrainTexture(tile.biome, tile.x, tile.y);
					if (texture) {
						activeSpriteKeys!.add(key);
						const sprite = this.terrainSpritePool.acquire(key);
						sprite.texture = texture;
						sprite.x = px;
						sprite.y = py;
						sprite.width = this.ctx.tileSize;
						sprite.height = this.ctx.tileSize;
						if (!sprite.parent) {
							this.tileSpriteContainer.addChild(sprite);
						}
					} else {
						const terrainColor = this.tiles_.getTileColor(tile);
						this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
						this.ctx.tileGraphics.fill({ color: terrainColor, alpha: 1.0 });
					}
				} else {
					const terrainColor = this.tiles_.getTileColor(tile);
					const terrainAlpha = this.ctx.isLargeMap ? 1.0 : 0.6;
					this.ctx.tileGraphics.rect(px + gap, py + gap, tileDrawSize, tileDrawSize);
					this.ctx.tileGraphics.fill({ color: terrainColor, alpha: terrainAlpha });
				}

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
					this.fog.renderFogForTile(key, px, py, false, true);
				} else {
					this.fog.renderFogForTile(key, px, py, false, false);
				}
			}
		}

		if (activeSpriteKeys) {
			this.terrainSpritePool.releaseExcept(activeSpriteKeys, this.tileSpriteContainer);
		}

		// Minimap viewport indicator
		this.minimap.updateViewportIndicator();

		perfLog(`renderVisibleTiles.total (viewport: ${tilesInView}, sprites: ${useSprites})`, startTime);
	}
}
