import type { Application, Container, Graphics } from 'pixi.js';
import type { Tile } from '$lib/types';
import { BASE_TILE_SIZE } from './types';

/**
 * Shared mutable state passed by reference to all renderer modules.
 * Avoids prop-drilling 10+ params through every module.
 */
export class RenderContext {
	app!: Application;
	worldContainer!: Container;
	agentContainer!: Container;
	objectContainer!: Container;
	gridGraphics!: Graphics;
	tileGraphics!: Graphics;
	fogGraphics!: Graphics;

	worldSize: number = 20;
	tileSize: number = BASE_TILE_SIZE;
	playerAgentId: string | null = null;
	isLargeMap: boolean = false;
	minZoom: number = 0.5;
	maxZoom: number = 3;

	// Fog of war state
	fogOfWarEnabled: boolean = true;
	visibleTiles: Set<string> = new Set();
	exploredTiles: Set<string> = new Set();

	// Tile data
	tiles: Tile[] = [];
	tileIndex: Map<string, Tile> = new Map();

	// Agent color cache
	agentColors: Map<string, number> = new Map();
}
