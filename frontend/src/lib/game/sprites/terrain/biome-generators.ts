import type { BiomePalette } from '../types';
import type { BiomeType } from '$lib/types';
import {
	createBuffer, fillBase, setPixel, scatterPixels,
	drawRect, drawLine, drawCircleFilled, drawDiamond
} from './patterns';

type BiomeGenFn = (variant: number, palette: BiomePalette) => Uint8Array;

function forestGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.15, variant * 111);
	scatterPixels(buf, p.shadow, 0.05, variant * 222);

	// Small tree silhouettes (0-2 per variant)
	if (variant === 0 || variant === 2) {
		// Tree trunk
		drawRect(buf, 7, 9, 2, 4, p.shadow);
		// Tree canopy
		drawRect(buf, 5, 5, 6, 4, p.detail);
		drawRect(buf, 6, 4, 4, 1, p.detail);
		setPixel(buf, 7, 3, p.secondary);
		setPixel(buf, 8, 3, p.secondary);
	}
	if (variant === 1) {
		// Two smaller trees
		drawRect(buf, 3, 11, 1, 3, p.shadow);
		drawRect(buf, 2, 8, 3, 3, p.detail);
		drawRect(buf, 11, 10, 1, 3, p.shadow);
		drawRect(buf, 10, 7, 3, 3, p.detail);
	}
	if (variant === 3) {
		// Grass tufts only
		scatterPixels(buf, p.highlight, 0.08, variant * 333);
	}
	return buf;
}

function desertGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.highlight, 0.06, variant * 101);
	scatterPixels(buf, p.secondary, 0.04, variant * 202);

	if (variant === 0) {
		// Small dune ridge
		for (let x = 2; x < 14; x++) {
			setPixel(buf, x, 8, p.shadow);
			setPixel(buf, x, 7, p.secondary);
		}
	}
	if (variant === 1) {
		// Cactus
		drawRect(buf, 7, 4, 2, 8, p.detail);
		drawRect(buf, 5, 6, 2, 1, p.detail);
		drawRect(buf, 5, 6, 1, 3, p.detail);
		drawRect(buf, 10, 5, 1, 3, p.detail);
		drawRect(buf, 9, 5, 2, 1, p.detail);
	}
	if (variant === 2) {
		// Small rocks
		setPixel(buf, 4, 10, p.shadow);
		setPixel(buf, 5, 10, p.shadow);
		setPixel(buf, 11, 6, p.shadow);
	}
	return buf;
}

function volcanicGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.08, variant * 131);

	// Lava crack lines
	if (variant === 0) {
		drawLine(buf, 2, 3, 13, 12, p.detail);
		drawLine(buf, 3, 3, 14, 12, p.highlight);
	}
	if (variant === 1) {
		drawLine(buf, 1, 10, 8, 4, p.detail);
		drawLine(buf, 8, 4, 14, 8, p.detail);
	}
	if (variant === 2) {
		// Lava pool
		drawCircleFilled(buf, 8, 8, 3, p.detail);
		drawCircleFilled(buf, 8, 8, 1, p.highlight);
	}
	if (variant === 3) {
		scatterPixels(buf, p.detail, 0.04, variant * 444);
		scatterPixels(buf, p.highlight, 0.02, variant * 555);
	}
	return buf;
}

function iceGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.highlight, 0.2, variant * 151);
	scatterPixels(buf, p.secondary, 0.08, variant * 252);

	// Crystal V-shapes
	if (variant === 0 || variant === 2) {
		setPixel(buf, 7, 5, p.highlight);
		setPixel(buf, 6, 6, p.highlight);
		setPixel(buf, 8, 6, p.highlight);
		setPixel(buf, 5, 7, p.detail);
		setPixel(buf, 9, 7, p.detail);
	}
	if (variant === 1) {
		// Snowdrift
		drawRect(buf, 3, 10, 10, 2, p.highlight);
		drawRect(buf, 5, 9, 6, 1, p.highlight);
	}
	if (variant === 3) {
		// Ice crack
		drawLine(buf, 2, 2, 13, 13, p.shadow);
	}
	return buf;
}

function savannaGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.06, variant * 171);

	// Sparse grass
	if (variant === 0 || variant === 2) {
		setPixel(buf, 4, 12, p.detail);
		setPixel(buf, 4, 11, p.detail);
		setPixel(buf, 10, 11, p.detail);
		setPixel(buf, 10, 10, p.detail);
	}
	if (variant === 1) {
		// Acacia tree silhouette
		drawRect(buf, 7, 8, 2, 5, p.shadow);
		drawRect(buf, 3, 5, 10, 2, p.detail);
		drawRect(buf, 4, 4, 8, 1, p.detail);
		drawRect(buf, 5, 3, 6, 1, p.shadow);
	}
	if (variant === 3) {
		scatterPixels(buf, p.highlight, 0.04, variant * 373);
	}
	return buf;
}

function badlandsGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);

	// Horizontal rock stripes
	for (let y = 0; y < 16; y += 3) {
		const stripeColor = y % 6 === 0 ? p.secondary : p.detail;
		for (let x = 0; x < 16; x++) {
			setPixel(buf, x, y, stripeColor);
		}
	}

	if (variant === 0) {
		scatterPixels(buf, p.shadow, 0.03, 191);
	}
	if (variant === 1) {
		// Crag
		drawRect(buf, 6, 2, 4, 6, p.shadow);
		drawRect(buf, 7, 1, 2, 1, p.shadow);
	}
	if (variant === 2) {
		scatterPixels(buf, p.highlight, 0.04, 292);
	}
	return buf;
}

function swampGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.1, variant * 211);

	// Water patches
	if (variant === 0 || variant === 3) {
		drawRect(buf, 3, 9, 4, 3, 0x2e5e4e, 180);
		drawRect(buf, 10, 4, 3, 4, 0x2e5e4e, 180);
	}
	if (variant === 1) {
		// Dead stump
		drawRect(buf, 7, 8, 2, 4, p.shadow);
		drawRect(buf, 6, 7, 4, 1, p.shadow);
		setPixel(buf, 5, 6, p.shadow);
		setPixel(buf, 10, 6, p.shadow);
	}
	if (variant === 2) {
		// Lily pads on water
		drawRect(buf, 2, 6, 5, 4, 0x2e5e4e, 180);
		drawCircleFilled(buf, 4, 8, 1, p.highlight);
	}
	return buf;
}

function crystalGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.highlight, 0.1, variant * 231);

	// Diamond crystal shapes
	if (variant === 0 || variant === 2) {
		drawDiamond(buf, 7, 7, 3, p.detail);
		drawDiamond(buf, 7, 7, 1, p.highlight);
	}
	if (variant === 1) {
		// Crystal cluster
		drawDiamond(buf, 4, 6, 2, p.detail);
		drawDiamond(buf, 10, 8, 2, p.secondary);
		setPixel(buf, 4, 6, p.highlight);
		setPixel(buf, 10, 8, p.highlight);
	}
	if (variant === 3) {
		// Sparkle dots
		scatterPixels(buf, p.detail, 0.06, 333);
		setPixel(buf, 3, 3, p.highlight);
		setPixel(buf, 12, 5, p.highlight);
		setPixel(buf, 7, 12, p.highlight);
	}
	return buf;
}

function voidGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.08, variant * 251);

	// Bright rift lines
	if (variant === 0) {
		drawLine(buf, 0, 8, 15, 6, p.detail);
		drawLine(buf, 0, 9, 15, 7, p.highlight, 160);
	}
	if (variant === 1) {
		drawLine(buf, 4, 0, 12, 15, p.detail);
	}
	if (variant === 2) {
		// Void portal
		drawCircleFilled(buf, 8, 8, 3, p.shadow);
		setPixel(buf, 8, 5, p.highlight);
		setPixel(buf, 8, 11, p.highlight);
		setPixel(buf, 5, 8, p.highlight);
		setPixel(buf, 11, 8, p.highlight);
	}
	if (variant === 3) {
		scatterPixels(buf, p.detail, 0.03, 353);
		scatterPixels(buf, p.highlight, 0.015, 454);
	}
	return buf;
}

function neonGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.06, variant * 271);

	// Neon vine curves
	if (variant === 0 || variant === 2) {
		for (let x = 2; x < 14; x++) {
			const y = 8 + Math.round(Math.sin(x * 0.8) * 2);
			setPixel(buf, x, y, p.detail);
			setPixel(buf, x, y + 1, p.detail, 120);
		}
	}
	if (variant === 1) {
		// Glowing mushroom
		drawRect(buf, 7, 9, 2, 4, p.secondary);
		drawCircleFilled(buf, 8, 8, 3, p.detail, 180);
		setPixel(buf, 8, 6, p.highlight);
	}
	if (variant === 3) {
		// Scattered glow dots
		scatterPixels(buf, p.detail, 0.05, 373);
		scatterPixels(buf, p.highlight, 0.02, 474);
	}
	return buf;
}

function plasmaGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.1, variant * 291);

	// Energy crackles
	if (variant === 0) {
		drawLine(buf, 2, 4, 7, 8, p.detail);
		drawLine(buf, 7, 8, 5, 12, p.detail);
		drawLine(buf, 7, 8, 13, 6, p.highlight);
	}
	if (variant === 1) {
		drawLine(buf, 10, 2, 6, 7, p.highlight);
		drawLine(buf, 6, 7, 12, 13, p.detail);
	}
	if (variant === 2) {
		// Plasma orb
		drawCircleFilled(buf, 8, 8, 3, p.detail);
		drawCircleFilled(buf, 8, 8, 1, p.highlight);
	}
	if (variant === 3) {
		scatterPixels(buf, p.detail, 0.06, 393);
		scatterPixels(buf, p.highlight, 0.03, 494);
	}
	return buf;
}

function ancientGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.05, variant * 311);

	if (variant === 0) {
		// Ruin pillars
		drawRect(buf, 3, 3, 3, 10, p.detail);
		drawRect(buf, 10, 5, 3, 8, p.detail);
		// Pillar caps
		drawRect(buf, 2, 2, 5, 1, p.highlight);
		drawRect(buf, 9, 4, 5, 1, p.highlight);
	}
	if (variant === 1) {
		// Brick pattern
		for (let y = 2; y < 14; y += 3) {
			for (let x = 0; x < 16; x += 5) {
				const offset = (y % 6 === 2) ? 0 : 2;
				drawRect(buf, x + offset, y, 4, 2, p.detail);
			}
		}
	}
	if (variant === 2) {
		// Rune markings
		drawRect(buf, 6, 4, 4, 8, p.shadow);
		setPixel(buf, 7, 5, p.highlight);
		setPixel(buf, 8, 7, p.highlight);
		setPixel(buf, 7, 9, p.highlight);
		setPixel(buf, 8, 11, p.highlight);
	}
	if (variant === 3) {
		// Scattered gold dust
		scatterPixels(buf, p.highlight, 0.04, 413);
		scatterPixels(buf, p.detail, 0.03, 514);
	}
	return buf;
}

function oceanGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.08, variant * 331);

	// Wave patterns
	if (variant === 0 || variant === 2) {
		for (let x = 0; x < 16; x++) {
			const y1 = 5 + Math.round(Math.sin(x * 0.6) * 1.5);
			const y2 = 11 + Math.round(Math.sin(x * 0.6 + 2) * 1.5);
			setPixel(buf, x, y1, p.highlight, 180);
			setPixel(buf, x, y2, p.highlight, 180);
		}
	}
	if (variant === 1) {
		// Foam patches
		scatterPixels(buf, p.highlight, 0.08, 331);
		drawRect(buf, 4, 6, 5, 1, p.highlight, 160);
		drawRect(buf, 8, 11, 4, 1, p.highlight, 160);
	}
	if (variant === 3) {
		// Deep dark variant
		scatterPixels(buf, p.shadow, 0.08, 433);
		setPixel(buf, 6, 7, p.detail);
		setPixel(buf, 10, 10, p.detail);
	}
	return buf;
}

function mountainGen(variant: number, p: BiomePalette): Uint8Array {
	const buf = createBuffer();
	fillBase(buf, p.base);
	scatterPixels(buf, p.secondary, 0.06, variant * 351);

	if (variant === 0) {
		// Peak triangle
		for (let row = 0; row < 10; row++) {
			const halfWidth = row;
			const y = 3 + row;
			for (let dx = -halfWidth; dx <= halfWidth; dx++) {
				setPixel(buf, 8 + dx, y, p.detail);
			}
		}
		// Snow cap
		setPixel(buf, 8, 3, p.highlight);
		setPixel(buf, 7, 4, p.highlight);
		setPixel(buf, 8, 4, p.highlight);
		setPixel(buf, 9, 4, p.highlight);
	}
	if (variant === 1) {
		// Crag shapes
		drawRect(buf, 4, 6, 3, 7, p.detail);
		drawRect(buf, 9, 4, 4, 9, p.detail);
		drawRect(buf, 10, 3, 2, 1, p.shadow);
	}
	if (variant === 2) {
		// Snow cap variant
		fillBase(buf, p.secondary);
		for (let y = 0; y < 6; y++) {
			for (let x = 0; x < 16; x++) {
				setPixel(buf, x, y, p.highlight);
			}
		}
		scatterPixels(buf, p.highlight, 0.1, 352);
	}
	if (variant === 3) {
		// Rocky rubble
		scatterPixels(buf, p.detail, 0.1, 453);
		scatterPixels(buf, p.shadow, 0.06, 554);
	}
	return buf;
}

export const BIOME_GENERATORS: Record<BiomeType, BiomeGenFn> = {
	forest: forestGen,
	desert: desertGen,
	volcanic: volcanicGen,
	ice: iceGen,
	savanna: savannaGen,
	badlands: badlandsGen,
	swamp: swampGen,
	crystal: crystalGen,
	void: voidGen,
	neon: neonGen,
	plasma: plasmaGen,
	ancient: ancientGen,
	ocean: oceanGen,
	mountain: mountainGen
};
