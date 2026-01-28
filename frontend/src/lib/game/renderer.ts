import { Application, Container, Graphics, Text, TextStyle } from 'pixi.js';
import type { Agent, Tile } from '$lib/types';

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
	private gridGraphics: Graphics;
	private tileGraphics: Graphics;
	private agentSprites: Map<string, AgentSprite> = new Map();
	private worldSize: number = 20;
	private playerAgentId: string | null = null;
	private agentColors: Map<string, number> = new Map();
	private animationFrame: number | null = null;

	constructor() {
		this.app = new Application();
		this.worldContainer = new Container();
		this.agentContainer = new Container();
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
