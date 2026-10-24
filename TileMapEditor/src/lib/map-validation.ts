import type { GameMap, LayerData, ObjectData, RoomData, TileDefs } from './map-model/types';
import { COLLISION_TYPES, FORMAT_VERSION, SUPPORTED_OBJECT_TYPES } from './map-model/types';

export type IssueSeverity = 'error' | 'warning';

export interface ValidationIssue {
	severity: IssueSeverity;
	path: string;
	message: string;
}

export const MAX_MAP_DIMENSION = 1024;
export const MAX_TILE_SIZE = 512;

const HEX_COLOR = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/;

export function isValidHexColor(value: unknown): value is string {
	return typeof value === 'string' && HEX_COLOR.test(value);
}

export function isPositiveInteger(value: unknown): value is number {
	return typeof value === 'number' && Number.isInteger(value) && value > 0;
}

function isNonNegativeInteger(value: unknown): value is number {
	return typeof value === 'number' && Number.isInteger(value) && value >= 0;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Mirrors the checks performed by the Go loader's ValidateMap, plus the
 * editor-specific checks required by the product spec (spawn bounds, object
 * types the loader can build). Only errors block export.
 */
export function validateMap(map: GameMap): ValidationIssue[] {
	const issues: ValidationIssue[] = [];
	const error = (path: string, message: string) =>
		issues.push({ severity: 'error', path, message });
	const warning = (path: string, message: string) =>
		issues.push({ severity: 'warning', path, message });

	if (map.format_version !== FORMAT_VERSION) {
		error('format_version', `must be ${FORMAT_VERSION} for this game build`);
	}
	if (typeof map.id !== 'string' || map.id.trim() === '') {
		error('id', 'map ID is required');
	}
	if (typeof map.name !== 'string' || map.name.trim() === '') {
		error('name', 'map name is required');
	}
	if (!isPositiveInteger(map.tile_size)) {
		error('tile_size', 'tile size must be a positive integer');
	} else if (map.tile_size > MAX_TILE_SIZE) {
		error('tile_size', `tile size must be at most ${MAX_TILE_SIZE}`);
	}
	if (!isPositiveInteger(map.width)) {
		error('width', 'map width must be a positive integer');
	} else if (map.width > MAX_MAP_DIMENSION) {
		error('width', `map width must be at most ${MAX_MAP_DIMENSION}`);
	}
	if (!isPositiveInteger(map.height)) {
		error('height', 'map height must be a positive integer');
	} else if (map.height > MAX_MAP_DIMENSION) {
		error('height', `map height must be at most ${MAX_MAP_DIMENSION}`);
	}

	validateTiles(map.tiles, error);

	if (!Array.isArray(map.layers) || map.layers.length === 0) {
		error('layers', 'map must have at least one layer');
	} else {
		validateLayers(map, error, warning);
	}

	validateSpawn(map, error);
	validateRooms(map, error);
	validateObjects(map, error);

	return issues;
}

function validateTiles(tiles: TileDefs, error: (path: string, message: string) => void): void {
	if (!isPlainObject(tiles)) {
		error('tiles', 'tiles must be an object keyed by tile ID');
		return;
	}

	const entries = Object.entries(tiles);
	if (entries.length === 0) {
		error('tiles', 'map must define at least one tile');
		return;
	}

	for (const [id, tile] of entries) {
		if (id.trim() === '') {
			error('tiles', 'tile IDs must not be empty');
			continue;
		}
		if (!isPlainObject(tile)) {
			error(`tiles.${id}`, 'tile definition must be an object');
			continue;
		}
		if (typeof tile.solid !== 'boolean') {
			error(`tiles.${id}.solid`, 'solid must be true or false');
		}
		if (!isPlainObject(tile.visual)) {
			error(`tiles.${id}.visual`, 'visual must be an object');
		} else if (!isValidHexColor(tile.visual.color)) {
			error(`tiles.${id}.visual.color`, 'color must be #rgb, #rrggbb, or #rrggbbaa');
		}
	}
}

function validateLayers(
	map: GameMap,
	error: (path: string, message: string) => void,
	warning: (path: string, message: string) => void
): void {
	const layerIDs = new Set<string>();
	const zIndexes = new Map<number, string>();

	for (const [index, layer] of map.layers.entries()) {
		const path = `layers[${index}]`;
		validateLayer(map, layer, path, layerIDs, error);

		if (isNonNegativeInteger(layer.z_index)) {
			const existing = zIndexes.get(layer.z_index);
			if (existing) {
				warning(path + '.z_index', `shares z-index ${layer.z_index} with layer "${existing}"`);
			} else {
				zIndexes.set(layer.z_index, layer.id);
			}
		}
	}
}

function validateLayer(
	map: GameMap,
	layer: LayerData,
	path: string,
	layerIDs: Set<string>,
	error: (path: string, message: string) => void
): void {
	if (typeof layer.id !== 'string' || layer.id.trim() === '') {
		error(path + '.id', 'layer ID is required');
	} else if (layerIDs.has(layer.id)) {
		error(path + '.id', `duplicate layer ID "${layer.id}"`);
	} else {
		layerIDs.add(layer.id);
	}

	if (layer.type !== 'tile') {
		error(path + '.type', 'layer type must be "tile"');
	}
	if (!isNonNegativeInteger(layer.z_index)) {
		error(path + '.z_index', 'z-index must be a non-negative integer');
	}
	if (!(COLLISION_TYPES as readonly string[]).includes(layer.collision)) {
		error(path + '.collision', `collision must be one of: ${COLLISION_TYPES.join(', ')}`);
	}

	if (!Array.isArray(layer.grid)) {
		error(path + '.grid', 'grid must be a two-dimensional array');
		return;
	}
	if (layer.grid.length !== map.height) {
		error(path + '.grid', `grid has ${layer.grid.length} rows; map height is ${map.height}`);
	}

	for (let y = 0; y < layer.grid.length; y += 1) {
		const row = layer.grid[y];
		if (!Array.isArray(row)) {
			error(`${path}.grid[${y}]`, 'grid row must be an array');
			continue;
		}
		if (row.length !== map.width) {
			error(`${path}.grid[${y}]`, `row has ${row.length} cells; map width is ${map.width}`);
		}
		for (let x = 0; x < row.length; x += 1) {
			const tileID = row[x];
			if (typeof tileID !== 'string') {
				error(`${path}.grid[${y}][${x}]`, 'cell must be a tile ID string or ""');
				continue;
			}
			if (tileID === '') continue;
			if (!Object.prototype.hasOwnProperty.call(map.tiles, tileID)) {
				error(`${path}.grid[${y}][${x}]`, `references undefined tile "${tileID}"`);
			}
		}
	}
}

function validateSpawn(map: GameMap, error: (path: string, message: string) => void): void {
	if (!isPlainObject(map.spawn)) {
		error('spawn', 'spawn must be an object with x and y');
		return;
	}
	const { x, y } = map.spawn;
	if (!Number.isInteger(x) || !Number.isInteger(y)) {
		error('spawn', 'spawn x and y must be integers');
		return;
	}
	if (x < 0 || y < 0 || x >= map.width || y >= map.height) {
		error('spawn', `spawn (${x}, ${y}) must be inside the ${map.width}x${map.height} map`);
	}
}

function validateRooms(map: GameMap, error: (path: string, message: string) => void): void {
	if (!Array.isArray(map.rooms)) {
		error('rooms', 'rooms must be an array');
		return;
	}

	const roomIDs = new Set<string>();
	for (const [index, room] of map.rooms.entries()) {
		const path = `rooms[${index}]`;
		validateRoom(map, room, path, roomIDs, error);
	}
}

function validateRoom(
	map: GameMap,
	room: RoomData,
	path: string,
	roomIDs: Set<string>,
	error: (path: string, message: string) => void
): void {
	if (!isPlainObject(room)) {
		error(path, 'room must be an object');
		return;
	}
	if (typeof room.id !== 'string' || room.id.trim() === '') {
		error(path + '.id', 'room ID is required');
	} else if (roomIDs.has(room.id)) {
		error(path + '.id', `duplicate room ID "${room.id}"`);
	} else {
		roomIDs.add(room.id);
	}
	if (typeof room.name !== 'string' || room.name.trim() === '') {
		error(path + '.name', 'room name is required');
	}
	if (!isPlainObject(room.bounds)) {
		error(path + '.bounds', 'room bounds must be an object');
		return;
	}

	const { x, y, width, height } = room.bounds;
	if (!Number.isInteger(x) || !Number.isInteger(y)) {
		error(path + '.bounds', 'bounds x and y must be integers');
		return;
	}
	// The Go validator requires x and y to be strictly greater than zero.
	if (x <= 0 || y <= 0) {
		error(path + '.bounds', 'bounds x and y must be greater than 0');
	}
	if (!isPositiveInteger(width) || !isPositiveInteger(height)) {
		error(path + '.bounds', 'bounds width and height must be positive integers');
		return;
	}
	if (x + width > map.width || y + height > map.height) {
		error(path + '.bounds', 'room must fit inside the map bounds');
	}
}

function validateObjects(map: GameMap, error: (path: string, message: string) => void): void {
	if (!Array.isArray(map.objects)) {
		error('objects', 'objects must be an array');
		return;
	}

	const objectIDs = new Set<string>();
	const roomIDs = new Set(map.rooms.map((room) => room.id));
	const layerIDs = new Set(map.layers.map((layer) => layer.id));

	for (const [index, object] of map.objects.entries()) {
		const path = `objects[${index}]`;
		validateObject(map, object, path, objectIDs, roomIDs, layerIDs, error);
	}
}

function validateObject(
	map: GameMap,
	object: ObjectData,
	path: string,
	objectIDs: Set<string>,
	roomIDs: Set<string>,
	layerIDs: Set<string>,
	error: (path: string, message: string) => void
): void {
	if (!isPlainObject(object)) {
		error(path, 'object must be an object');
		return;
	}
	if (typeof object.id !== 'string' || object.id.trim() === '') {
		error(path + '.id', 'object ID is required');
	} else if (objectIDs.has(object.id)) {
		error(path + '.id', `duplicate object ID "${object.id}"`);
	} else {
		objectIDs.add(object.id);
	}
	if (typeof object.type !== 'string' || object.type.trim() === '') {
		error(path + '.type', 'object type is required');
	} else {
		if (!(SUPPORTED_OBJECT_TYPES as readonly string[]).includes(object.type)) {
			error(
				path + '.type',
				`unsupported object type "${object.type}"; the game loader supports: ${SUPPORTED_OBJECT_TYPES.join(', ')}`
			);
		}
		if (!Object.prototype.hasOwnProperty.call(map.tiles, object.type)) {
			error(path + '.type', `object type "${object.type}" must also exist as a tile definition`);
		}
	}
	if (typeof object.layer !== 'string' || object.layer.trim() === '') {
		error(path + '.layer', 'object layer is required');
	} else if (!layerIDs.has(object.layer)) {
		error(path + '.layer', `references undefined layer "${object.layer}"`);
	}
	if (!Number.isInteger(object.x) || !Number.isInteger(object.y)) {
		error(path, 'object x and y must be integers');
	} else if (object.x < 0 || object.y < 0 || object.x >= map.width || object.y >= map.height) {
		error(path, `object (${object.x}, ${object.y}) must be inside the map`);
	}
	if (!isPlainObject(object.properties)) {
		error(path + '.properties', 'properties must be an object');
		return;
	}
	if (!Object.prototype.hasOwnProperty.call(object.properties, 'room_id')) {
		return;
	}
	const roomID = object.properties.room_id;
	if (typeof roomID !== 'string') {
		error(path + '.properties.room_id', 'room_id must be a string');
	} else if (!roomIDs.has(roomID)) {
		error(path + '.properties.room_id', `references undefined room "${roomID}"`);
	}
}

export function hasBlockingErrors(issues: ValidationIssue[]): boolean {
	return issues.some((issue) => issue.severity === 'error');
}
