import { createBuffer, setPixel, drawRect, drawCircleFilled, drawDiamond } from '../terrain/patterns';

export function generateShrine(): Uint8Array {
	const buf = createBuffer();
	// Golden altar
	drawRect(buf, 4, 9, 8, 4, 0xdaa520);
	drawRect(buf, 3, 8, 10, 1, 0xffd700);
	drawRect(buf, 5, 13, 6, 1, 0xb8860b);
	// Star on top
	setPixel(buf, 8, 4, 0xffd700);
	setPixel(buf, 7, 5, 0xffd700);
	setPixel(buf, 9, 5, 0xffd700);
	setPixel(buf, 6, 6, 0xffd700);
	setPixel(buf, 10, 6, 0xffd700);
	setPixel(buf, 8, 6, 0xffd700);
	setPixel(buf, 7, 7, 0xffd700);
	setPixel(buf, 9, 7, 0xffd700);
	// Glow
	setPixel(buf, 8, 3, 0xffff00, 180);
	return buf;
}

export function generateCache(): Uint8Array {
	const buf = createBuffer();
	// Treasure chest
	drawRect(buf, 3, 7, 10, 6, 0xdaa520);
	drawRect(buf, 3, 6, 10, 1, 0xb8860b); // lid top
	drawRect(buf, 4, 5, 8, 1, 0xb8860b); // lid curve
	// Metal bands
	drawRect(buf, 3, 9, 10, 1, 0x8b4513);
	// Lock
	setPixel(buf, 8, 9, 0xffd700);
	setPixel(buf, 7, 9, 0xffd700);
	// Highlights
	setPixel(buf, 5, 7, 0xf0d060);
	setPixel(buf, 6, 7, 0xf0d060);
	return buf;
}

export function generatePortal(): Uint8Array {
	const buf = createBuffer();
	// Purple swirl arcs
	drawCircleFilled(buf, 8, 8, 5, 0x9932cc);
	drawCircleFilled(buf, 8, 8, 3, 0x4b0082);
	drawCircleFilled(buf, 8, 8, 1, 0xff00ff);
	// Swirl highlights
	setPixel(buf, 6, 4, 0xff66ff, 200);
	setPixel(buf, 10, 5, 0xff66ff, 200);
	setPixel(buf, 5, 9, 0xff66ff, 200);
	setPixel(buf, 11, 10, 0xff66ff, 200);
	// Outer glow
	setPixel(buf, 8, 2, 0xcc66ff, 120);
	setPixel(buf, 8, 14, 0xcc66ff, 120);
	setPixel(buf, 2, 8, 0xcc66ff, 120);
	setPixel(buf, 14, 8, 0xcc66ff, 120);
	return buf;
}

export function generateObelisk(): Uint8Array {
	const buf = createBuffer();
	// Stone pillar with runes
	drawRect(buf, 6, 2, 4, 12, 0x696969);
	drawRect(buf, 5, 13, 6, 1, 0x555555); // base
	// Tapered top
	drawRect(buf, 7, 1, 2, 1, 0x808080);
	// Rune markings (glowing)
	setPixel(buf, 7, 4, 0x00ffaa);
	setPixel(buf, 8, 6, 0x00ffaa);
	setPixel(buf, 7, 8, 0x00ffaa);
	setPixel(buf, 8, 10, 0x00ffaa);
	// Stone texture
	setPixel(buf, 7, 5, 0x777777);
	setPixel(buf, 9, 7, 0x5a5a5a);
	setPixel(buf, 6, 9, 0x5a5a5a);
	return buf;
}
