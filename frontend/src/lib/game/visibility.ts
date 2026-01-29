import type { Agent, Position, WorldObject } from '$lib/types';

/**
 * Visibility state for a tile
 */
export enum TileVisibility {
	UNEXPLORED = 0, // Never seen - solid fog
	EXPLORED = 1, // Previously seen - dark overlay
	VISIBLE = 2 // Currently in vision - full brightness
}

/**
 * Calculate the set of tiles currently visible to the player.
 * Vision is a square radius (axis-aligned) centered on the player.
 *
 * Vision Formula: effectiveVision = baseVision(3) + (visionLevel - 1) + beaconBonuses
 *
 * MULTIPLAYER: This function will be called per-player with their specific agent
 *
 * @param playerAgent - The player's agent
 * @param worldObjects - All world objects (for beacon detection)
 * @param baseVisionRadius - Base vision radius from config (default 3)
 * @returns Set of "x,y" keys for visible tiles
 */
export function calculateVisibleTiles(
	playerAgent: Agent | null,
	worldObjects: WorldObject[],
	baseVisionRadius: number = 3
): Set<string> {
	if (!playerAgent) return new Set();

	const visible = new Set<string>();

	// Calculate effective vision: base + level bonus
	const visionBonus = (playerAgent.vision_level || 1) - 1;
	let effectiveVision = baseVisionRadius + visionBonus;

	// Add player's direct vision (circular radius)
	addCircleVision(visible, playerAgent.position, effectiveVision);

	// MULTIPLAYER: Filter beacons by owner_id matching current player
	// Add vision from owned beacons
	const ownedBeacons = worldObjects.filter(
		(obj) =>
			obj.type === 'structure' &&
			obj.structure_type === 'beacon' &&
			obj.owner_id === playerAgent.id
	);

	const beaconVisionBonus = 2; // From config: beacon_vision_bonus: 2

	for (const beacon of ownedBeacons) {
		// Beacon adds vision from its position (circular)
		addCircleVision(visible, beacon.position, beaconVisionBonus);
	}

	return visible;
}

/**
 * Add all tiles within a circular radius to the visible set.
 * Uses Euclidean distance for a round field of view.
 */
function addCircleVision(set: Set<string>, center: Position, radius: number): void {
	const r2 = radius * radius;
	for (let dy = -radius; dy <= radius; dy++) {
		for (let dx = -radius; dx <= radius; dx++) {
			if (dx * dx + dy * dy > r2) continue;
			const x = center.x + dx;
			const y = center.y + dy;
			if (x >= 0 && y >= 0) {
				set.add(`${x},${y}`);
			}
		}
	}
}

/**
 * Update the explored tiles set with newly visible tiles.
 * Returns the updated explored set for persistence.
 *
 * @param visible - Currently visible tiles
 * @param explored - Previously explored tiles (will be mutated)
 * @returns The explored set with new tiles added
 */
export function updateExploredTiles(visible: Set<string>, explored: Set<string>): Set<string> {
	for (const key of visible) {
		explored.add(key);
	}
	return explored;
}

/**
 * Get the visibility state of a specific tile.
 *
 * @param x - Tile X coordinate
 * @param y - Tile Y coordinate
 * @param visible - Set of currently visible tile keys
 * @param explored - Set of previously explored tile keys
 * @returns The visibility state of the tile
 */
export function getTileVisibility(
	x: number,
	y: number,
	visible: Set<string>,
	explored: Set<string>
): TileVisibility {
	const key = `${x},${y}`;
	if (visible.has(key)) {
		return TileVisibility.VISIBLE;
	}
	if (explored.has(key)) {
		return TileVisibility.EXPLORED;
	}
	return TileVisibility.UNEXPLORED;
}
