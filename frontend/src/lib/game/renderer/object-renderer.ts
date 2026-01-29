import { Container, Graphics } from 'pixi.js';
import type { WorldObject } from '$lib/types';
import type { RenderContext } from './render-context';
import { perfLog } from './types';

export class WorldObjectRenderer {
	private ctx: RenderContext;
	private objectSprites: Map<string, Container> = new Map();

	constructor(ctx: RenderContext) {
		this.ctx = ctx;
	}

	updateWorldObjects(objects: WorldObject[]) {
		const startTime = performance.now();

		const visibleObjects = this.ctx.fogOfWarEnabled
			? objects.filter(
					(obj) =>
						this.ctx.visibleTiles.has(`${obj.position.x},${obj.position.y}`) ||
						obj.owner_id === this.ctx.playerAgentId
				)
			: objects;

		const maxObjects = this.ctx.isLargeMap ? 500 : visibleObjects.length;
		const objectsToRender = visibleObjects.slice(0, maxObjects);

		// Remove old sprites
		const currentIds = new Set(objectsToRender.map((o) => o.id));
		for (const [id, sprite] of this.objectSprites) {
			if (!currentIds.has(id)) {
				this.ctx.objectContainer.removeChild(sprite);
				this.objectSprites.delete(id);
			}
		}

		let created = 0;
		for (const obj of objectsToRender) {
			if (this.objectSprites.has(obj.id)) continue;

			const container = this.createObjectSprite(obj);
			if (container) {
				this.objectSprites.set(obj.id, container);
				this.ctx.objectContainer.addChild(container);
				created++;
			}
		}
		perfLog(
			`updateWorldObjects (${visibleObjects.length}/${objects.length} visible, ${created} created)`,
			startTime
		);
	}

	private createObjectSprite(obj: WorldObject): Container | null {
		const container = new Container();
		const x = obj.position.x * this.ctx.tileSize + this.ctx.tileSize / 2;
		const y = obj.position.y * this.ctx.tileSize + this.ctx.tileSize / 2;

		const graphics = new Graphics();

		switch (obj.type) {
			case 'structure':
				this.drawStructure(graphics, obj);
				break;
			case 'resource':
				this.drawResource(graphics, obj);
				break;
			case 'interactive':
				this.drawInteractive(graphics, obj);
				break;
			case 'dropped_item':
				this.drawDroppedItem(graphics);
				break;
			default:
				return null;
		}

		container.addChild(graphics);
		container.x = x;
		container.y = y;
		return container;
	}

	private drawStructure(g: Graphics, obj: WorldObject) {
		const size = Math.max(2, this.ctx.tileSize / 2.5);
		switch (obj.structure_type) {
			case 'wall':
				g.rect(-size, -size, size * 2, size * 2);
				g.fill({ color: 0x8b4513 });
				if (!this.ctx.isLargeMap) {
					g.setStrokeStyle({ width: 1, color: 0x5c3317 });
					g.stroke();
				}
				break;
			case 'beacon':
				g.moveTo(0, -size);
				g.lineTo(size, 0);
				g.lineTo(0, size);
				g.lineTo(-size, 0);
				g.closePath();
				g.fill({ color: 0x00ffff, alpha: 0.7 });
				if (!this.ctx.isLargeMap) {
					g.setStrokeStyle({ width: 2, color: 0x00cccc });
					g.stroke();
				}
				break;
			case 'trap':
				if (!obj.hidden) {
					g.circle(0, 0, size / 2);
					g.fill({ color: 0xff0000, alpha: 0.5 });
				}
				break;
		}
	}

	private drawResource(g: Graphics, obj: WorldObject) {
		const size = Math.max(2, this.ctx.tileSize / 3);
		let color = 0x888888;

		switch (obj.resource_type) {
			case 'wood':
				color = 0x8b4513;
				break;
			case 'stone':
				color = 0x808080;
				break;
			case 'crystal':
				color = 0x9966ff;
				break;
			case 'herb':
				color = 0x00ff00;
				break;
		}

		g.circle(0, 0, size);
		g.fill({ color, alpha: 0.8 });
		if (!this.ctx.isLargeMap) {
			g.setStrokeStyle({ width: 1, color: 0x333333 });
			g.stroke();
		}
	}

	private drawInteractive(g: Graphics, obj: WorldObject) {
		const size = Math.max(2, this.ctx.tileSize / 3);

		switch (obj.interactive_type) {
			case 'shrine':
				this.drawStar(g, 0, 0, size, 5, 0xffd700);
				break;
			case 'cache':
				g.rect(-size, -size / 2, size * 2, size);
				g.fill({ color: 0xdaa520 });
				if (!this.ctx.isLargeMap) {
					g.setStrokeStyle({ width: 1, color: 0x8b4513 });
					g.stroke();
				}
				break;
			case 'portal':
				g.circle(0, 0, size);
				g.fill({ color: 0x9932cc, alpha: 0.7 });
				g.circle(0, 0, size / 2);
				g.fill({ color: 0xff00ff, alpha: 0.5 });
				break;
			case 'obelisk':
				g.rect(-size / 3, -size, (size * 2) / 3, size * 2);
				g.fill({ color: 0x696969 });
				if (!this.ctx.isLargeMap) {
					g.setStrokeStyle({ width: 1, color: 0x333333 });
					g.stroke();
				}
				break;
		}

		if (obj.activated && this.ctx.tileSize >= 8) {
			g.circle(0, -size - 4, 3);
			g.fill({ color: 0x00ff00 });
		}
	}

	private drawDroppedItem(g: Graphics) {
		const size = Math.max(2, this.ctx.tileSize / 4);
		g.circle(0, 0, size);
		g.fill({ color: 0xffa500, alpha: 0.8 });
		if (!this.ctx.isLargeMap) {
			g.setStrokeStyle({ width: 1, color: 0x8b4513 });
			g.stroke();
		}
	}

	private drawStar(
		g: Graphics,
		cx: number,
		cy: number,
		size: number,
		points: number,
		color: number
	) {
		const outerRadius = size;
		const innerRadius = size / 2;

		g.moveTo(cx, cy - outerRadius);
		for (let i = 0; i < points * 2; i++) {
			const radius = i % 2 === 0 ? outerRadius : innerRadius;
			const angle = (Math.PI * i) / points - Math.PI / 2;
			g.lineTo(cx + Math.cos(angle) * radius, cy + Math.sin(angle) * radius);
		}
		g.closePath();
		g.fill({ color });
	}
}
