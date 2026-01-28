// Game types matching backend

export interface Position {
	x: number;
	y: number;
}

export interface Tile {
	x: number;
	y: number;
	owner_id: string | null;
	terrain: 'plains' | 'forest' | 'mountain' | 'water';
}

export interface Agent {
	id: string;
	name: string;
	position: Position;
	is_adversary: boolean;
	adversary_type?: string;
	hp: number;
	max_hp: number;
	energy: number;
	max_energy: number;
	vision_level: number;
	memory_level: number;
	strength_level: number;
	storage_level: number;
	is_dead: boolean;
}

// Item System
export type ItemCategory = 'tool' | 'consumable' | 'material' | 'placeable' | 'equipment';
export type ItemRarity = 'common' | 'uncommon' | 'rare' | 'epic' | 'legendary';

export interface ItemDefinition {
	id: string;
	name: string;
	description: string;
	category: ItemCategory;
	rarity: ItemRarity;
	max_stack: number;
	energy_cost: number;
	craft_cost?: Record<string, number>;
	properties?: Record<string, unknown>;
	consumable: boolean;
	placeable: boolean;
	usable: boolean;
	equippable: boolean;
}

export interface ItemInstance {
	id: string;
	definition_id: string;
	quantity: number;
	durability?: number;
	metadata?: Record<string, unknown>;
}

export interface InventorySlot {
	item?: ItemInstance;
}

export interface InventorySnapshot {
	owner_id: string;
	slots: InventorySlot[];
	max_slots: number;
	weapon?: ItemInstance;
	armor?: ItemInstance;
	trinket?: ItemInstance;
}

// World Objects
export type WorldObjectType = 'dropped_item' | 'structure' | 'resource' | 'interactive';
export type StructureType = 'wall' | 'beacon' | 'trap';
export type ResourceType = 'wood' | 'stone' | 'crystal' | 'herb';
export type InteractiveType = 'shrine' | 'cache' | 'portal' | 'obelisk';

export interface WorldObject {
	id: string;
	type: WorldObjectType;
	position: Position;
	owner_id?: string;

	// Structure fields
	structure_type?: StructureType;
	hp?: number;
	max_hp?: number;
	hidden?: boolean;
	blocks_movement?: boolean;

	// Resource fields
	resource_type?: ResourceType;
	remaining?: number;

	// Interactive fields
	interactive_type?: InteractiveType;
	message?: string;
	activated?: boolean;

	// Dropped item fields
	item?: ItemInstance;
}

// Recipe System
export interface Recipe {
	id: string;
	name: string;
	description?: string;
	result_item: string;
	result_qty: number;
	ingredients: Record<string, number>;
	energy_cost: number;
	can_craft: boolean;
}

export interface GameMessage {
	tick: number;
	from_agent_id: string;
	to_agent_id: string | null;
	content: string;
}

export interface ActionResult {
	agent_id: string;
	action: string;
	success: boolean;
	message?: string;
	old_pos?: Position;
	new_pos?: Position;
	claimed_at?: Position;
	target_id?: string;
	damage_dealt?: number;
	item_id?: string;
	item_quantity?: number;
	harvested?: string;
	crafted?: string;
	placed?: string;
	upgraded?: string;
	new_level?: number;
}

export interface TileChange {
	x: number;
	y: number;
	owner_id: string | null;
}

export interface TickChanges {
	tiles: TileChange[];
	agents: Agent[];
	messages: GameMessage[];
	results: ActionResult[];
	objects_added?: WorldObject[];
	objects_removed?: string[];
	respawned?: string[];
}

export interface TickUpdate {
	type: 'tick';
	tick: number;
	game_id: string;
	changes: TickChanges;
}

export interface WorldSnapshot {
	size: number;
	seed: number;
	tiles: Tile[];
}

export interface FullGameState {
	type: 'full_state';
	game_id: string;
	tick: number;
	status: 'waiting' | 'running' | 'finished';
	world: WorldSnapshot;
	agents: Agent[];
	messages: GameMessage[];
	world_objects?: WorldObject[];
}

export interface GameInfo {
	id: string;
	status: 'waiting' | 'running' | 'finished';
	player_count: number;
	max_players: number;
	tick: number;
}

export interface AdversaryType {
	type: string;
	name: string;
}

export type WebSocketMessage = TickUpdate | FullGameState | { type: 'pong' } | { type: 'game_over'; winner: string; scores: Record<string, number> };
