import { createBuffer, setPixel, drawRect, drawCircleFilled } from '../terrain/patterns';

export function generateDroppedItem(): Uint8Array {
	const buf = createBuffer();
	// Small orange pouch
	drawCircleFilled(buf, 8, 9, 3, 0xffa500);
	drawCircleFilled(buf, 8, 9, 2, 0xe69500);
	// Pouch top / tie
	drawRect(buf, 7, 5, 2, 2, 0x8b4513);
	setPixel(buf, 8, 5, 0xdaa520);
	// String
	setPixel(buf, 6, 6, 0x8b4513);
	setPixel(buf, 10, 6, 0x8b4513);
	// Highlight
	setPixel(buf, 7, 8, 0xffcc66);
	return buf;
}
