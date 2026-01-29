import type { Texture } from 'pixi.js';
import { pixelDataToTexture } from '../texture-util';
import { generateWall, generateBeacon, generateTrap } from './structures';
import { generateWood, generateStone, generateCrystal, generateHerb } from './resources';
import { generateShrine, generateCache, generatePortal, generateObelisk } from './interactives';
import { generateDroppedItem } from './dropped-item';

type GeneratorFn = () => Uint8Array;

const OBJECT_GENERATORS: Record<string, GeneratorFn> = {
	'structure:wall': generateWall,
	'structure:beacon': generateBeacon,
	'structure:trap': generateTrap,
	'resource:wood': generateWood,
	'resource:stone': generateStone,
	'resource:crystal': generateCrystal,
	'resource:herb': generateHerb,
	'interactive:shrine': generateShrine,
	'interactive:cache': generateCache,
	'interactive:portal': generatePortal,
	'interactive:obelisk': generateObelisk,
	'dropped_item': generateDroppedItem
};

export class ObjectSpriteGenerator {
	private textures: Map<string, Texture> = new Map();

	initialize() {
		for (const [key, gen] of Object.entries(OBJECT_GENERATORS)) {
			const pixelData = gen();
			const texture = pixelDataToTexture(pixelData);
			this.textures.set(key, texture);
		}
	}

	getTexture(type: string, subType?: string): Texture | null {
		const key = subType ? `${type}:${subType}` : type;
		return this.textures.get(key) ?? null;
	}

	destroy() {
		for (const texture of this.textures.values()) {
			texture.destroy(true);
		}
		this.textures.clear();
	}
}
