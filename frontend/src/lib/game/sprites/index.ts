import type { Texture } from 'pixi.js';
import type { BiomeType } from '$lib/types';
import { MIN_SPRITE_TILE_SIZE } from './types';
import { TerrainSpriteGenerator } from './terrain/index';
import { ObjectSpriteGenerator } from './objects/index';
import { AgentSpriteGenerator } from './agents/index';

export class SpriteAssetManager {
	private terrain = new TerrainSpriteGenerator();
	private objects = new ObjectSpriteGenerator();
	private agents = new AgentSpriteGenerator();

	/** Generate all pre-cached textures. Call once after Pixi app is initialized. */
	initialize() {
		this.terrain.initialize();
		this.objects.initialize();
		// Agent textures are generated on demand (few unique colors)
	}

	getTerrainTexture(biome: BiomeType, x: number, y: number): Texture | null {
		return this.terrain.getTexture(biome, x, y);
	}

	getObjectTexture(type: string, subType?: string): Texture | null {
		return this.objects.getTexture(type, subType);
	}

	getAgentTexture(color: number, isPlayer: boolean): Texture {
		return this.agents.getTexture(color, isPlayer);
	}

	shouldUseSprites(tileSize: number): boolean {
		return tileSize >= MIN_SPRITE_TILE_SIZE;
	}

	destroy() {
		this.terrain.destroy();
		this.objects.destroy();
		this.agents.destroy();
	}
}
