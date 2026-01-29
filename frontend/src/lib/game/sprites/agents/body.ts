import { createBuffer, setPixel, drawRect, drawCircleFilled } from '../terrain/patterns';

/**
 * Generate a 16x16 pixel art humanoid for the given body color.
 * isPlayer=true adds a small golden crown.
 */
export function generateAgentBody(color: number, isPlayer: boolean): Uint8Array {
	const buf = createBuffer();

	const dark = darkenColor(color, 0.6);
	const light = lightenColor(color, 0.3);

	// 1px dark outline body silhouette (drawn first, then body on top)
	// Head outline
	drawCircleFilled(buf, 8, 4, 3, 0x222222);
	// Body outline
	drawRect(buf, 5, 7, 7, 6, 0x222222);
	// Legs outline
	drawRect(buf, 5, 13, 3, 3, 0x222222);
	drawRect(buf, 9, 13, 3, 3, 0x222222);

	// Head (4x4 colored circle)
	drawCircleFilled(buf, 8, 4, 2, color);
	setPixel(buf, 7, 3, light); // highlight

	// Eyes
	setPixel(buf, 7, 4, 0xffffff);
	setPixel(buf, 9, 4, 0xffffff);
	setPixel(buf, 7, 5, 0x222222); // pupils
	setPixel(buf, 9, 5, 0x222222);

	// Body / torso (4x6)
	drawRect(buf, 6, 7, 5, 5, color);
	// Shading
	drawRect(buf, 6, 7, 1, 5, dark);
	drawRect(buf, 10, 7, 1, 5, dark);
	setPixel(buf, 7, 8, light);

	// Arms
	drawRect(buf, 5, 8, 1, 4, dark);
	drawRect(buf, 11, 8, 1, 4, dark);

	// Legs
	drawRect(buf, 6, 12, 2, 2, dark);
	drawRect(buf, 9, 12, 2, 2, dark);
	// Feet
	setPixel(buf, 6, 14, 0x333333);
	setPixel(buf, 7, 14, 0x333333);
	setPixel(buf, 9, 14, 0x333333);
	setPixel(buf, 10, 14, 0x333333);

	// Player crown
	if (isPlayer) {
		setPixel(buf, 7, 1, 0xffd700);
		setPixel(buf, 8, 0, 0xffd700);
		setPixel(buf, 9, 1, 0xffd700);
		setPixel(buf, 6, 2, 0xffd700);
		setPixel(buf, 7, 2, 0xffd700);
		setPixel(buf, 8, 2, 0xffd700);
		setPixel(buf, 9, 2, 0xffd700);
		setPixel(buf, 10, 2, 0xffd700);
	}

	return buf;
}

function darkenColor(color: number, factor: number): number {
	const r = Math.round(((color >> 16) & 0xff) * factor);
	const g = Math.round(((color >> 8) & 0xff) * factor);
	const b = Math.round((color & 0xff) * factor);
	return (r << 16) | (g << 8) | b;
}

function lightenColor(color: number, factor: number): number {
	const r = Math.min(255, Math.round(((color >> 16) & 0xff) * (1 + factor)));
	const g = Math.min(255, Math.round(((color >> 8) & 0xff) * (1 + factor)));
	const b = Math.min(255, Math.round((color & 0xff) * (1 + factor)));
	return (r << 16) | (g << 8) | b;
}
