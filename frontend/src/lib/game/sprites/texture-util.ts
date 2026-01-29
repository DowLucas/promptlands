import { Texture } from 'pixi.js';
import { SPRITE_RESOLUTION } from './types';

const S = SPRITE_RESOLUTION;

/**
 * Convert a raw RGBA pixel buffer (16x16) into a Pixi.js Texture
 * using an OffscreenCanvas as the intermediate resource.
 */
export function pixelDataToTexture(pixelData: Uint8Array): Texture {
	const canvas = new OffscreenCanvas(S, S);
	const ctx = canvas.getContext('2d')!;
	const imageData = new ImageData(S, S);
	imageData.data.set(pixelData);
	ctx.putImageData(imageData, 0, 0);

	return Texture.from({
		resource: canvas as unknown as HTMLCanvasElement,
		scaleMode: 'nearest'
	});
}
