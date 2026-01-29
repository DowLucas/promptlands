import type { Texture } from 'pixi.js';
import type { BiomeType } from '$lib/types';
import { BIOME_VARIANTS, ALL_BIOMES, tileVariant } from '../types';
import { pixelDataToTexture } from '../texture-util';
import { BIOME_PALETTES } from './biome-palettes';
import { BIOME_GENERATORS } from './biome-generators';

export class TerrainSpriteGenerator {
	private textures: Map<string, Texture> = new Map();

	/** Generate all terrain textures (14 biomes x 4 variants = 56 textures) */
	initialize() {
		for (const biome of ALL_BIOMES) {
			const palette = BIOME_PALETTES[biome];
			const generator = BIOME_GENERATORS[biome];

			for (let v = 0; v < BIOME_VARIANTS; v++) {
				const pixelData = generator(v, palette);
				const texture = pixelDataToTexture(pixelData);
				this.textures.set(`${biome}:${v}`, texture);
			}
		}
	}

	getTexture(biome: BiomeType, x: number, y: number): Texture | null {
		const v = tileVariant(x, y);
		return this.textures.get(`${biome}:${v}`) ?? null;
	}

	destroy() {
		for (const texture of this.textures.values()) {
			texture.destroy(true);
		}
		this.textures.clear();
	}
}
