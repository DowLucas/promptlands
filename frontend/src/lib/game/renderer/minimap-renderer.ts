import { Container, Graphics } from 'pixi.js';
import type { RenderContext } from './render-context';
import type { ViewportManager } from './viewport-manager';
import type { TileRenderer } from './tile-renderer';

const MINIMAP_SIZE = 150;
const MINIMAP_PADDING = 10;

export class MinimapRenderer {
	private ctx: RenderContext;
	private viewport: ViewportManager;
	private tileRenderer: TileRenderer;
	private miniMapContainer: Container | null = null;
	private miniMapBackground: Graphics | null = null;
	private miniMapTiles: Graphics | null = null;
	private miniMapViewport: Graphics | null = null;

	constructor(ctx: RenderContext, viewport: ViewportManager, tileRenderer: TileRenderer) {
		this.ctx = ctx;
		this.viewport = viewport;
		this.tileRenderer = tileRenderer;
	}

	init() {
		this.miniMapContainer = new Container();
		this.miniMapContainer.zIndex = 1000;

		this.miniMapBackground = new Graphics();
		this.miniMapContainer.addChild(this.miniMapBackground);

		this.miniMapTiles = new Graphics();
		this.miniMapContainer.addChild(this.miniMapTiles);

		this.miniMapViewport = new Graphics();
		this.miniMapContainer.addChild(this.miniMapViewport);

		this.ctx.app.stage.addChild(this.miniMapContainer);
		this.updatePosition();
	}

	updatePosition() {
		if (!this.miniMapContainer) return;
		this.miniMapContainer.x = MINIMAP_PADDING;
		this.miniMapContainer.y = this.ctx.app.screen.height - MINIMAP_SIZE - MINIMAP_PADDING;
	}

	render() {
		if (!this.miniMapBackground || !this.miniMapTiles || !this.miniMapViewport) return;
		if (!this.ctx.isLargeMap) {
			if (this.miniMapContainer) {
				this.miniMapContainer.visible = false;
			}
			return;
		}

		if (this.miniMapContainer) {
			this.miniMapContainer.visible = true;
		}

		// Draw background
		this.miniMapBackground.clear();
		this.miniMapBackground.rect(0, 0, MINIMAP_SIZE, MINIMAP_SIZE);
		this.miniMapBackground.fill({ color: 0x1a1a2e, alpha: 0.9 });
		this.miniMapBackground.setStrokeStyle({ width: 2, color: 0xffd700 });
		this.miniMapBackground.stroke();

		const tilesPerPixel = this.ctx.worldSize / MINIMAP_SIZE;

		this.miniMapTiles.clear();

		for (let my = 0; my < MINIMAP_SIZE; my += 1) {
			for (let mx = 0; mx < MINIMAP_SIZE; mx += 1) {
				const tileX = Math.floor(mx * tilesPerPixel);
				const tileY = Math.floor(my * tilesPerPixel);
				const key = `${tileX},${tileY}`;
				const tile = this.ctx.tileIndex.get(key);

				if (tile) {
					const isVisible =
						!this.ctx.fogOfWarEnabled || this.ctx.visibleTiles.has(key);
					const isExplored = this.ctx.exploredTiles.has(key);

					if (isVisible) {
						const color = this.tileRenderer.getTileColor(tile);
						this.miniMapTiles.rect(mx, my, 1, 1);
						this.miniMapTiles.fill({ color, alpha: 0.8 });
					} else if (isExplored) {
						const color = this.tileRenderer.getTileColor(tile);
						this.miniMapTiles.rect(mx, my, 1, 1);
						this.miniMapTiles.fill({ color, alpha: 0.3 });
					}
				}
			}
		}

		this.updateViewportIndicator();
	}

	updateViewportIndicator() {
		if (!this.miniMapViewport || !this.ctx.isLargeMap) return;

		this.miniMapViewport.clear();

		const bounds = this.viewport.getViewportBounds();
		const scale = MINIMAP_SIZE / this.ctx.worldSize;

		const vpX = bounds.minX * scale;
		const vpY = bounds.minY * scale;
		const vpW = (bounds.maxX - bounds.minX + 1) * scale;
		const vpH = (bounds.maxY - bounds.minY + 1) * scale;

		this.miniMapViewport.rect(vpX, vpY, vpW, vpH);
		this.miniMapViewport.fill({ color: 0xffffff, alpha: 0.2 });
		this.miniMapViewport.setStrokeStyle({ width: 2, color: 0xffffff });
		this.miniMapViewport.stroke();
	}

	hideForSmallMap() {
		if (this.miniMapContainer) {
			this.miniMapContainer.visible = false;
		}
	}
}
