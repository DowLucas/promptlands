import { Application, Container, Graphics, Text, TextStyle } from 'pixi.js';
import type { Agent, Tile, WorldObject } from '$lib/types';

const TILE_SIZE = 28;
const COLORS = {
	background: 0x1a1a2e,
	empty: 0x2d3748,
	plains: 0x7cb342,    // Light green (improved visibility)
	forest: 0x2e7d32,    // Dark green (improved visibility)
	mountain: 0x78909c,  // Blue-gray (improved visibility)
	water: 0x1976d2,     // Ocean blue (improved visibility)
	gridLine: 0x4a5568,
	player: 0xffd700,
	adversary: 0xff6b6b,
	neutral: 0x888888,
	border: 0x6366f1
};

// Generate distinct colors for different agents
function getAgentColor(agentId: string, isPlayer: boolean): number {
	if (isPlayer) return COLORS.player;

	// Generate a color based on agent ID hash
	let hash = 0;
	for (let i = 0; i < agentId.length; i++) {
		hash = agentId.charCodeAt(i) + ((hash << 5) - hash);
	}

	const hue = Math.abs(hash % 360);
	return hslToHex(hue, 70, 50);
}

function hslToHex(h: number, s: number, l: number): number {
	s /= 100;
	l /= 100;
	const a = s * Math.min(l, 1 - l);
	const f = (n: number) => {
		const k = (n + h / 30) % 12;
		const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
		return Math.round(255 * color);
	};
	return (f(0) << 16) | (f(8) << 8) | f(4);
}

interface AgentSprite {
	container: Container;
	targetX: number;
	targetY: number;
	currentX: number;
	currentY: number;
}

export class GameRenderer {
	private app: Application;
	private worldContainer: Container;
	private agentContainer: Container;
	private objectContainer: Container;
	private gridGraphics: Graphics;
	private tileGraphics: Graphics;
	private agentSprites: Map<string, AgentSprite> = new Map();
	private objectSprites: Map<string, Container> = new Map();
	private worldSize: number = 20;
	private playerAgentId: string | null = null;
	private agentColors: Map<string, number> = new Map();
	private animationFrame: number | null = null;

	constructor() {
		this.app = new Application();
		this.worldContainer = new Container();
		this.agentContainer = new Container();
		this.objectContainer = new Container();
		this.gridGraphics = new Graphics();
		this.tileGraphics = new Graphics();
	}

	async init(canvas: HTMLCanvasElement) {
		await this.app.init({
			canvas,
			width: 800,
			height: 600,
			backgroundColor: 0x1a1a2e,
			antialias: true,
			resolution: window.devicePixelRatio || 1,
			autoDensity: true
		});

		this.worldContainer.addChild(this.tileGraphics);
		this.worldContainer.addChild(this.gridGraphics);
		this.worldContainer.addChild(this.objectContainer);
		this.worldContainer.addChild(this.agentContainer);
		this.app.stage.addChild(this.worldContainer);

		// Center the world
		this.centerWorld();

		// Make world draggable
		this.setupInteraction();

		// Start animation loop
		this.startAnimationLoop();
	}

	private startAnimationLoop() {
		const animate = () => {
			this.animateAgents();
			this.animationFrame = requestAnimationFrame(animate);
		};
		this.animationFrame = requestAnimationFrame(animate);
	}

	private animateAgents() {
		const speed = 0.15; // Adjust for faster/slower movement

		for (const [id, sprite] of this.agentSprites) {
			// Calculate distance to target
			const dx = sprite.targetX - sprite.currentX;
			const dy = sprite.targetY - sprite.currentY;

			// Move towards target
			if (Math.abs(dx) > 0.5 || Math.abs(dy) > 0.5) {
				sprite.currentX += dx * speed;
				sprite.currentY += dy * speed;
			} else {
				// Snap to target when close
				sprite.currentX = sprite.targetX;
				sprite.currentY = sprite.targetY;
			}

			// Update visual position
			sprite.container.x = sprite.currentX;
			sprite.container.y = sprite.currentY;
		}
	}

	private centerWorld() {
		const worldPixelSize = this.worldSize * TILE_SIZE;
		this.worldContainer.x = (this.app.screen.width - worldPixelSize) / 2;
		this.worldContainer.y = (this.app.screen.height - worldPixelSize) / 2;
	}

	private setupInteraction() {
		let dragging = false;
		let lastPos = { x: 0, y: 0 };

		this.app.stage.eventMode = 'static';
		this.app.stage.hitArea = this.app.screen;

		this.app.stage.on('pointerdown', (e) => {
			dragging = true;
			lastPos = { x: e.global.x, y: e.global.y };
		});

		this.app.stage.on('pointerup', () => {
			dragging = false;
		});

		this.app.stage.on('pointerupoutside', () => {
			dragging = false;
		});

		this.app.stage.on('pointermove', (e) => {
			if (dragging) {
				const dx = e.global.x - lastPos.x;
				const dy = e.global.y - lastPos.y;
				this.worldContainer.x += dx;
				this.worldContainer.y += dy;
				lastPos = { x: e.global.x, y: e.global.y };
			}
		});

		// Zoom with scroll wheel
		this.app.canvas.addEventListener('wheel', (e) => {
			e.preventDefault();
			const scale = this.worldContainer.scale.x;
			const delta = e.deltaY > 0 ? 0.9 : 1.1;
			const newScale = Math.max(0.5, Math.min(3, scale * delta));
			this.worldContainer.scale.set(newScale);
		});
	}

	setPlayerAgentId(agentId: string | null) {
		this.playerAgentId = agentId;
	}

	setWorldSize(size: number) {
		this.worldSize = size;
		this.drawGrid();
		this.centerWorld();
	}

	private drawGrid() {
		this.gridGraphics.clear();

		const size = this.worldSize * TILE_SIZE;

		// Draw grid lines
		this.gridGraphics.setStrokeStyle({ width: 1, color: COLORS.gridLine, alpha: 0.3 });

		for (let i = 0; i <= this.worldSize; i++) {
			// Vertical lines
			this.gridGraphics.moveTo(i * TILE_SIZE, 0);
			this.gridGraphics.lineTo(i * TILE_SIZE, size);
			// Horizontal lines
			this.gridGraphics.moveTo(0, i * TILE_SIZE);
			this.gridGraphics.lineTo(size, i * TILE_SIZE);
		}
		this.gridGraphics.stroke();
	}

	updateTiles(tiles: Tile[], agents: Agent[]) {
		// Build agent color map
		this.agentColors.clear();
		for (const agent of agents) {
			const isPlayer = agent.id === this.playerAgentId;
			this.agentColors.set(agent.id, getAgentColor(agent.id, isPlayer));
		}

		this.tileGraphics.clear();

		for (const tile of tiles) {
			const x = tile.x * TILE_SIZE;
			const y = tile.y * TILE_SIZE;

			// Get base terrain color
			let terrainColor = COLORS.plains;
			switch (tile.terrain) {
				case 'forest': terrainColor = COLORS.forest; break;
				case 'mountain': terrainColor = COLORS.mountain; break;
				case 'water': terrainColor = COLORS.water; break;
			}

			// Draw base terrain
			this.tileGraphics.rect(x + 1, y + 1, TILE_SIZE - 2, TILE_SIZE - 2);
			this.tileGraphics.fill({ color: terrainColor, alpha: 0.6 });

			// If owned, draw ownership overlay on top of terrain
			if (tile.owner_id) {
				const ownerColor = this.agentColors.get(tile.owner_id) || COLORS.neutral;
				this.tileGraphics.rect(x + 1, y + 1, TILE_SIZE - 2, TILE_SIZE - 2);
				this.tileGraphics.fill({ color: ownerColor, alpha: 0.5 });
			}
		}
	}

	updateAgents(agents: Agent[]) {
		// Remove old sprites
		const currentIds = new Set(agents.map(a => a.id));
		for (const [id, sprite] of this.agentSprites) {
			if (!currentIds.has(id)) {
				this.agentContainer.removeChild(sprite.container);
				this.agentSprites.delete(id);
			}
		}

		// Update or create sprites
		for (const agent of agents) {
			// Skip dead agents - don't render them
			if (agent.is_dead) {
				const existingSprite = this.agentSprites.get(agent.id);
				if (existingSprite) {
					this.agentContainer.removeChild(existingSprite.container);
					this.agentSprites.delete(agent.id);
				}
				continue;
			}

			const targetX = agent.position.x * TILE_SIZE + TILE_SIZE / 2;
			const targetY = agent.position.y * TILE_SIZE + TILE_SIZE / 2;

			let sprite = this.agentSprites.get(agent.id);

			if (!sprite) {
				// Create new sprite at current position
				const container = this.createAgentSprite(agent);
				sprite = {
					container,
					targetX,
					targetY,
					currentX: targetX,
					currentY: targetY
				};
				this.agentSprites.set(agent.id, sprite);
				this.agentContainer.addChild(container);
			} else {
				// Update target position (animation loop will handle movement)
				sprite.targetX = targetX;
				sprite.targetY = targetY;
			}
		}
	}

	private createAgentSprite(agent: Agent): Container {
		const container = new Container();
		const isPlayer = agent.id === this.playerAgentId;
		const color = getAgentColor(agent.id, isPlayer);

		// Agent body - larger and more visible
		const body = new Graphics();
		body.circle(0, 0, TILE_SIZE / 2.5);
		body.fill({ color });
		body.setStrokeStyle({ width: 2, color: isPlayer ? 0xffffff : 0x333333 });
		body.stroke();
		container.addChild(body);

		// Inner highlight
		const highlight = new Graphics();
		highlight.circle(-2, -2, TILE_SIZE / 6);
		highlight.fill({ color: 0xffffff, alpha: 0.3 });
		container.addChild(highlight);

		// Name label
		const style = new TextStyle({
			fontSize: 11,
			fill: 0xffffff,
			fontFamily: 'Arial',
			fontWeight: 'bold',
			stroke: { color: 0x000000, width: 3 }
		});
		const label = new Text({ text: agent.name, style });
		label.anchor.set(0.5, 2.5);
		container.addChild(label);

		container.x = agent.position.x * TILE_SIZE + TILE_SIZE / 2;
		container.y = agent.position.y * TILE_SIZE + TILE_SIZE / 2;

		return container;
	}

	updateWorldObjects(objects: WorldObject[]) {
		// Remove old sprites
		const currentIds = new Set(objects.map(o => o.id));
		for (const [id, sprite] of this.objectSprites) {
			if (!currentIds.has(id)) {
				this.objectContainer.removeChild(sprite);
				this.objectSprites.delete(id);
			}
		}

		// Update or create sprites
		for (const obj of objects) {
			if (this.objectSprites.has(obj.id)) continue;

			const container = this.createObjectSprite(obj);
			if (container) {
				this.objectSprites.set(obj.id, container);
				this.objectContainer.addChild(container);
			}
		}
	}

	private createObjectSprite(obj: WorldObject): Container | null {
		const container = new Container();
		const x = obj.position.x * TILE_SIZE + TILE_SIZE / 2;
		const y = obj.position.y * TILE_SIZE + TILE_SIZE / 2;

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
		const size = TILE_SIZE / 2.5;
		switch (obj.structure_type) {
			case 'wall':
				// Brown square for wall
				g.rect(-size, -size, size * 2, size * 2);
				g.fill({ color: 0x8b4513 });
				g.setStrokeStyle({ width: 1, color: 0x5c3317 });
				g.stroke();
				break;
			case 'beacon':
				// Cyan diamond for beacon
				g.moveTo(0, -size);
				g.lineTo(size, 0);
				g.lineTo(0, size);
				g.lineTo(-size, 0);
				g.closePath();
				g.fill({ color: 0x00ffff, alpha: 0.7 });
				g.setStrokeStyle({ width: 2, color: 0x00cccc });
				g.stroke();
				break;
			case 'trap':
				// Hidden traps shouldn't be visible, but if revealed show as red
				if (!obj.hidden) {
					g.circle(0, 0, size / 2);
					g.fill({ color: 0xff0000, alpha: 0.5 });
				}
				break;
		}
	}

	private drawResource(g: Graphics, obj: WorldObject) {
		const size = TILE_SIZE / 3;
		let color = 0x888888;

		switch (obj.resource_type) {
			case 'wood':
				color = 0x8b4513; // Brown
				break;
			case 'stone':
				color = 0x808080; // Gray
				break;
			case 'crystal':
				color = 0x9966ff; // Purple
				break;
			case 'herb':
				color = 0x00ff00; // Green
				break;
		}

		// Draw as a small circle/node
		g.circle(0, 0, size);
		g.fill({ color, alpha: 0.8 });
		g.setStrokeStyle({ width: 1, color: 0x333333 });
		g.stroke();

		// Show remaining count
		if (obj.remaining !== undefined && obj.remaining > 0) {
			const style = new TextStyle({
				fontSize: 8,
				fill: 0xffffff,
				fontFamily: 'Arial',
				fontWeight: 'bold'
			});
			const text = new Text({ text: String(obj.remaining), style });
			text.anchor.set(0.5, 0.5);
			// Note: Text needs to be on parent container, not graphics
		}
	}

	private drawInteractive(g: Graphics, obj: WorldObject) {
		const size = TILE_SIZE / 3;

		switch (obj.interactive_type) {
			case 'shrine':
				// Gold star
				this.drawStar(g, 0, 0, size, 5, 0xffd700);
				break;
			case 'cache':
				// Brown chest
				g.rect(-size, -size / 2, size * 2, size);
				g.fill({ color: 0xdaa520 });
				g.setStrokeStyle({ width: 1, color: 0x8b4513 });
				g.stroke();
				break;
			case 'portal':
				// Purple swirl
				g.circle(0, 0, size);
				g.fill({ color: 0x9932cc, alpha: 0.7 });
				g.circle(0, 0, size / 2);
				g.fill({ color: 0xff00ff, alpha: 0.5 });
				break;
			case 'obelisk':
				// Gray pillar
				g.rect(-size / 3, -size, size * 2 / 3, size * 2);
				g.fill({ color: 0x696969 });
				g.setStrokeStyle({ width: 1, color: 0x333333 });
				g.stroke();
				break;
		}

		// Show if activated
		if (obj.activated) {
			g.circle(0, -size - 4, 3);
			g.fill({ color: 0x00ff00 });
		}
	}

	private drawDroppedItem(g: Graphics) {
		const size = TILE_SIZE / 4;
		// Simple yellow bag/pouch
		g.circle(0, 0, size);
		g.fill({ color: 0xffa500, alpha: 0.8 });
		g.setStrokeStyle({ width: 1, color: 0x8b4513 });
		g.stroke();
	}

	private drawStar(g: Graphics, cx: number, cy: number, size: number, points: number, color: number) {
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

	resize(width: number, height: number) {
		this.app.renderer.resize(width, height);
		this.centerWorld();
	}

	destroy() {
		if (this.animationFrame) {
			cancelAnimationFrame(this.animationFrame);
		}
		this.app.destroy(true);
	}
}
