import { SPRITE_RESOLUTION } from '../types';

const S = SPRITE_RESOLUTION;

// RGBA pixel buffer helpers. All drawing operates on a Uint8Array of size S*S*4.

function colorToRGBA(color: number, alpha = 255): [number, number, number, number] {
	return [
		(color >> 16) & 0xff,
		(color >> 8) & 0xff,
		color & 0xff,
		alpha
	];
}

export function createBuffer(): Uint8Array {
	return new Uint8Array(S * S * 4);
}

export function setPixel(buf: Uint8Array, x: number, y: number, color: number, alpha = 255) {
	if (x < 0 || x >= S || y < 0 || y >= S) return;
	const i = (y * S + x) * 4;
	const [r, g, b, a] = colorToRGBA(color, alpha);
	buf[i] = r;
	buf[i + 1] = g;
	buf[i + 2] = b;
	buf[i + 3] = a;
}

export function fillBase(buf: Uint8Array, color: number) {
	const [r, g, b] = colorToRGBA(color);
	for (let i = 0; i < S * S * 4; i += 4) {
		buf[i] = r;
		buf[i + 1] = g;
		buf[i + 2] = b;
		buf[i + 3] = 255;
	}
}

export function scatterPixels(
	buf: Uint8Array,
	color: number,
	density: number,
	seed: number,
	alpha = 255
) {
	let s = seed | 0;
	const next = () => {
		s = (s * 1664525 + 1013904223) | 0;
		return (s >>> 0) / 4294967296;
	};
	for (let y = 0; y < S; y++) {
		for (let x = 0; x < S; x++) {
			if (next() < density) {
				setPixel(buf, x, y, color, alpha);
			}
		}
	}
}

export function drawRect(
	buf: Uint8Array,
	x0: number,
	y0: number,
	w: number,
	h: number,
	color: number,
	alpha = 255
) {
	for (let dy = 0; dy < h; dy++) {
		for (let dx = 0; dx < w; dx++) {
			setPixel(buf, x0 + dx, y0 + dy, color, alpha);
		}
	}
}

export function drawLine(
	buf: Uint8Array,
	x0: number,
	y0: number,
	x1: number,
	y1: number,
	color: number,
	alpha = 255
) {
	// Bresenham's line
	let dx = Math.abs(x1 - x0);
	let dy = Math.abs(y1 - y0);
	const sx = x0 < x1 ? 1 : -1;
	const sy = y0 < y1 ? 1 : -1;
	let err = dx - dy;
	let cx = x0;
	let cy = y0;
	while (true) {
		setPixel(buf, cx, cy, color, alpha);
		if (cx === x1 && cy === y1) break;
		const e2 = 2 * err;
		if (e2 > -dy) { err -= dy; cx += sx; }
		if (e2 < dx) { err += dx; cy += sy; }
	}
}

export function drawCircleFilled(
	buf: Uint8Array,
	cx: number,
	cy: number,
	r: number,
	color: number,
	alpha = 255
) {
	for (let y = -r; y <= r; y++) {
		for (let x = -r; x <= r; x++) {
			if (x * x + y * y <= r * r) {
				setPixel(buf, cx + x, cy + y, color, alpha);
			}
		}
	}
}

export function drawDiamond(
	buf: Uint8Array,
	cx: number,
	cy: number,
	size: number,
	color: number,
	alpha = 255
) {
	for (let y = -size; y <= size; y++) {
		for (let x = -size; x <= size; x++) {
			if (Math.abs(x) + Math.abs(y) <= size) {
				setPixel(buf, cx + x, cy + y, color, alpha);
			}
		}
	}
}
