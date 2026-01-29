import { Container, Graphics, Text, TextStyle } from 'pixi.js';
import type { Agent, ActionResult } from '$lib/types';
import type { RenderContext } from './render-context';
import type { AgentSprite } from './types';
import { getAgentColor, perfLog } from './types';

export class AgentRenderer {
	private ctx: RenderContext;
	private agentSprites: Map<string, AgentSprite> = new Map();

	constructor(ctx: RenderContext) {
		this.ctx = ctx;
	}

	private currentTick = 0;

	animateAgents() {
		const speed = 0.15;

		for (const [, sprite] of this.agentSprites) {
			const dx = sprite.targetX - sprite.currentX;
			const dy = sprite.targetY - sprite.currentY;

			if (Math.abs(dx) > 0.5 || Math.abs(dy) > 0.5) {
				sprite.currentX += dx * speed;
				sprite.currentY += dy * speed;
			} else {
				sprite.currentX = sprite.targetX;
				sprite.currentY = sprite.targetY;
			}

			sprite.container.x = sprite.currentX;
			sprite.container.y = sprite.currentY;

			// Remove expired reasoning bubbles
			if (sprite.reasoningExpiry !== undefined && this.currentTick > sprite.reasoningExpiry) {
				this.removeReasoningBubble(sprite);
			}
		}
	}

	updateReasonings(results: ActionResult[], currentTick: number) {
		this.currentTick = currentTick;

		for (const result of results) {
			if (!result.reasoning) continue;

			const sprite = this.agentSprites.get(result.agent_id);
			if (!sprite) continue;

			// Remove old bubble if present
			this.removeReasoningBubble(sprite);

			// Create new reasoning bubble
			const maxWidth = Math.max(80, Math.min(120, this.ctx.tileSize * 4));
			const fontSize = Math.max(7, Math.min(10, this.ctx.tileSize / 3));

			const textStyle = new TextStyle({
				fontSize,
				fill: 0xffffff,
				fontFamily: 'Arial',
				wordWrap: true,
				wordWrapWidth: maxWidth - 8
			});

			const text = new Text({ text: result.reasoning, style: textStyle });
			text.anchor.set(0.5, 1);

			const padding = 4;
			const bgWidth = Math.min(maxWidth, text.width + padding * 2);
			const bgHeight = text.height + padding * 2;

			const bg = new Graphics();
			bg.roundRect(-bgWidth / 2, -bgHeight, bgWidth, bgHeight, 4);
			bg.fill({ color: 0x1a1a2e, alpha: 0.85 });
			bg.stroke({ width: 1, color: 0x4a5568, alpha: 0.6 });

			// Position above the agent name label
			const yOffset = -(this.ctx.tileSize / 2 + bgHeight + 8);
			bg.y = yOffset;
			text.y = yOffset;

			sprite.container.addChild(bg);
			sprite.container.addChild(text);
			sprite.reasoningBg = bg;
			sprite.reasoningText = text;
			sprite.reasoningExpiry = currentTick + 2;
		}
	}

	private removeReasoningBubble(sprite: AgentSprite) {
		if (sprite.reasoningText) {
			sprite.container.removeChild(sprite.reasoningText);
			sprite.reasoningText.destroy();
			sprite.reasoningText = undefined;
		}
		if (sprite.reasoningBg) {
			sprite.container.removeChild(sprite.reasoningBg);
			sprite.reasoningBg.destroy();
			sprite.reasoningBg = undefined;
		}
		sprite.reasoningExpiry = undefined;
	}

	updateAgents(agents: Agent[]) {
		const startTime = performance.now();

		const visibleAgents = this.ctx.fogOfWarEnabled
			? agents.filter(
					(agent) =>
						agent.id === this.ctx.playerAgentId ||
						this.ctx.visibleTiles.has(`${agent.position.x},${agent.position.y}`)
				)
			: agents;

		// Remove old sprites
		const currentIds = new Set(visibleAgents.map((a) => a.id));
		for (const [id, sprite] of this.agentSprites) {
			if (!currentIds.has(id)) {
				this.ctx.agentContainer.removeChild(sprite.container);
				this.agentSprites.delete(id);
			}
		}

		for (const agent of visibleAgents) {
			if (agent.is_dead) {
				const existingSprite = this.agentSprites.get(agent.id);
				if (existingSprite) {
					this.ctx.agentContainer.removeChild(existingSprite.container);
					this.agentSprites.delete(agent.id);
				}
				continue;
			}

			const targetX = agent.position.x * this.ctx.tileSize + this.ctx.tileSize / 2;
			const targetY = agent.position.y * this.ctx.tileSize + this.ctx.tileSize / 2;

			let sprite = this.agentSprites.get(agent.id);

			if (!sprite) {
				const container = this.createAgentSprite(agent);
				sprite = {
					container,
					targetX,
					targetY,
					currentX: targetX,
					currentY: targetY
				};
				this.agentSprites.set(agent.id, sprite);
				this.ctx.agentContainer.addChild(container);
			} else {
				sprite.targetX = targetX;
				sprite.targetY = targetY;
			}
		}
		perfLog(`updateAgents (${visibleAgents.length}/${agents.length} visible)`, startTime);
	}

	private createAgentSprite(agent: Agent): Container {
		const container = new Container();
		const isPlayer = agent.id === this.ctx.playerAgentId;
		const color = getAgentColor(agent.id, isPlayer);

		const agentRadius = Math.max(2, this.ctx.tileSize / 2.5);
		const strokeWidth = this.ctx.isLargeMap ? 1 : 2;

		const body = new Graphics();
		body.circle(0, 0, agentRadius);
		body.fill({ color });
		body.setStrokeStyle({ width: strokeWidth, color: isPlayer ? 0xffffff : 0x333333 });
		body.stroke();
		container.addChild(body);

		if (this.ctx.tileSize >= 12) {
			const highlight = new Graphics();
			highlight.circle(-agentRadius / 5, -agentRadius / 5, agentRadius / 3);
			highlight.fill({ color: 0xffffff, alpha: 0.3 });
			container.addChild(highlight);
		}

		if (this.ctx.tileSize >= 8) {
			const fontSize = Math.max(6, Math.min(11, this.ctx.tileSize / 2));
			const style = new TextStyle({
				fontSize,
				fill: 0xffffff,
				fontFamily: 'Arial',
				fontWeight: 'bold',
				stroke: { color: 0x000000, width: this.ctx.isLargeMap ? 2 : 3 }
			});
			const label = new Text({ text: agent.name, style });
			label.anchor.set(0.5, 2.5);
			container.addChild(label);
		}

		container.x = agent.position.x * this.ctx.tileSize + this.ctx.tileSize / 2;
		container.y = agent.position.y * this.ctx.tileSize + this.ctx.tileSize / 2;

		return container;
	}
}
