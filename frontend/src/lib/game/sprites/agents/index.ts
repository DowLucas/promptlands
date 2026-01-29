import type { Texture } from 'pixi.js';
import { pixelDataToTexture } from '../texture-util';
import { generateAgentBody } from './body';

export class AgentSpriteGenerator {
	private textures: Map<string, Texture> = new Map();

	/** Get or generate an agent texture for the given color/player combo */
	getTexture(color: number, isPlayer: boolean): Texture {
		const key = `${color.toString(16)}:${isPlayer ? 'p' : 'e'}`;
		let texture = this.textures.get(key);
		if (!texture) {
			const pixelData = generateAgentBody(color, isPlayer);
			texture = pixelDataToTexture(pixelData);
			this.textures.set(key, texture);
		}
		return texture;
	}

	destroy() {
		for (const texture of this.textures.values()) {
			texture.destroy(true);
		}
		this.textures.clear();
	}
}
