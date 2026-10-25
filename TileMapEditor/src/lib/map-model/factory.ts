import type { Cell, CollisionName, GameMap, LayerData, TileDef, TileDefs } from './types';
import { FORMAT_VERSION } from './types';

export const DEFAULT_MAP_WIDTH = 48;
export const DEFAULT_MAP_HEIGHT = 34;
export const DEFAULT_TILE_SIZE = 16;

export function emptyGrid(width: number, height: number): Cell[][] {
	return Array.from({ length: height }, () => Array.from({ length: width }, () => ''));
}

export function createLayer(
	id: string,
	collision: CollisionName,
	width: number,
	height: number,
	zIndex = 0
): LayerData {
	return {
		id,
		type: 'tile',
		z_index: zIndex,
		collision,
		grid: emptyGrid(width, height)
	};
}

export function createTileDef(color = '#7a828c', solid = false): TileDef {
	return { solid, visual: { color } };
}

/** Tiles every new map starts with, matching the hand-authored game maps. */
export const DEFAULT_TILES: TileDefs = {
	void: createTileDef('#101820', true),
	floor: createTileDef('#555B62', false),
	wall: createTileDef('#89929C', true),
	door: createTileDef('#9A7047', true)
};

export function createDefaultTiles(): TileDefs {
	return Object.fromEntries(
		Object.entries(DEFAULT_TILES).map(([id, tile]) => [id, { ...tile, visual: { ...tile.visual } }])
	);
}

export interface NewMapOptions {
	id?: string;
	name?: string;
	width?: number;
	height?: number;
	tileSize?: number;
}

/**
 * A blank map with the game's standard floor, wall, and structure layers, so
 * the editor is immediately usable and exports match existing maps.
 */
export function createEmptyMap(options: NewMapOptions = {}): GameMap {
	const width = options.width ?? DEFAULT_MAP_WIDTH;
	const height = options.height ?? DEFAULT_MAP_HEIGHT;
	const tileSize = options.tileSize ?? DEFAULT_TILE_SIZE;

	return {
		format_version: FORMAT_VERSION,
		id: options.id ?? 'untitled',
		name: options.name ?? 'Untitled Map',
		tile_size: tileSize,
		width,
		height,
		tiles: createDefaultTiles(),
		layers: [
			createLayer('floor', 'none', width, height, 0),
			createLayer('wall', 'tile', width, height, 1),
			createLayer('structure', 'object', width, height, 2)
		],
		objects: [],
		spawn: { x: 0, y: 0 }
	};
}
