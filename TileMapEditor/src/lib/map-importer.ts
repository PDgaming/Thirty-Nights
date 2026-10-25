import type {
	Cell,
	CollisionName,
	GameMap,
	LayerData,
	ObjectData,
	Spawn,
	TileDef,
	TileDefs
} from './map-model/types';
import { COLLISION_TYPES, FORMAT_VERSION } from './map-model/types';
import { MAX_MAP_DIMENSION, MAX_TILE_SIZE } from './map-validation';

export class MapImportError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'MapImportError';
	}
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function fail(path: string, message: string): never {
	throw new MapImportError(`${path}: ${message}`);
}

function readString(value: unknown, path: string): string {
	if (typeof value !== 'string') fail(path, 'must be a string');
	return value;
}

function readInteger(value: unknown, path: string, minimum?: number): number {
	if (typeof value !== 'number' || !Number.isInteger(value)) {
		fail(path, 'must be an integer');
	}
	if (minimum !== undefined && value < minimum) {
		fail(path, `must be at least ${minimum}`);
	}
	return value;
}

function readPositiveDimension(value: unknown, path: string, maximum: number): number {
	const number = readInteger(value, path, 1);
	if (number > maximum) fail(path, `must be at most ${maximum}`);
	return number;
}

export function parseMapJson(text: string): GameMap {
	const source = text.charCodeAt(0) === 0xfeff ? text.slice(1) : text;
	let raw: unknown;
	try {
		raw = JSON.parse(source);
	} catch (error) {
		throw new MapImportError(
			`Not valid JSON: ${error instanceof Error ? error.message : 'could not parse file'}`
		);
	}
	return parseMapData(raw);
}

/**
 * Normalizes a game map object into the editor's model. Extra fields such as
 * the retired `rooms` array are ignored, and missing optional collections are
 * filled with empty arrays.
 */
export function parseMapData(raw: unknown): GameMap {
	if (!isRecord(raw)) fail('map', 'must be a JSON object');

	const formatVersion = readInteger(raw.format_version, 'format_version', 1);
	if (formatVersion !== FORMAT_VERSION) {
		fail(
			'format_version',
			`${formatVersion} is not supported by this editor (expected ${FORMAT_VERSION})`
		);
	}

	const width = readPositiveDimension(raw.width, 'width', MAX_MAP_DIMENSION);
	const height = readPositiveDimension(raw.height, 'height', MAX_MAP_DIMENSION);
	const tileSize = readPositiveDimension(raw.tile_size, 'tile_size', MAX_TILE_SIZE);

	const tiles = parseTiles(raw.tiles);
	const layers = parseLayers(raw.layers, width, height);
	const objects = parseObjects(raw.objects);
	const spawn = parseSpawn(raw.spawn);

	return {
		format_version: formatVersion,
		id: readString(raw.id, 'id'),
		name: readString(raw.name, 'name'),
		tile_size: tileSize,
		width,
		height,
		tiles,
		layers,
		objects,
		spawn
	};
}

function parseTiles(value: unknown): TileDefs {
	if (!isRecord(value)) fail('tiles', 'must be an object keyed by tile ID');

	const tiles: TileDefs = {};
	for (const [id, rawTile] of Object.entries(value)) {
		const path = `tiles.${id}`;
		if (id.trim() === '') fail(path, 'tile ID must not be empty');
		if (!isRecord(rawTile)) fail(path, 'must be an object');
		if (typeof rawTile.solid !== 'boolean') fail(`${path}.solid`, 'must be true or false');
		if (!isRecord(rawTile.visual)) fail(`${path}.visual`, 'must be an object');

		const tile: TileDef = {
			solid: rawTile.solid,
			visual: { color: readString(rawTile.visual.color, `${path}.visual.color`) }
		};
		tiles[id] = tile;
	}
	return tiles;
}

function parseLayers(value: unknown, width: number, height: number): LayerData[] {
	if (!Array.isArray(value)) fail('layers', 'must be an array');
	if (value.length === 0) fail('layers', 'map must contain at least one layer');

	return value.map((rawLayer, index) => {
		const path = `layers[${index}]`;
		if (!isRecord(rawLayer)) fail(path, 'must be an object');
		if (rawLayer.type !== 'tile') fail(`${path}.type`, 'must be "tile"');

		const collision = readString(rawLayer.collision, `${path}.collision`);
		if (!(COLLISION_TYPES as readonly string[]).includes(collision)) {
			fail(`${path}.collision`, `must be one of: ${COLLISION_TYPES.join(', ')}`);
		}

		return {
			id: readString(rawLayer.id, `${path}.id`),
			type: 'tile',
			z_index: readInteger(rawLayer.z_index, `${path}.z_index`),
			collision: collision as CollisionName,
			grid: parseGrid(rawLayer.grid, width, height, `${path}.grid`)
		};
	});
}

function parseGrid(value: unknown, width: number, height: number, path: string): Cell[][] {
	if (!Array.isArray(value)) fail(path, 'must be a two-dimensional array');
	if (value.length !== height) {
		fail(path, `has ${value.length} rows but the map height is ${height}`);
	}

	return value.map((rawRow, y) => {
		const rowPath = `${path}[${y}]`;
		if (!Array.isArray(rawRow)) fail(rowPath, 'must be an array');
		if (rawRow.length !== width) {
			fail(rowPath, `has ${rawRow.length} cells but the map width is ${width}`);
		}
		return rawRow.map((cell, x) => readString(cell, `${rowPath}[${x}]`));
	});
}

function parseObjects(value: unknown): ObjectData[] {
	if (value === undefined) return [];
	if (!Array.isArray(value)) fail('objects', 'must be an array');

	return value.map((rawObject, index) => {
		const path = `objects[${index}]`;
		if (!isRecord(rawObject)) fail(path, 'must be an object');

		let properties: Record<string, unknown> = {};
		if (rawObject.properties !== undefined) {
			if (!isRecord(rawObject.properties)) fail(`${path}.properties`, 'must be an object');
			properties = { ...rawObject.properties };
		}

		return {
			id: readString(rawObject.id, `${path}.id`),
			type: readString(rawObject.type, `${path}.type`),
			layer: readString(rawObject.layer, `${path}.layer`),
			x: readInteger(rawObject.x, `${path}.x`),
			y: readInteger(rawObject.y, `${path}.y`),
			properties
		};
	});
}

function parseSpawn(value: unknown): Spawn {
	if (value === undefined) return { x: 0, y: 0 };
	if (!isRecord(value)) fail('spawn', 'must be an object');

	return {
		x: readInteger(value.x, 'spawn.x'),
		y: readInteger(value.y, 'spawn.y')
	};
}
