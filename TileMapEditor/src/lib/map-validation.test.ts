import { describe, expect, it } from 'vitest';
import { createEmptyMap, createLayer, createTileDef } from './map-model/factory';
import type { GameMap } from './map-model/types';
import { hasBlockingErrors, validateMap } from './map-validation';

function validMap(): GameMap {
	const map = createEmptyMap({ width: 4, height: 3, id: 'test', name: 'Test' });
	map.tiles = {
		floor: createTileDef('#555B62', false),
		wall: createTileDef('#89929C', true)
	};
	map.layers = [createLayer('floor', 'none', 4, 3, 0)];
	map.layers[0].grid[1][2] = 'floor';
	map.spawn = { x: 1, y: 2 };
	return map;
}

function messages(map: GameMap): string[] {
	return validateMap(map).map((issue) => `${issue.path}: ${issue.message}`);
}

describe('validateMap', () => {
	it('accepts a structurally valid map', () => {
		expect(validateMap(validMap())).toEqual([]);
	});

	it('rejects wrong format version and missing metadata', () => {
		const map = validMap();
		map.format_version = 2;
		map.id = ' ';
		map.name = '';
		const text = messages(map).join('|');
		expect(text).toContain('format_version');
		expect(text).toContain('id');
		expect(text).toContain('name');
	});

	it('requires at least one tile and one layer', () => {
		const map = validMap();
		map.tiles = {};
		map.layers = [];
		const text = messages(map).join('|');
		expect(text).toContain('tiles');
		expect(text).toContain('layers');
	});

	it('rejects invalid tile colours', () => {
		const map = validMap();
		map.tiles.floor.visual.color = 'red';
		expect(messages(map).join('|')).toContain('color must be');
	});

	it('rejects grid references to undefined tiles', () => {
		const map = validMap();
		map.layers[0].grid[0][0] = 'ghost';
		expect(messages(map).join('|')).toContain('undefined tile "ghost"');
	});

	it('rejects malformed grid dimensions', () => {
		const map = validMap();
		map.layers[0].grid.pop();
		map.layers[0].grid[0].pop();
		const text = messages(map).join('|');
		expect(text).toContain('rows; map height');
		expect(text).toContain('cells; map width');
	});

	it('rejects invalid collisions and negative z-index', () => {
		const map = validMap();
		map.layers[0].collision = 'bogus' as never;
		map.layers[0].z_index = -2;
		const text = messages(map).join('|');
		expect(text).toContain('collision must be');
		expect(text).toContain('z-index');
	});

	it('requires the spawn to be inside the map', () => {
		const map = validMap();
		map.spawn = { x: 9, y: 0 };
		expect(messages(map).join('|')).toContain('spawn (9, 0)');
	});

	it('validates object IDs, layers, and bounds', () => {
		const map = validMap();
		map.objects = [
			{
				id: 'chest-1',
				type: 'chest',
				layer: 'nope',
				x: 9,
				y: 0,
				properties: {}
			}
		];
		const text = messages(map).join('|');
		expect(text).toContain('undefined layer');
		expect(text).toContain('must be inside the map');
	});

	it('does not flag object types it does not recognize', () => {
		const map = validMap();
		map.objects = [{ id: 'chest-1', type: 'chest', layer: 'floor', x: 1, y: 1, properties: {} }];
		expect(validateMap(map)).toEqual([]);
	});

	it('requires a door object to have a door tile definition', () => {
		const map = validMap();
		map.objects = [{ id: 'door-1', type: 'door', layer: 'floor', x: 1, y: 1, properties: {} }];
		expect(messages(map).join('|')).toContain('must also exist as a tile definition');
	});

	it('accepts a door object when the door tile exists', () => {
		const map = validMap();
		map.tiles.door = createTileDef('#9A7047', true);
		map.objects = [{ id: 'door-1', type: 'door', layer: 'floor', x: 2, y: 1, properties: {} }];
		expect(hasBlockingErrors(validateMap(map))).toBe(false);
	});

	it('reports duplicate layer IDs', () => {
		const map = validMap();
		map.layers.push(createLayer('floor', 'none', 4, 3, 1));
		expect(messages(map).join('|')).toContain('duplicate layer ID');
	});
});
