import type { RenderContext } from './render-context';
import type { ViewportBounds } from './types';
import { BASE_TILE_SIZE } from './types';

export class ViewportManager {
	private ctx: RenderContext;
	private lastViewport: ViewportBounds | null = null;
	private viewportUpdateScheduled: boolean = false;
	private onViewportChange: (() => void) | null = null;

	constructor(ctx: RenderContext) {
		this.ctx = ctx;
	}

	setOnViewportChange(cb: () => void) {
		this.onViewportChange = cb;
	}

	centerWorld() {
		const worldPixelSize = this.ctx.worldSize * this.ctx.tileSize;
		this.ctx.worldContainer.x = (this.ctx.app.screen.width - worldPixelSize) / 2;
		this.ctx.worldContainer.y = (this.ctx.app.screen.height - worldPixelSize) / 2;
	}

	clampPan() {
		const scale = this.ctx.worldContainer.scale.x;
		const worldPixels = this.ctx.worldSize * this.ctx.tileSize * scale;
		const screenW = this.ctx.app.screen.width;
		const screenH = this.ctx.app.screen.height;

		// Buffer = half the screen so edge tiles can reach the viewport center
		const bufferX = screenW / 2;
		const bufferY = screenH / 2;

		if (worldPixels <= screenW) {
			this.ctx.worldContainer.x = (screenW - worldPixels) / 2;
		} else {
			const maxX = bufferX;
			const minX = screenW - worldPixels - bufferX;
			this.ctx.worldContainer.x = Math.min(
				Math.max(this.ctx.worldContainer.x, minX),
				maxX
			);
		}

		if (worldPixels <= screenH) {
			this.ctx.worldContainer.y = (screenH - worldPixels) / 2;
		} else {
			const maxY = bufferY;
			const minY = screenH - worldPixels - bufferY;
			this.ctx.worldContainer.y = Math.min(
				Math.max(this.ctx.worldContainer.y, minY),
				maxY
			);
		}
	}

	screenToWorld(screenX: number, screenY: number): { x: number; y: number } {
		const scale = this.ctx.worldContainer.scale.x;
		return {
			x: (screenX - this.ctx.worldContainer.x) / scale,
			y: (screenY - this.ctx.worldContainer.y) / scale
		};
	}

	getViewportBounds(): ViewportBounds {
		const scale = this.ctx.worldContainer.scale.x;
		const worldX = this.ctx.worldContainer.x;
		const worldY = this.ctx.worldContainer.y;
		const screenWidth = this.ctx.app.screen.width;
		const screenHeight = this.ctx.app.screen.height;

		const buffer = 5;
		let minX = Math.floor(-worldX / (this.ctx.tileSize * scale)) - buffer;
		let minY = Math.floor(-worldY / (this.ctx.tileSize * scale)) - buffer;
		let maxX = Math.ceil((screenWidth - worldX) / (this.ctx.tileSize * scale)) + buffer;
		let maxY = Math.ceil((screenHeight - worldY) / (this.ctx.tileSize * scale)) + buffer;

		minX = Math.max(0, minX);
		minY = Math.max(0, minY);
		maxX = Math.min(this.ctx.worldSize - 1, maxX);
		maxY = Math.min(this.ctx.worldSize - 1, maxY);

		const MAX_VIEWPORT_TILES = 200;
		const viewportWidth = maxX - minX + 1;
		const viewportHeight = maxY - minY + 1;

		if (viewportWidth > MAX_VIEWPORT_TILES) {
			const centerX = (minX + maxX) / 2;
			minX = Math.max(0, Math.floor(centerX - MAX_VIEWPORT_TILES / 2));
			maxX = Math.min(this.ctx.worldSize - 1, minX + MAX_VIEWPORT_TILES - 1);
		}
		if (viewportHeight > MAX_VIEWPORT_TILES) {
			const centerY = (minY + maxY) / 2;
			minY = Math.max(0, Math.floor(centerY - MAX_VIEWPORT_TILES / 2));
			maxY = Math.min(this.ctx.worldSize - 1, minY + MAX_VIEWPORT_TILES - 1);
		}

		return { minX, maxX, minY, maxY };
	}

	viewportChanged(newBounds: ViewportBounds): boolean {
		if (!this.lastViewport) return true;
		return (
			newBounds.minX !== this.lastViewport.minX ||
			newBounds.maxX !== this.lastViewport.maxX ||
			newBounds.minY !== this.lastViewport.minY ||
			newBounds.maxY !== this.lastViewport.maxY
		);
	}

	setLastViewport(bounds: ViewportBounds) {
		this.lastViewport = bounds;
	}

	scheduleViewportUpdate() {
		if (this.viewportUpdateScheduled) return;
		this.viewportUpdateScheduled = true;
		requestAnimationFrame(() => {
			this.viewportUpdateScheduled = false;
			this.onViewportChange?.();
		});
	}

	setWorldSize(size: number) {
		this.ctx.worldSize = size;

		if (size >= 1024) {
			this.ctx.isLargeMap = true;
			if (size >= 2048) {
				this.ctx.tileSize = 4;
				this.ctx.maxZoom = 5;
			} else {
				this.ctx.tileSize = 8;
				this.ctx.maxZoom = 4;
			}
		} else if (size >= 512) {
			this.ctx.isLargeMap = true;
			this.ctx.tileSize = 12;
			this.ctx.maxZoom = 3;
		} else {
			this.ctx.isLargeMap = false;
			this.ctx.tileSize = BASE_TILE_SIZE;
			this.ctx.maxZoom = 3;
		}

		const MAX_VIEWPORT_SIZE = 200;
		const maxViewportPixels = MAX_VIEWPORT_SIZE * this.ctx.tileSize;
		const screenSize = Math.max(this.ctx.app.screen.width, this.ctx.app.screen.height);
		this.ctx.minZoom = Math.max(0.1, screenSize / maxViewportPixels);

		if (this.ctx.isLargeMap) {
			this.ctx.worldContainer.scale.set(this.ctx.minZoom);
			this.clampPan();
		}
	}

	resize(width: number, height: number) {
		this.ctx.app.renderer.resize(width, height);

		const MAX_VIEWPORT_SIZE = 200;
		const maxViewportPixels = MAX_VIEWPORT_SIZE * this.ctx.tileSize;
		const screenSize = Math.max(width, height);
		this.ctx.minZoom = Math.max(0.1, screenSize / maxViewportPixels);

		const currentZoom = this.ctx.worldContainer.scale.x;
		if (currentZoom < this.ctx.minZoom) {
			this.ctx.worldContainer.scale.set(this.ctx.minZoom);
		}

		this.centerWorld();

		if (this.ctx.isLargeMap) {
			this.clampPan();
		}
	}

	getZoom(): number {
		return this.ctx.worldContainer.scale.x;
	}

	setZoom(zoom: number) {
		const clampedZoom = Math.max(this.ctx.minZoom, Math.min(this.ctx.maxZoom, zoom));
		this.ctx.worldContainer.scale.set(clampedZoom);
	}

	centerOn(x: number, y: number) {
		const worldX = x * this.ctx.tileSize + this.ctx.tileSize / 2;
		const worldY = y * this.ctx.tileSize + this.ctx.tileSize / 2;
		const scale = this.ctx.worldContainer.scale.x;
		this.ctx.worldContainer.x = this.ctx.app.screen.width / 2 - worldX * scale;
		this.ctx.worldContainer.y = this.ctx.app.screen.height / 2 - worldY * scale;

		if (this.ctx.isLargeMap) {
			this.clampPan();
			this.scheduleViewportUpdate();
		}
	}
}
