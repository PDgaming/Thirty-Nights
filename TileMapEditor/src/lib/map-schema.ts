import type { GameMap } from './map-model/types';
import { serializeMap } from './map-serializer';

/**
 * Top-level fields the editor UI reads and writes. A custom schema must keep
 * these properties, otherwise the exported JSON would not describe the map the
 * user edited.
 */
export const COMPULSORY_MAP_FIELDS = [
	'format_version',
	'id',
	'name',
	'tile_size',
	'width',
	'height',
	'tiles',
	'layers',
	'objects',
	'spawn'
] as const;

export type MapSchema = Record<string, unknown>;

export class MapSchemaError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'MapSchemaError';
	}
}

/**
 * The default export contract. The editor ships with this schema, and it can be
 * replaced from the settings modal as long as the compulsory fields remain.
 */
export const DEFAULT_MAP_SCHEMA: MapSchema = {
	$schema: 'https://json-schema.org/draft/2020-12/schema',
	$id: 'https://thirty-nights.local/schemas/map.schema.json',
	title: 'Thirty Nights tile map',
	type: 'object',
	required: [...COMPULSORY_MAP_FIELDS],
	properties: {
		format_version: { type: 'integer', const: 1 },
		id: { type: 'string', description: 'Stable map identifier used by the game.' },
		name: { type: 'string', description: 'Display name shown in the editor.' },
		tile_size: { type: 'integer', minimum: 1, description: 'Size of one cell in game units.' },
		width: { type: 'integer', minimum: 1, description: 'Map width in cells.' },
		height: { type: 'integer', minimum: 1, description: 'Map height in cells.' },
		tiles: {
			type: 'object',
			description: 'Tile definitions keyed by tile ID.',
			additionalProperties: {
				type: 'object',
				required: ['solid', 'visual'],
				properties: {
					solid: { type: 'boolean' },
					visual: {
						type: 'object',
						required: ['color'],
						properties: { color: { type: 'string' } }
					}
				}
			}
		},
		layers: {
			type: 'array',
			description: 'Tile layers ordered from bottom to top.',
			items: {
				type: 'object',
				required: ['id', 'type', 'z_index', 'collision', 'grid'],
				properties: {
					id: { type: 'string' },
					type: { const: 'tile' },
					z_index: { type: 'integer' },
					collision: { enum: ['none', 'tile', 'object'] },
					grid: {
						type: 'array',
						description: 'Row-major cells: grid[y][x]. Empty cells use "".',
						items: { type: 'array', items: { type: 'string' } }
					}
				}
			}
		},
		objects: {
			type: 'array',
			items: {
				type: 'object',
				required: ['id', 'type', 'layer', 'x', 'y'],
				properties: {
					id: { type: 'string' },
					type: { type: 'string' },
					layer: { type: 'string' },
					x: { type: 'integer' },
					y: { type: 'integer' },
					properties: { type: 'object' }
				}
			}
		},
		spawn: {
			type: 'object',
			required: ['x', 'y'],
			properties: {
				x: { type: 'integer' },
				y: { type: 'integer' }
			}
		}
	},
	additionalProperties: true
};

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function schemaProperties(schema: MapSchema): Record<string, unknown> {
	const properties = schema.properties;
	return isRecord(properties) ? properties : {};
}

function schemaRequired(schema: MapSchema): string[] {
	const required = schema.required;
	if (!Array.isArray(required)) return [];
	return required.filter((name): name is string => typeof name === 'string');
}

/**
 * Basic structural checks only: valid JSON object, a properties object, and the
 * fields the editor depends on. Full JSON Schema validation is intentionally
 * out of scope.
 */
export function validateMapSchema(raw: unknown): asserts raw is MapSchema {
	if (!isRecord(raw)) {
		throw new MapSchemaError('Schema must be a JSON object.');
	}

	const properties = raw.properties;
	if (!isRecord(properties)) {
		throw new MapSchemaError('Schema needs a "properties" object.');
	}
	if (Object.keys(properties).length === 0) {
		throw new MapSchemaError('Schema "properties" must not be empty.');
	}

	if (raw.required !== undefined) {
		if (!Array.isArray(raw.required) || raw.required.some((name) => typeof name !== 'string')) {
			throw new MapSchemaError('Schema "required" must be an array of property names.');
		}
	}

	const required = new Set(schemaRequired(raw));
	for (const field of COMPULSORY_MAP_FIELDS) {
		if (!Object.prototype.hasOwnProperty.call(properties, field)) {
			throw new MapSchemaError(`Schema properties must include the compulsory field "${field}".`);
		}
		if (!required.has(field)) {
			throw new MapSchemaError(`Schema "required" must include the compulsory field "${field}".`);
		}
	}

	for (const name of required) {
		if (!Object.prototype.hasOwnProperty.call(properties, name)) {
			throw new MapSchemaError(`Required property "${name}" is missing from "properties".`);
		}
		const definition = properties[name];
		if (
			!(COMPULSORY_MAP_FIELDS as readonly string[]).includes(name) &&
			(!isRecord(definition) || !Object.prototype.hasOwnProperty.call(definition, 'default'))
		) {
			throw new MapSchemaError(
				`Required property "${name}" needs a "default" value so the editor can export it.`
			);
		}
	}
}

export function parseMapSchema(text: string): MapSchema {
	let raw: unknown;
	try {
		raw = JSON.parse(text);
	} catch (error) {
		throw new MapSchemaError(
			`Schema is not valid JSON: ${error instanceof Error ? error.message : 'parse error'}`
		);
	}
	validateMapSchema(raw);
	return raw;
}

/**
 * Build the export object from the schema's property order. Properties the map
 * model provides use the map's values; extra properties use their schema
 * "default" values.
 */
export function buildExportJson(map: GameMap, schema: MapSchema): Record<string, unknown> {
	const serialized = serializeMap(map) as unknown as Record<string, unknown>;
	const properties = schemaProperties(schema);
	const output: Record<string, unknown> = {};

	for (const key of Object.keys(properties)) {
		if (Object.prototype.hasOwnProperty.call(serialized, key)) {
			output[key] = serialized[key];
			continue;
		}
		const definition = properties[key];
		if (isRecord(definition) && Object.prototype.hasOwnProperty.call(definition, 'default')) {
			output[key] = structuredClone(definition.default);
		}
	}

	// Compulsory fields are always present, even if a schema slipped through.
	for (const key of COMPULSORY_MAP_FIELDS) {
		if (!Object.prototype.hasOwnProperty.call(output, key)) output[key] = serialized[key];
	}

	return output;
}

export function mapToJsonWithSchema(map: GameMap, schema: MapSchema): string {
	return JSON.stringify(buildExportJson(map, schema), null, 2);
}
