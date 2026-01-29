import type { Container, Text, Graphics } from 'pixi.js';
import type { BiomeType } from '$lib/types';

// Base tile size - will be scaled for large worlds
export const BASE_TILE_SIZE = 28;

// Biome colors matching the backend biome definitions
export const BIOME_COLORS: Record<BiomeType, number> = {
	// Core biomes
	forest: 0x228b22, // Forest green
	desert: 0xedc9af, // Sandy beige
	volcanic: 0x8b0000, // Dark red
	ice: 0xb0e0e6, // Powder blue
	savanna: 0xbdb76b, // Dark khaki
	badlands: 0xcd853f, // Peru/tan
	swamp: 0x556b2f, // Dark olive green
	crystal: 0xe6e6fa, // Lavender

	// Fantasy/Sci-Fi biomes
	void: 0x191970, // Midnight blue
	neon: 0x39ff14, // Neon green
	plasma: 0xff1493, // Deep pink
	ancient: 0x8b4513, // Saddle brown

	// Barrier biomes
	ocean: 0x1e90ff, // Dodger blue
	mountain: 0x696969 // Dim gray
};

// Legacy terrain colors for backward compatibility
export const TERRAIN_COLORS = {
	plains: 0x7cb342, // Light green
	forest: 0x2e7d32, // Dark green
	mountain: 0x78909c, // Blue-gray
	water: 0x1976d2 // Ocean blue
};

export const COLORS = {
	background: 0x1a1a2e,
	empty: 0x2d3748,
	gridLine: 0x4a5568,
	player: 0xffd700,
	adversary: 0xff6b6b,
	neutral: 0x888888,
	border: 0x6366f1
};

// Fog of war constants
export const FOG_UNEXPLORED = 0x000000;
export const FOG_EXPLORED = 0x000000;
export const FOG_UNEXPLORED_ALPHA = 0.5;
export const FOG_EXPLORED_ALPHA = 0.3;

export interface AgentSprite {
	container: Container;
	targetX: number;
	targetY: number;
	currentX: number;
	currentY: number;
	reasoningText?: Text;
	reasoningBg?: Graphics;
	reasoningExpiry?: number;
}

export interface ViewportBounds {
	minX: number;
	maxX: number;
	minY: number;
	maxY: number;
}

// Generate distinct colors for different agents
export function getAgentColor(agentId: string, isPlayer: boolean): number {
	if (isPlayer) return COLORS.player;

	let hash = 0;
	for (let i = 0; i < agentId.length; i++) {
		hash = agentId.charCodeAt(i) + ((hash << 5) - hash);
	}

	const hue = Math.abs(hash % 360);
	return hslToHex(hue, 70, 50);
}

export function hslToHex(h: number, s: number, l: number): number {
	s /= 100;
	l /= 100;
	const a = s * Math.min(l, 1 - l);
	const f = (n: number) => {
		const k = (n + h / 30) % 12;
		const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
		return Math.round(255 * color);
	};
	return (f(0) << 16) | (f(8) << 8) | f(4);
}

// Debug timing utility - enabled in dev mode
export const DEBUG_PERF = import.meta.env.DEV;
export function perfLog(label: string, startTime: number) {
	if (DEBUG_PERF) {
		const elapsed = performance.now() - startTime;
		if (elapsed > 1) {
			console.log(`[Renderer] ${label}: ${elapsed.toFixed(2)}ms`);
		}
	}
}
