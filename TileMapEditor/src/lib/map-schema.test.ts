import { describe, expect, it } from 'vitest';
import { createEmptyMap, createTileDef } from './map-model/factory';
import {
	buildExportJson,
	DEFAULT_MAP_SCHEMA,
	mapToJsonWithSchema,
	MapSchemaError,
	parseMapSchema
} from './map-schema';

function map() {
	const result = createEmptyMap({ width: 3, height: 2, id: 'test-map', name: 'Test Map' });
	result.tiles = { floor: createTileDef('#555B62') };
	result.layers[0].grid[0][1] = 'floor';
	result.spawn = { x: 1, y: 1 };
	return result;
}

function schemaWith(extra: Record<string, unknown>) {
	return {
		...DEFAULT_MAP_SCHEMA,
		properties: {
			...((DEFAULT_MAP_SCHEMA.properties as Record<string, unknown>) ?? {}),
			...extra
		}
	};
}

describe('map schema', () => {
	it('accepts the default schema', () => {
		expect(() => parseMapSchema(JSON.stringify(DEFAULT_MAP_SCHEMA))).not.toThrow();
	});

	it('rejects schemas missing a compulsory property', () => {
		const properties = { ...(DEFAULT_MAP_SCHEMA.properties as Record<string, unknown>) };
		delete properties.width;
		const schema = { ...DEFAULT_MAP_SCHEMA, properties };

		expect(() => parseMapSchema(JSON.stringify(schema))).toThrow(MapSchemaError);
		expect(() => parseMapSchema(JSON.stringify(schema))).toThrow(/width/);
	});

	it('rejects schemas that drop a compulsory field from required', () => {
		const schema = {
			...DEFAULT_MAP_SCHEMA,
			required: (DEFAULT_MAP_SCHEMA.required as string[]).filter((name) => name !== 'height')
		};

		expect(() => parseMapSchema(JSON.stringify(schema))).toThrow(/height/);
	});

	it('rejects extra required fields without a default', () => {
		const schema = {
			...schemaWith({ weather: { type: 'string' } }),
			required: [...(DEFAULT_MAP_SCHEMA.required as string[]), 'weather']
		};

		expect(() => parseMapSchema(JSON.stringify(schema))).toThrow(/weather/);
	});

	it('exports extra properties using their default values', () => {
		const schema = schemaWith({
			weather: { type: 'string', default: 'rain' },
			room_count: { type: 'integer', default: 3 }
		});
		const exported = buildExportJson(map(), schema);

		expect(exported.weather).toBe('rain');
		expect(exported.room_count).toBe(3);
	});

	it('follows the schema property order', () => {
		const properties = DEFAULT_MAP_SCHEMA.properties as Record<string, unknown>;
		const schema = {
			...DEFAULT_MAP_SCHEMA,
			properties: {
				spawn: properties.spawn,
				id: properties.id,
				name: properties.name,
				format_version: properties.format_version,
				tile_size: properties.tile_size,
				width: properties.width,
				height: properties.height,
				tiles: properties.tiles,
				layers: properties.layers,
				objects: properties.objects
			}
		};

		const exported = buildExportJson(map(), schema);
		expect(Object.keys(exported)[0]).toBe('spawn');
		expect(Object.keys(exported)).toHaveLength(10);
	});

	it('serializes with the map values and schema extras', () => {
		const text = mapToJsonWithSchema(map(), schemaWith({ weather: { default: 'clear' } }));
		const parsed = JSON.parse(text);

		expect(parsed.id).toBe('test-map');
		expect(parsed.width).toBe(3);
		expect(parsed.weather).toBe('clear');
	});
});
