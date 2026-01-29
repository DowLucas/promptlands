<script lang="ts">
	import type { Position } from '$lib/types';
	import { objectsAtPosition, agents, tileIndex } from '$lib/stores/game';
	import { itemDefinitions, rarityColors } from '$lib/stores/inventory';

	export let position: Position;
	export let screenX: number;
	export let screenY: number;

	// Tile data
	$: key = `${position.x},${position.y}`;
	$: tile = $tileIndex.get(key);
	$: objects = $objectsAtPosition.get(key) || [];
	$: tileAgents = $agents.filter(a => a.position.x === position.x && a.position.y === position.y && !a.is_dead);

	// Biome display name
	$: biomeName = tile?.biome
		? tile.biome.charAt(0).toUpperCase() + tile.biome.slice(1)
		: tile?.terrain
			? tile.terrain.charAt(0).toUpperCase() + tile.terrain.slice(1)
			: 'Unknown';

	// Owner
	$: ownerAgent = tile?.owner_id ? $agents.find(a => a.id === tile!.owner_id) : null;

	// Position the tooltip, offset from click and clamped to viewport
	$: style = (() => {
		const offset = 12;
		let left = screenX + offset;
		let top = screenY + offset;

		// Clamp to avoid overflow (assume max panel width ~240px, height ~300px)
		const maxLeft = window.innerWidth - 260;
		const maxTop = window.innerHeight - 320;
		if (left > maxLeft) left = screenX - 252;
		if (top > maxTop) top = screenY - 312;
		if (left < 4) left = 4;
		if (top < 4) top = 4;

		return `left: ${left}px; top: ${top}px;`;
	})();

	function getItemName(defId: string): string {
		const def = $itemDefinitions.get(defId);
		return def?.name || defId;
	}

	function getItemRarity(defId: string): string {
		const def = $itemDefinitions.get(defId);
		return def?.rarity || 'common';
	}

	function getRarityColor(rarity: string): string {
		return rarityColors[rarity] || '#9ca3af';
	}

	function getHPColor(hp: number, maxHP: number): string {
		const ratio = hp / maxHP;
		if (ratio > 0.6) return '#22c55e';
		if (ratio > 0.3) return '#eab308';
		return '#ef4444';
	}

	function structureLabel(type: string | undefined): string {
		switch (type) {
			case 'wall': return 'Wall';
			case 'beacon': return 'Beacon';
			case 'trap': return 'Trap';
			default: return type || 'Structure';
		}
	}

	function resourceLabel(type: string | undefined): string {
		switch (type) {
			case 'wood': return 'Wood';
			case 'stone': return 'Stone';
			case 'crystal': return 'Crystal';
			case 'herb': return 'Herb';
			default: return type || 'Resource';
		}
	}

	function interactiveLabel(type: string | undefined): string {
		switch (type) {
			case 'shrine': return 'Shrine';
			case 'cache': return 'Cache';
			case 'portal': return 'Portal';
			case 'obelisk': return 'Obelisk';
			default: return type || 'Interactive';
		}
	}

	$: hasContent = tileAgents.length > 0 || objects.length > 0 || ownerAgent;
</script>

<div class="tile-inspector" {style}>
	<div class="header">
		<span class="coords">({position.x}, {position.y})</span>
		<span class="biome">{biomeName}</span>
	</div>

	{#if ownerAgent}
		<div class="section owner-section">
			<span class="owner-label">Owned by</span>
			<span class="owner-name">{ownerAgent.name}</span>
		</div>
	{/if}

	{#if tileAgents.length > 0}
		<div class="section">
			<div class="section-title">Agents</div>
			{#each tileAgents as agent}
				<div class="agent-row">
					<span class="agent-name">{agent.name}</span>
					<div class="hp-bar-mini">
						<div
							class="hp-fill"
							style="width: {(agent.hp / agent.max_hp) * 100}%; background: {getHPColor(agent.hp, agent.max_hp)}"
						></div>
					</div>
					<span class="hp-text">{agent.hp}/{agent.max_hp}</span>
				</div>
			{/each}
		</div>
	{/if}

	{#if objects.length > 0}
		<div class="section">
			<div class="section-title">Objects</div>
			{#each objects as obj}
				<div class="object-row">
					{#if obj.type === 'structure'}
						<span class="obj-label">{structureLabel(obj.structure_type)}</span>
						{#if obj.hp !== undefined && obj.max_hp}
							<div class="hp-bar-mini">
								<div
									class="hp-fill"
									style="width: {(obj.hp / obj.max_hp) * 100}%; background: {getHPColor(obj.hp, obj.max_hp)}"
								></div>
							</div>
							<span class="hp-text">{obj.hp}/{obj.max_hp}</span>
						{/if}
					{:else if obj.type === 'resource'}
						<span class="obj-label">{resourceLabel(obj.resource_type)}</span>
						{#if obj.remaining !== undefined}
							<span class="obj-detail">x{obj.remaining}</span>
						{/if}
					{:else if obj.type === 'interactive'}
						<span class="obj-label">{interactiveLabel(obj.interactive_type)}</span>
						{#if obj.activated}
							<span class="activated-badge">Active</span>
						{/if}
					{:else if obj.type === 'dropped_item' && obj.item}
						{@const rarity = getItemRarity(obj.item.definition_id)}
						<span class="obj-label" style="color: {getRarityColor(rarity)}">{getItemName(obj.item.definition_id)}</span>
						<span class="obj-detail">x{obj.item.quantity}</span>
					{/if}
				</div>
			{/each}
		</div>
	{/if}

	{#if !hasContent}
		<div class="empty-msg">Nothing here</div>
	{/if}
</div>

<style>
	.tile-inspector {
		position: absolute;
		z-index: 30;
		pointer-events: none;
		background: #1f2937;
		border: 1px solid #374151;
		border-radius: 8px;
		padding: 10px 12px;
		min-width: 180px;
		max-width: 240px;
		font-size: 13px;
		color: #e2e8f0;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 8px;
		padding-bottom: 6px;
		border-bottom: 1px solid #374151;
	}

	.coords {
		font-family: monospace;
		color: #9ca3af;
		font-size: 12px;
	}

	.biome {
		color: #a5b4fc;
		font-weight: 600;
		font-size: 12px;
	}

	.section {
		margin-bottom: 8px;
	}

	.section-title {
		font-size: 10px;
		text-transform: uppercase;
		color: #6b7280;
		margin-bottom: 4px;
		letter-spacing: 0.05em;
	}

	.owner-section {
		display: flex;
		gap: 6px;
		align-items: center;
		margin-bottom: 8px;
	}

	.owner-label {
		color: #6b7280;
		font-size: 12px;
	}

	.owner-name {
		color: #fbbf24;
		font-weight: 600;
		font-size: 12px;
	}

	.agent-row {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 2px 0;
	}

	.agent-name {
		flex-shrink: 0;
		font-size: 12px;
		color: #f3f4f6;
		font-weight: 500;
	}

	.hp-bar-mini {
		flex: 1;
		height: 5px;
		background: #374151;
		border-radius: 3px;
		overflow: hidden;
		min-width: 40px;
	}

	.hp-fill {
		height: 100%;
		transition: width 0.3s ease;
	}

	.hp-text {
		font-size: 10px;
		color: #9ca3af;
		font-family: monospace;
		flex-shrink: 0;
	}

	.object-row {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 2px 0;
	}

	.obj-label {
		font-size: 12px;
		color: #d1d5db;
	}

	.obj-detail {
		font-size: 11px;
		color: #9ca3af;
		font-family: monospace;
	}

	.activated-badge {
		font-size: 10px;
		background: #065f46;
		color: #34d399;
		padding: 1px 5px;
		border-radius: 3px;
	}

	.empty-msg {
		color: #6b7280;
		font-size: 12px;
		text-align: center;
		padding: 4px 0;
	}
</style>
