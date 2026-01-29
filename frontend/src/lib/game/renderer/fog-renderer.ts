import type { RenderContext } from './render-context';
import { FOG_UNEXPLORED, FOG_EXPLORED, FOG_UNEXPLORED_ALPHA, FOG_EXPLORED_ALPHA } from './types';

export class FogOfWarRenderer {
	private ctx: RenderContext;

	constructor(ctx: RenderContext) {
		this.ctx = ctx;
	}

	/**
	 * Render fog overlay for a single tile.
	 * Called from TileRenderer's tile loop for single-pass performance.
	 */
	renderFogForTile(
		key: string,
		px: number,
		py: number,
		isVisible: boolean,
		isExplored: boolean
	) {
		if (isVisible) return; // No fog for visible tiles

		if (isExplored) {
			this.ctx.fogGraphics.rect(px, py, this.ctx.tileSize, this.ctx.tileSize);
			this.ctx.fogGraphics.fill({ color: FOG_EXPLORED, alpha: FOG_EXPLORED_ALPHA });
		} else {
			this.ctx.fogGraphics.rect(px, py, this.ctx.tileSize, this.ctx.tileSize);
			this.ctx.fogGraphics.fill({ color: FOG_UNEXPLORED, alpha: FOG_UNEXPLORED_ALPHA });
		}
	}

	clear() {
		this.ctx.fogGraphics.clear();
	}

	setFogOfWarEnabled(enabled: boolean) {
		this.ctx.fogOfWarEnabled = enabled;
		if (!enabled) {
			this.ctx.fogGraphics.clear();
		}
	}

	/**
	 * Update visibility state for fog of war.
	 * Returns the set of newly explored tiles (for persisting to store).
	 */
	updateVisibility(visible: Set<string>, explored: Set<string>): Set<string> {
		this.ctx.visibleTiles = visible;

		const newlyExplored = new Set<string>();
		for (const key of visible) {
			if (!this.ctx.exploredTiles.has(key)) {
				newlyExplored.add(key);
			}
			this.ctx.exploredTiles.add(key);
			explored.add(key);
		}

		return newlyExplored;
	}

	setExploredTiles(explored: Set<string>) {
		this.ctx.exploredTiles = new Set(explored);
	}
}
