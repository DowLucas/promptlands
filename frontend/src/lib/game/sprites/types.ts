import type { Texture } from 'pixi.js';
import type { BiomeType } from '$lib/types';

export const SPRITE_RESOLUTION = 16;
export const BIOME_VARIANTS = 4;
export const MIN_SPRITE_TILE_SIZE = 8;

export interface BiomePalette {
	base: number;
	secondary: number;
	detail: number;
	shadow: number;
	highlight: number;
}

export interface TextureEntry {
	texture: Texture;
	label: string;
}

export interface SpriteSystemConfig {
	resolution: number;
	variants: number;
	minTileSize: number;
}

export type BiomeGenerator = (variant: number, palette: BiomePalette) => Uint8Array;

export type ObjectTextureKey = string; // e.g. "structure:wall", "resource:wood"
export type AgentTextureKey = string; // hex color string

export function tileVariant(x: number, y: number): number {
	return ((x * 7 + y * 13) & 0x7fffffff) % BIOME_VARIANTS;
}

export function seededRandom(seed: number): () => number {
	let s = seed | 0;
	return () => {
		s = (s * 1664525 + 1013904223) | 0;
		return ((s >>> 0) / 4294967296);
	};
}

export const ALL_BIOMES: BiomeType[] = [
	'forest', 'desert', 'volcanic', 'ice', 'savanna', 'badlands',
	'swamp', 'crystal', 'void', 'neon', 'plasma', 'ancient',
	'ocean', 'mountain'
];
