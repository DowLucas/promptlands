import type { RenderContext } from './render-context';
import type { ViewportManager } from './viewport-manager';

export class InputHandler {
	private ctx: RenderContext;
	private viewport: ViewportManager;
	private onTileClick: ((tileX: number, tileY: number, screenX: number, screenY: number) => void) | null = null;
	private onTileHover: ((tileX: number, tileY: number, screenX: number, screenY: number) => void) | null = null;
	private onTileHoverEnd: (() => void) | null = null;
	private onCameraMove: (() => void) | null = null;

	constructor(ctx: RenderContext, viewport: ViewportManager) {
		this.ctx = ctx;
		this.viewport = viewport;
	}

	setOnTileClick(cb: (tileX: number, tileY: number, screenX: number, screenY: number) => void) {
		this.onTileClick = cb;
	}

	setOnTileHover(cb: (tileX: number, tileY: number, screenX: number, screenY: number) => void) {
		this.onTileHover = cb;
	}

	setOnTileHoverEnd(cb: () => void) {
		this.onTileHoverEnd = cb;
	}

	setOnCameraMove(cb: () => void) {
		this.onCameraMove = cb;
	}

	setup() {
		let dragging = false;
		let lastPos = { x: 0, y: 0 };
		let dragStartPos = { x: 0, y: 0 };
		let hasDragged = false;

		this.ctx.app.stage.eventMode = 'static';
		this.ctx.app.stage.hitArea = this.ctx.app.screen;

		this.ctx.app.stage.on('pointerdown', (e) => {
			dragging = true;
			hasDragged = false;
			lastPos = { x: e.global.x, y: e.global.y };
			dragStartPos = { x: e.global.x, y: e.global.y };
		});

		this.ctx.app.stage.on('pointerup', (e) => {
			if (!hasDragged && this.onTileClick) {
				const dist = Math.hypot(
					e.global.x - dragStartPos.x,
					e.global.y - dragStartPos.y
				);
				if (dist < 5) {
					const canvasX = e.global.x;
					const canvasY = e.global.y;
					const worldPos = this.viewport.screenToWorld(canvasX, canvasY);
					const tileX = Math.floor(worldPos.x / this.ctx.tileSize);
					const tileY = Math.floor(worldPos.y / this.ctx.tileSize);

					if (
						tileX >= 0 &&
						tileX < this.ctx.worldSize &&
						tileY >= 0 &&
						tileY < this.ctx.worldSize
					) {
						this.onTileClick(tileX, tileY, e.global.x, e.global.y);
					}
				}
			}
			dragging = false;
		});

		this.ctx.app.stage.on('pointerupoutside', () => {
			dragging = false;
		});

		this.ctx.app.stage.on('pointermove', (e) => {
			if (dragging) {
				const dx = e.global.x - lastPos.x;
				const dy = e.global.y - lastPos.y;

				if (Math.abs(dx) > 0 || Math.abs(dy) > 0) {
					hasDragged = true;
				}

				this.ctx.worldContainer.x += dx;
				this.ctx.worldContainer.y += dy;
				lastPos = { x: e.global.x, y: e.global.y };

				if (this.ctx.isLargeMap) {
					this.viewport.clampPan();
					this.viewport.scheduleViewportUpdate();
				}

				this.onCameraMove?.();
			} else if (this.onTileHover) {
				const worldPos = this.viewport.screenToWorld(e.global.x, e.global.y);
				const tileX = Math.floor(worldPos.x / this.ctx.tileSize);
				const tileY = Math.floor(worldPos.y / this.ctx.tileSize);

				if (
					tileX >= 0 &&
					tileX < this.ctx.worldSize &&
					tileY >= 0 &&
					tileY < this.ctx.worldSize
				) {
					this.onTileHover(tileX, tileY, e.global.x, e.global.y);
				} else {
					this.onTileHoverEnd?.();
				}
			}
		});

		this.ctx.app.stage.on('pointerleave', () => {
			this.onTileHoverEnd?.();
		});

		// Zoom with scroll wheel - cursor-centric zoom
		this.ctx.app.canvas.addEventListener('wheel', (e) => {
			e.preventDefault();

			const oldScale = this.ctx.worldContainer.scale.x;
			const delta = e.deltaY > 0 ? 0.9 : 1.1;
			const newScale = Math.max(
				this.ctx.minZoom,
				Math.min(this.ctx.maxZoom, oldScale * delta)
			);

			const mouseX = e.offsetX;
			const mouseY = e.offsetY;
			const worldPos = this.viewport.screenToWorld(mouseX, mouseY);

			this.ctx.worldContainer.scale.set(newScale);
			this.ctx.worldContainer.x = mouseX - worldPos.x * newScale;
			this.ctx.worldContainer.y = mouseY - worldPos.y * newScale;

			this.viewport.clampPan();
			if (this.ctx.isLargeMap) {
				this.viewport.scheduleViewportUpdate();
			}

			this.onCameraMove?.();
		});
	}
}
