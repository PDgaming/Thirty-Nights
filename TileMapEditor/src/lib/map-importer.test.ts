import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { createEmptyMap, createTileDef } from './map-model/factory';
import { mapToJson } from './map-serializer';
import { MapImportError, parseMapData, parseMapJson } from './map-importer';

function baseRaw(overrides: Record<string, unknown> = {}) {
	return {
		format_version: 1,
		id: 'deck-01',
		name: 'Deck One',
		tile_size: 16,
		width: 3,
		height: 2,
		tiles: {
			floor: { solid: false, visual: { color: '#555B62' } }
		},
		layers: [
			{
				id: 'floor',
				type: 'tile',
				z_index: 0,
				collision: 'none',
				grid: [
					['floor', '', 'floor'],
					['', 'floor', '']
				]
			}
		],
		objects: [],
		spawn: { x: 1, y: 1 },
		...overrides
	};
}

describe('parseMapJson', () => {
	it('round-trips an exported map', () => {
		const map = createEmptyMap({ width: 3, height: 2, id: 'deck-01', name: 'Deck One' });
		map.tiles = { floor: createTileDef('#555B62'), door: createTileDef('#9A7047', true) };
		map.layers[0].grid[0][1] = 'floor';
		map.objects = [
			{ id: 'door-1', type: 'door', layer: 'floor', x: 1, y: 0, properties: { locked: false } }
		];
		map.spawn = { x: 2, y: 1 };

		expect(parseMapJson(mapToJson(map))).toEqual(map);
	});

	it('ignores retired rooms and unknown top-level fields', () => {
		const parsed = parseMapJson(
			JSON.stringify(
				baseRaw({
					rooms: [{ id: 'r1', name: 'Cabin', bounds: { x: 1, y: 1, width: 1, height: 1 } }],
					something_new: true
				})
			)
		);
		expect(parsed).not.toHaveProperty('rooms');
		expect(parsed).not.toHaveProperty('something_new');
	});

	it('defaults missing objects and spawn', () => {
		const raw = baseRaw();
		delete (raw as Record<string, unknown>).objects;
		delete (raw as Record<string, unknown>).spawn;
		const parsed = parseMapData(raw);
		expect(parsed.objects).toEqual([]);
		expect(parsed.spawn).toEqual({ x: 0, y: 0 });
	});

	it('rejects malformed JSON with a readable message', () => {
		expect(() => parseMapJson('{ nope')).toThrow(MapImportError);
		expect(() => parseMapJson('{ nope')).toThrow('Not valid JSON');
	});

	it('rejects unsupported format versions', () => {
		expect(() => parseMapData(baseRaw({ format_version: 2 }))).toThrow('format_version');
	});

	it('rejects grids that do not match the map dimensions', () => {
		const raw = baseRaw();
		(raw.layers as Array<{ grid: string[][] }>)[0].grid.pop();
		expect(() => parseMapData(raw)).toThrow('map height');
	});

	it('rejects invalid layer type and collision', () => {
		const raw = baseRaw();
		(raw.layers as Array<Record<string, unknown>>)[0].type = 'objects';
		expect(() => parseMapData(raw)).toThrow('must be "tile"');

		const other = baseRaw();
		(other.layers as Array<Record<string, unknown>>)[0].collision = 'bogus';
		expect(() => parseMapData(other)).toThrow('collision');
	});

	it('rejects malformed tile definitions', () => {
		const raw = baseRaw({
			tiles: { floor: { solid: false, visual: { color: 12 } } }
		});
		expect(() => parseMapData(raw)).toThrow('tiles.floor.visual.color');
	});

	it('parses the existing game map from the repository', () => {
		const path = fileURLToPath(new URL('../../../asset/maps/1.json', import.meta.url));
		const source = readFileSync(path, 'utf8');
		const raw = JSON.parse(source) as {
			id: string;
			width: number;
			height: number;
			layers: Array<{ id: string }>;
			objects: unknown[];
		};
		const parsed = parseMapJson(source);

		expect(parsed.id).toBe(raw.id);
		expect(parsed.width).toBe(raw.width);
		expect(parsed.height).toBe(raw.height);
		expect(parsed.layers.map((layer) => layer.id)).toEqual(raw.layers.map((layer) => layer.id));
		expect(parsed.objects).toHaveLength(raw.objects.length);
		expect(parsed).not.toHaveProperty('rooms');
	});

	it('defaults missing object properties', () => {
		const parsed = parseMapData(
			baseRaw({
				objects: [{ id: 'door-1', type: 'door', layer: 'floor', x: 1, y: 0 }]
			})
		);
		expect(parsed.objects[0].properties).toEqual({});
	});
});
