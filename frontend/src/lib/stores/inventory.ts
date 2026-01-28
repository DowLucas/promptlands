import { writable, derived } from 'svelte/store';
import type { ItemDefinition, InventorySnapshot, ItemInstance, Recipe } from '$lib/types';

// Item definitions store (loaded once)
export const itemDefinitions = writable<Map<string, ItemDefinition>>(new Map());

// Player's inventory
export const playerInventory = writable<InventorySnapshot | null>(null);

// Available recipes
export const recipes = writable<Recipe[]>([]);

// Actions
export function setItemDefinitions(items: ItemDefinition[]) {
	const map = new Map<string, ItemDefinition>();
	for (const item of items) {
		map.set(item.id, item);
	}
	itemDefinitions.set(map);
}

export function setPlayerInventory(inventory: InventorySnapshot) {
	playerInventory.set(inventory);
}

export function clearPlayerInventory() {
	playerInventory.set(null);
}

export function setRecipes(recipeList: Recipe[]) {
	recipes.set(recipeList);
}

// Derived stores
export const inventoryItems = derived(playerInventory, ($inv) => {
	if (!$inv) return [];
	return $inv.slots.filter(slot => slot.item).map(slot => slot.item!);
});

export const inventoryItemCounts = derived(playerInventory, ($inv) => {
	const counts = new Map<string, number>();
	if (!$inv) return counts;

	for (const slot of $inv.slots) {
		if (slot.item) {
			const current = counts.get(slot.item.definition_id) || 0;
			counts.set(slot.item.definition_id, current + slot.item.quantity);
		}
	}
	return counts;
});

export const equippedItems = derived(playerInventory, ($inv) => {
	if (!$inv) return { weapon: null, armor: null, trinket: null };
	return {
		weapon: $inv.weapon || null,
		armor: $inv.armor || null,
		trinket: $inv.trinket || null
	};
});

export const emptySlotCount = derived(playerInventory, ($inv) => {
	if (!$inv) return 0;
	return $inv.slots.filter(slot => !slot.item).length;
});

// Helper to get item definition
export function getItemDef(id: string): ItemDefinition | undefined {
	let def: ItemDefinition | undefined;
	itemDefinitions.subscribe(defs => {
		def = defs.get(id);
	})();
	return def;
}

// Helper to format item for display
export function formatItem(item: ItemInstance): string {
	const def = getItemDef(item.definition_id);
	const name = def?.name || item.definition_id;
	return `${name} x${item.quantity}`;
}

// Rarity colors
export const rarityColors: Record<string, string> = {
	common: '#9ca3af',
	uncommon: '#22c55e',
	rare: '#3b82f6',
	epic: '#a855f7',
	legendary: '#f59e0b'
};
