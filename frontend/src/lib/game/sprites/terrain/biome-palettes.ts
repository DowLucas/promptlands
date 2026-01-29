import type { BiomePalette } from '../types';
import type { BiomeType } from '$lib/types';

export const BIOME_PALETTES: Record<BiomeType, BiomePalette> = {
	forest: {
		base: 0x228b22,
		secondary: 0x2ea043,
		detail: 0x1a6b1a,
		shadow: 0x145214,
		highlight: 0x3dbb3d
	},
	desert: {
		base: 0xedc9af,
		secondary: 0xd4a574,
		detail: 0xc49660,
		shadow: 0xb08050,
		highlight: 0xf5dcc5
	},
	volcanic: {
		base: 0x8b0000,
		secondary: 0xa52a2a,
		detail: 0xff4500,
		shadow: 0x5c0000,
		highlight: 0xff6a00
	},
	ice: {
		base: 0xb0e0e6,
		secondary: 0xadd8e6,
		detail: 0xe0f0ff,
		shadow: 0x87ceeb,
		highlight: 0xffffff
	},
	savanna: {
		base: 0xbdb76b,
		secondary: 0xa0982e,
		detail: 0x8b8000,
		shadow: 0x6b6b00,
		highlight: 0xd4cc60
	},
	badlands: {
		base: 0xcd853f,
		secondary: 0xb8732e,
		detail: 0xa0622a,
		shadow: 0x8b4513,
		highlight: 0xe09050
	},
	swamp: {
		base: 0x556b2f,
		secondary: 0x4a5e27,
		detail: 0x3d5020,
		shadow: 0x2e3d18,
		highlight: 0x6b8e23
	},
	crystal: {
		base: 0xe6e6fa,
		secondary: 0xd8bfd8,
		detail: 0xdda0dd,
		shadow: 0xba55d3,
		highlight: 0xfff0ff
	},
	void: {
		base: 0x191970,
		secondary: 0x0d0d50,
		detail: 0x6a0dad,
		shadow: 0x0a0a30,
		highlight: 0x9370db
	},
	neon: {
		base: 0x0a3a2a,
		secondary: 0x0d4d3a,
		detail: 0x39ff14,
		shadow: 0x062018,
		highlight: 0x7fff00
	},
	plasma: {
		base: 0xff1493,
		secondary: 0xff6eb4,
		detail: 0xff8c00,
		shadow: 0xcc0066,
		highlight: 0xffff00
	},
	ancient: {
		base: 0x8b4513,
		secondary: 0xa0522d,
		detail: 0xdaa520,
		shadow: 0x5c2e0e,
		highlight: 0xffd700
	},
	ocean: {
		base: 0x1e90ff,
		secondary: 0x1a75d0,
		detail: 0x4aa8ff,
		shadow: 0x0f5fa0,
		highlight: 0x87ceeb
	},
	mountain: {
		base: 0x696969,
		secondary: 0x808080,
		detail: 0x555555,
		shadow: 0x404040,
		highlight: 0xa0a0a0
	}
};
