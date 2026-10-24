export const COLLISION_TYPES = ['none', 'tile', 'object'] as const;
export type CollisionName = (typeof COLLISION_TYPES)[number];

export const SUPPORTED_OBJECT_TYPES = ['door'] as const;
export const FORMAT_VERSION = 1;

export interface TileVisual {
	color: string;
}

export interface TileDef {
	solid: boolean;
	visual: TileVisual;
}

export type TileDefs = Record<string, TileDef>;

/** Empty cells are represented by the empty string, matching the Go loader. */
export type Cell = string;

export interface LayerData {
	id: string;
	type: 'tile';
	z_index: number;
	collision: CollisionName;
	/** Row-major grid: grid[y][x]. */
	grid: Cell[][];
}

export interface Rectangle {
	x: number;
	y: number;
	width: number;
	height: number;
}

export interface RoomData {
	id: string;
	name: string;
	bounds: Rectangle;
}

export interface ObjectData {
	id: string;
	type: string;
	layer: string;
	x: number;
	y: number;
	properties: Record<string, unknown>;
}

export interface Spawn {
	x: number;
	y: number;
}

/**
 * The exact shape expected by the game's Go map loader. Editor-only state
 * (zoom, selection, visibility, tileset image) must never live here.
 */
export interface GameMap {
	format_version: number;
	id: string;
	name: string;
	tile_size: number;
	width: number;
	height: number;
	tiles: TileDefs;
	layers: LayerData[];
	rooms: RoomData[];
	objects: ObjectData[];
	spawn: Spawn;
}

export interface CellPosition {
	x: number;
	y: number;
}

export function isCollisionName(value: string): value is CollisionName {
	return (COLLISION_TYPES as readonly string[]).includes(value);
}
