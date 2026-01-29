import { writable } from 'svelte/store';
import type { Position } from '$lib/types';

export interface TileInspection {
	position: Position;
	screenX: number;
	screenY: number;
}

export const inspectedTile = writable<TileInspection | null>(null);
export const hudVisible = writable<boolean>(true);

export function inspectTile(position: Position, screenX: number, screenY: number) {
	inspectedTile.set({ position, screenX, screenY });
}

export function clearInspection() {
	inspectedTile.set(null);
}

export function toggleHud() {
	hudVisible.update(v => !v);
}
