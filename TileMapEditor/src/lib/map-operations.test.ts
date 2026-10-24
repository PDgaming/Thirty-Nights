import { describe, expect, it } from 'vitest';
import { createEmptyMap, createLayer, createTileDef } from './map-model/factory';
import {
	addLayer,
	addTile,
	countTileReferences,
	describeResizeLoss,
	lineBetween,
	moveLayer,
	removeLayer,
	removeTile,
	renameLayer,
	renameTile,
	resizeMap,
	setCell,
	uniqueLayerId,
	uniqueTileId,
	updateTile
} from './map-operations';

function baseMap() {
	const map = createEmptyMap({ width: 4, height: 3, id: 'test', name: 'Test' });
	map.tiles = { floor: createTileDef('#111111'), wall: createTileDef('#222222', true) };
	map.layers = [createLayer('floor', 'none', 4, 3, 0), createLayer('walls', 'tile', 4, 3, 1)];
	return map;
}

describe('lineBetween', () => {
	it('returns a single cell', () => {
		expect(lineBetween({ x: 2, y: 2 }, { x: 2, y: 2 })).toEqual([{ x: 2, y: 2 }]);
	});

	it('fills a horizontal line without gaps', () => {
		expect(lineBetween({ x: 0, y: 0 }, { x: 3, y: 0 })).toEqual([
			{ x: 0, y: 0 },
			{ x: 1, y: 0 },
			{ x: 2, y: 0 },
			{ x: 3, y: 0 }
		]);
	});

	it('fills a diagonal line including both ends', () => {
		const cells = lineBetween({ x: 0, y: 0 }, { x: 3, y: 2 });
		expect(cells[0]).toEqual({ x: 0, y: 0 });
		expect(cells.at(-1)).toEqual({ x: 3, y: 2 });
		expect(cells).toHaveLength(4);
	});
});

describe('cell editing', () => {
	it('sets and clears cells through the layer', () => {
		const map = baseMap();
		expect(setCell(map, 'floor', 1, 2, 'floor')).toBe(true);
		expect(map.layers[0].grid[2][1]).toBe('floor');
		expect(setCell(map, 'floor', 1, 2, '')).toBe(true);
		expect(map.layers[0].grid[2][1]).toBe('');
	});

	it('rejects out-of-bounds edits', () => {
		const map = baseMap();
		expect(setCell(map, 'floor', 4, 0, 'floor')).toBe(false);
		expect(setCell(map, 'missing', 0, 0, 'floor')).toBe(false);
	});
});

describe('layer operations', () => {
	it('adds layers above existing ones', () => {
		const map = baseMap();
		const layer = addLayer(map, 'objects', 'object');
		expect(layer?.z_index).toBe(2);
		expect(map.layers).toHaveLength(3);
		expect(addLayer(map, 'objects', 'object')).toBeNull();
	});

	it('renames layers and updates object references', () => {
		const map = baseMap();
		map.objects = [{ id: 'o1', type: 'door', layer: 'floor', x: 0, y: 0, properties: {} }];
		expect(renameLayer(map, 'floor', 'ground')).toBe(true);
		expect(map.objects[0].layer).toBe('ground');
		expect(renameLayer(map, 'ground', 'walls')).toBe(false);
	});

	it('keeps at least one layer', () => {
		const map = baseMap();
		expect(removeLayer(map, 'floor')).toBe(true);
		expect(map.layers).toHaveLength(1);
		expect(removeLayer(map, 'walls')).toBe(false);
	});

	it('reorders layers and normalizes z-indexes', () => {
		const map = baseMap();
		expect(moveLayer(map, 'floor', 1)).toBe(true);
		expect(map.layers.map((layer) => layer.id)).toEqual(['walls', 'floor']);
		expect(map.layers.map((layer) => layer.z_index)).toEqual([0, 1]);
	});

	it('generates unique IDs', () => {
		const map = baseMap();
		expect(uniqueLayerId(map, 'floor')).toBe('floor_2');
		expect(uniqueTileId(map, 'floor')).toBe('floor_2');
		expect(uniqueTileId(map, 'ceiling')).toBe('ceiling');
	});
});

describe('tile operations', () => {
	it('adds and updates tile definitions', () => {
		const map = baseMap();
		expect(addTile(map, 'door', '#333333', true)).toBe(true);
		expect(addTile(map, 'door', '#333333')).toBe(false);
		expect(updateTile(map, 'door', { visual: { color: '#444444' }, solid: false })).toBe(true);
		expect(map.tiles.door).toMatchObject({ solid: false, visual: { color: '#444444' } });
	});

	it('renames tile IDs in the definition map and every grid', () => {
		const map = baseMap();
		map.layers[0].grid[0][0] = 'floor';
		map.layers[1].grid[1][1] = 'floor';
		expect(renameTile(map, 'floor', 'ground')).toBe(true);
		expect(map.tiles.floor).toBeUndefined();
		expect(map.tiles.ground).toBeDefined();
		expect(map.layers[0].grid[0][0]).toBe('ground');
		expect(map.layers[1].grid[1][1]).toBe('ground');
		expect(renameTile(map, 'ground', 'wall')).toBe(false);
	});

	it('refuses to delete referenced tiles', () => {
		const map = baseMap();
		map.layers[0].grid[0][0] = 'floor';
		expect(countTileReferences(map, 'floor')).toBe(1);
		expect(removeTile(map, 'floor')).toBe(false);
		expect(removeTile(map, 'wall')).toBe(true);
	});

	it('refuses to delete a tile used as an object type', () => {
		const map = baseMap();
		map.objects = [{ id: 'door-1', type: 'wall', layer: 'floor', x: 0, y: 0, properties: {} }];
		expect(removeTile(map, 'wall')).toBe(false);
	});
});

describe('resize', () => {
	it('preserves the overlapping region and reports loss', () => {
		const map = baseMap();
		map.layers[0].grid[0][0] = 'floor';
		map.layers[0].grid[2][3] = 'floor';
		map.spawn = { x: 3, y: 2 };

		const plan = describeResizeLoss(map, 2, 2);
		expect(plan.removedCells).toBe(1);
		expect(plan.spawnMoved).toBe(true);

		const result = resizeMap(map, 2, 2);
		expect(result.removedCells).toBe(1);
		expect(map.width).toBe(2);
		expect(map.height).toBe(2);
		expect(map.layers[0].grid).toHaveLength(2);
		expect(map.layers[0].grid[0]).toHaveLength(2);
		expect(map.layers[0].grid[0][0]).toBe('floor');
		expect(map.spawn).toEqual({ x: 1, y: 1 });
	});

	it('drops rooms and objects outside the new bounds', () => {
		const map = baseMap();
		map.rooms = [
			{ id: 'inside', name: 'Inside', bounds: { x: 1, y: 1, width: 1, height: 1 } },
			{ id: 'outside', name: 'Outside', bounds: { x: 3, y: 2, width: 1, height: 1 } }
		];
		map.objects = [
			{ id: 'in', type: 'floor', layer: 'floor', x: 0, y: 0, properties: {} },
			{ id: 'out', type: 'floor', layer: 'floor', x: 3, y: 2, properties: {} }
		];
		resizeMap(map, 2, 2);
		expect(map.rooms.map((room) => room.id)).toEqual(['inside']);
		expect(map.objects.map((object) => object.id)).toEqual(['in']);
	});
});
