import { describe, expect, it } from 'vitest';
import { createEmptyMap, createLayer, createTileDef } from './map-model/factory';
import { mapToJson, normalizeFilename, serializeMap } from './map-serializer';

function map() {
	const result = createEmptyMap({ width: 3, height: 2, id: 'deck-01', name: 'Deck One' });
	result.tiles = { floor: createTileDef('#555B62'), door: createTileDef('#9A7047', true) };
	result.layers = [createLayer('floor', 'none', 3, 2, 0)];
	result.layers[0].grid[0][1] = 'floor';
	result.objects = [
		{
			id: 'door-1',
			type: 'door',
			layer: 'floor',
			x: 1,
			y: 0,
			width: 1,
			height: 1,
			properties: { locked: false }
		}
	];
	result.spawn = { x: 2, y: 1 };
	return result;
}

describe('serializeMap', () => {
	it('emits exactly the game schema top-level fields', () => {
		const serialized = serializeMap(map());
		expect(Object.keys(serialized)).toEqual([
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
		]);
	});

	it('preserves layer, object, and spawn shapes', () => {
		const serialized = serializeMap(map());
		expect(serialized.layers[0]).toEqual({
			id: 'floor',
			type: 'tile',
			z_index: 0,
			collision: 'none',
			grid: [
				['', 'floor', ''],
				['', '', '']
			]
		});
		expect(serialized.tiles.door).toEqual({ solid: true, visual: { color: '#9A7047' } });
		expect(serialized.objects[0]).toEqual({
			id: 'door-1',
			type: 'door',
			layer: 'floor',
			x: 1,
			y: 0,
			width: 1,
			height: 1,
			properties: { locked: false }
		});
		expect(serialized.spawn).toEqual({ x: 2, y: 1 });
	});

	it('serializes multi-tile objects with explicit dimensions', () => {
		const source = map();
		source.width = 5;
		source.height = 5;
		source.objects = [
			{
				id: 'door-1',
				type: 'door',
				layer: 'floor',
				x: 0,
				y: 0,
				width: 2,
				height: 1,
				properties: {}
			},
			{
				id: 'seat-1',
				type: 'seat',
				layer: 'floor',
				x: 3,
				y: 0,
				width: 1,
				height: 2,
				properties: {}
			},
			{
				id: 'console-1',
				type: 'console',
				layer: 'floor',
				x: 0,
				y: 2,
				width: 3,
				height: 3,
				properties: {}
			}
		];

		expect(serializeMap(source).objects).toEqual([
			{
				id: 'door-1',
				type: 'door',
				layer: 'floor',
				x: 0,
				y: 0,
				width: 2,
				height: 1,
				properties: {}
			},
			{
				id: 'seat-1',
				type: 'seat',
				layer: 'floor',
				x: 3,
				y: 0,
				width: 1,
				height: 2,
				properties: {}
			},
			{
				id: 'console-1',
				type: 'console',
				layer: 'floor',
				x: 0,
				y: 2,
				width: 3,
				height: 3,
				properties: {}
			}
		]);
	});

	it('does not mutate the source map and deep-copies nested arrays', () => {
		const source = map();
		const serialized = serializeMap(source);
		serialized.layers[0].grid[0][1] = 'door';
		expect(source.layers[0].grid[0][1]).toBe('floor');
		serialized.tiles.floor.visual.color = '#ffffff';
		expect(source.tiles.floor.visual.color).toBe('#555B62');
	});

	it('produces parseable JSON', () => {
		const parsed = JSON.parse(mapToJson(map()));
		expect(parsed.id).toBe('deck-01');
		expect(parsed.layers[0].grid[0][1]).toBe('floor');
		expect(parsed.rooms).toBeUndefined();
	});
});

describe('normalizeFilename', () => {
	it('adds a json extension', () => {
		expect(normalizeFilename('deck', 'map')).toBe('deck.json');
		expect(normalizeFilename('deck.JSON', 'map')).toBe('deck.JSON');
	});

	it('falls back for blank names and strips illegal characters', () => {
		expect(normalizeFilename('  ', 'map.json')).toBe('map.json');
		expect(normalizeFilename('deck:1/2', 'map')).toBe('deck-1-2.json');
	});
});
