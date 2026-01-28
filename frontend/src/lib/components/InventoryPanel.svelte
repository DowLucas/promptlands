<script lang="ts">
	import type { InventorySnapshot, ItemInstance, ItemDefinition } from '$lib/types';
	import { itemDefinitions, rarityColors } from '$lib/stores/inventory';

	export let inventory: InventorySnapshot | null = null;

	let selectedSlot: number | null = null;
	let selectedItem: ItemInstance | null = null;

	function getItemDef(id: string): ItemDefinition | undefined {
		return $itemDefinitions.get(id);
	}

	function selectSlot(index: number, item: ItemInstance | undefined) {
		if (item) {
			selectedSlot = index;
			selectedItem = item;
		} else {
			selectedSlot = null;
			selectedItem = null;
		}
	}

	function getRarityColor(rarity: string): string {
		return rarityColors[rarity] || '#9ca3af';
	}

	function getCategoryIcon(category: string): string {
		switch (category) {
			case 'material':
				return '📦';
			case 'consumable':
				return '🧪';
			case 'placeable':
				return '🏗️';
			case 'equipment':
				return '⚔️';
			case 'tool':
				return '🔧';
			default:
				return '❓';
		}
	}
</script>

<div class="inventory-panel">
	<h3>Inventory</h3>

	{#if inventory}
		<!-- Equipment Slots -->
		<div class="equipment-section">
			<div class="equipment-slot" class:has-item={inventory.weapon}>
				<span class="slot-label">Weapon</span>
				{#if inventory.weapon}
					{@const def = getItemDef(inventory.weapon.definition_id)}
					<div class="equipped-item" style="border-color: {getRarityColor(def?.rarity || 'common')}">
						<span class="item-name">{def?.name || inventory.weapon.definition_id}</span>
					</div>
				{:else}
					<div class="empty-slot">Empty</div>
				{/if}
			</div>

			<div class="equipment-slot" class:has-item={inventory.armor}>
				<span class="slot-label">Armor</span>
				{#if inventory.armor}
					{@const def = getItemDef(inventory.armor.definition_id)}
					<div class="equipped-item" style="border-color: {getRarityColor(def?.rarity || 'common')}">
						<span class="item-name">{def?.name || inventory.armor.definition_id}</span>
					</div>
				{:else}
					<div class="empty-slot">Empty</div>
				{/if}
			</div>

			<div class="equipment-slot" class:has-item={inventory.trinket}>
				<span class="slot-label">Trinket</span>
				{#if inventory.trinket}
					{@const def = getItemDef(inventory.trinket.definition_id)}
					<div class="equipped-item" style="border-color: {getRarityColor(def?.rarity || 'common')}">
						<span class="item-name">{def?.name || inventory.trinket.definition_id}</span>
					</div>
				{:else}
					<div class="empty-slot">Empty</div>
				{/if}
			</div>
		</div>

		<!-- Inventory Grid -->
		<div class="inventory-grid">
			{#each inventory.slots as slot, index}
				<button
					class="inventory-slot"
					class:has-item={slot.item}
					class:selected={selectedSlot === index}
					on:click={() => selectSlot(index, slot.item)}
				>
					{#if slot.item}
						{@const def = getItemDef(slot.item.definition_id)}
						<div class="slot-content" style="border-color: {getRarityColor(def?.rarity || 'common')}">
							<span class="item-icon">{getCategoryIcon(def?.category || 'material')}</span>
							<span class="item-qty">{slot.item.quantity}</span>
						</div>
					{/if}
				</button>
			{/each}
		</div>

		<!-- Item Details -->
		{#if selectedItem}
			{@const def = getItemDef(selectedItem.definition_id)}
			<div class="item-details">
				<div class="detail-header" style="color: {getRarityColor(def?.rarity || 'common')}">
					{def?.name || selectedItem.definition_id}
				</div>
				{#if def}
					<div class="detail-rarity">{def.rarity}</div>
					<div class="detail-desc">{def.description || 'No description'}</div>
					<div class="detail-info">
						<span>Quantity: {selectedItem.quantity}</span>
						{#if def.max_stack > 1}
							<span>Max Stack: {def.max_stack}</span>
						{/if}
						{#if def.energy_cost > 0}
							<span>Energy Cost: {def.energy_cost}</span>
						{/if}
					</div>
					<div class="detail-actions">
						{#if def.usable}
							<span class="action-badge">Usable</span>
						{/if}
						{#if def.placeable}
							<span class="action-badge">Placeable</span>
						{/if}
						{#if def.equippable}
							<span class="action-badge">Equippable</span>
						{/if}
					</div>
				{/if}
			</div>
		{/if}

		<div class="inventory-footer">
			{inventory.slots.filter(s => s.item).length}/{inventory.max_slots} slots used
		</div>
	{:else}
		<div class="no-inventory">No inventory data</div>
	{/if}
</div>

<style>
	.inventory-panel {
		background: #1f2937;
		border: 1px solid #374151;
		border-radius: 8px;
		padding: 12px;
	}

	h3 {
		margin: 0 0 12px 0;
		color: #f3f4f6;
		font-size: 16px;
	}

	.equipment-section {
		display: flex;
		gap: 8px;
		margin-bottom: 12px;
		padding-bottom: 12px;
		border-bottom: 1px solid #374151;
	}

	.equipment-slot {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}

	.slot-label {
		font-size: 10px;
		color: #6b7280;
		text-transform: uppercase;
	}

	.equipped-item {
		width: 100%;
		padding: 8px;
		background: #111827;
		border: 2px solid;
		border-radius: 6px;
		text-align: center;
	}

	.empty-slot {
		width: 100%;
		padding: 8px;
		background: #111827;
		border: 1px dashed #374151;
		border-radius: 6px;
		text-align: center;
		color: #4b5563;
		font-size: 12px;
	}

	.item-name {
		font-size: 11px;
		color: #d1d5db;
	}

	.inventory-grid {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 4px;
		margin-bottom: 12px;
	}

	.inventory-slot {
		aspect-ratio: 1;
		background: #111827;
		border: 1px solid #374151;
		border-radius: 4px;
		padding: 0;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.inventory-slot:hover {
		border-color: #4b5563;
	}

	.inventory-slot.has-item {
		background: #1e293b;
	}

	.inventory-slot.selected {
		border-color: #3b82f6;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
	}

	.slot-content {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		border: 2px solid;
		border-radius: 4px;
		position: relative;
	}

	.item-icon {
		font-size: 18px;
	}

	.item-qty {
		position: absolute;
		bottom: 2px;
		right: 4px;
		font-size: 10px;
		font-weight: 600;
		color: #fff;
		text-shadow: 0 0 2px #000;
	}

	.item-details {
		background: #111827;
		border-radius: 6px;
		padding: 12px;
		margin-bottom: 12px;
	}

	.detail-header {
		font-weight: 600;
		font-size: 14px;
		margin-bottom: 4px;
	}

	.detail-rarity {
		font-size: 11px;
		color: #6b7280;
		text-transform: capitalize;
		margin-bottom: 8px;
	}

	.detail-desc {
		font-size: 12px;
		color: #9ca3af;
		margin-bottom: 8px;
		line-height: 1.4;
	}

	.detail-info {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		font-size: 11px;
		color: #6b7280;
		margin-bottom: 8px;
	}

	.detail-actions {
		display: flex;
		gap: 4px;
	}

	.action-badge {
		background: #374151;
		color: #d1d5db;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 10px;
	}

	.inventory-footer {
		text-align: center;
		color: #6b7280;
		font-size: 12px;
	}

	.no-inventory {
		color: #6b7280;
		text-align: center;
		padding: 24px;
	}
</style>
