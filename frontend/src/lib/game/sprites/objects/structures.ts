import { createBuffer, fillBase, setPixel, drawRect, drawDiamond } from '../terrain/patterns';

export function generateWall(): Uint8Array {
	const buf = createBuffer();
	// Brick pattern
	fillBase(buf, 0x8b4513);
	// Mortar lines (horizontal)
	for (let y = 0; y < 16; y += 4) {
		for (let x = 0; x < 16; x++) {
			setPixel(buf, x, y, 0x5c3317);
		}
	}
	// Mortar lines (vertical, offset per row)
	for (let row = 0; row < 4; row++) {
		const offset = row % 2 === 0 ? 0 : 4;
		const y0 = row * 4;
		for (let x = offset; x < 16; x += 8) {
			for (let dy = 0; dy < 4; dy++) {
				setPixel(buf, x, y0 + dy, 0x5c3317);
			}
		}
	}
	// Highlight bricks
	setPixel(buf, 3, 2, 0xa0603a);
	setPixel(buf, 7, 6, 0xa0603a);
	setPixel(buf, 11, 10, 0xa0603a);
	return buf;
}

export function generateBeacon(): Uint8Array {
	const buf = createBuffer();
	// Transparent background (alpha=0 by default from createBuffer)

	// Cyan crystal diamond
	drawDiamond(buf, 8, 8, 5, 0x00cccc);
	drawDiamond(buf, 8, 8, 3, 0x00ffff);
	drawDiamond(buf, 8, 8, 1, 0xffffff);

	// Glow corners
	setPixel(buf, 8, 2, 0x66ffff, 180);
	setPixel(buf, 8, 14, 0x66ffff, 180);
	setPixel(buf, 2, 8, 0x66ffff, 180);
	setPixel(buf, 14, 8, 0x66ffff, 180);
	return buf;
}

export function generateTrap(): Uint8Array {
	const buf = createBuffer();
	// Subtle spike marks on transparent bg

	// Small spikes pointing up
	setPixel(buf, 4, 9, 0xff3333);
	setPixel(buf, 4, 8, 0xff3333);
	setPixel(buf, 4, 7, 0xff0000);

	setPixel(buf, 8, 9, 0xff3333);
	setPixel(buf, 8, 8, 0xff3333);
	setPixel(buf, 8, 7, 0xff0000);

	setPixel(buf, 12, 9, 0xff3333);
	setPixel(buf, 12, 8, 0xff3333);
	setPixel(buf, 12, 7, 0xff0000);

	// Base line
	for (let x = 3; x < 14; x++) {
		setPixel(buf, x, 10, 0x660000);
	}
	return buf;
}
