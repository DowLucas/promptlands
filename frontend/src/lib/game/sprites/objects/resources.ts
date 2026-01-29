import { createBuffer, setPixel, drawRect, drawCircleFilled } from '../terrain/patterns';

export function generateWood(): Uint8Array {
	const buf = createBuffer();
	// Log cross-section
	drawCircleFilled(buf, 8, 8, 5, 0x8b4513);
	drawCircleFilled(buf, 8, 8, 4, 0xa0522d);
	// Tree rings
	drawCircleFilled(buf, 8, 8, 2, 0x8b4513);
	setPixel(buf, 8, 8, 0xdeb887); // Center pith
	// Bark outline pixels
	setPixel(buf, 8, 3, 0x5c2e0e);
	setPixel(buf, 8, 13, 0x5c2e0e);
	setPixel(buf, 3, 8, 0x5c2e0e);
	setPixel(buf, 13, 8, 0x5c2e0e);
	return buf;
}

export function generateStone(): Uint8Array {
	const buf = createBuffer();
	// Gray rock shape
	drawRect(buf, 4, 5, 8, 7, 0x808080);
	drawRect(buf, 5, 4, 6, 1, 0x808080);
	drawRect(buf, 5, 12, 6, 1, 0x808080);
	// Highlights and shadows
	setPixel(buf, 5, 5, 0xa0a0a0);
	setPixel(buf, 6, 5, 0xa0a0a0);
	setPixel(buf, 10, 10, 0x606060);
	setPixel(buf, 11, 10, 0x606060);
	// Crack
	setPixel(buf, 7, 7, 0x555555);
	setPixel(buf, 8, 8, 0x555555);
	setPixel(buf, 9, 9, 0x555555);
	return buf;
}

export function generateCrystal(): Uint8Array {
	const buf = createBuffer();
	// Purple crystal cluster
	// Main crystal
	drawRect(buf, 7, 3, 3, 10, 0x9966ff);
	drawRect(buf, 8, 2, 1, 1, 0xcc99ff);
	setPixel(buf, 8, 3, 0xcc99ff); // highlight tip
	// Side crystal left
	drawRect(buf, 4, 6, 2, 6, 0x7744cc);
	setPixel(buf, 4, 5, 0xaa77ee);
	// Side crystal right
	drawRect(buf, 11, 5, 2, 7, 0x7744cc);
	setPixel(buf, 11, 4, 0xaa77ee);
	// Sparkle
	setPixel(buf, 8, 5, 0xffffff);
	setPixel(buf, 5, 7, 0xeeddff);
	setPixel(buf, 12, 6, 0xeeddff);
	return buf;
}

export function generateHerb(): Uint8Array {
	const buf = createBuffer();
	// Green plant with stem and leaves
	// Stem
	drawRect(buf, 7, 8, 2, 5, 0x228b22);
	// Leaves
	drawRect(buf, 5, 5, 3, 3, 0x32cd32);
	drawRect(buf, 9, 4, 3, 3, 0x32cd32);
	drawRect(buf, 6, 7, 2, 2, 0x2ea043);
	// Leaf veins
	setPixel(buf, 6, 6, 0x1a6b1a);
	setPixel(buf, 10, 5, 0x1a6b1a);
	// Flower bud
	setPixel(buf, 7, 4, 0xff69b4);
	setPixel(buf, 8, 3, 0xff69b4);
	setPixel(buf, 8, 4, 0xff1493);
	return buf;
}
