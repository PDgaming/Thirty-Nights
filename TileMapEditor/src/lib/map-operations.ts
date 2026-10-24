import type {
	Cell,
	CellPosition,
	CollisionName,
	GameMap,
	LayerData,
	RoomData,
	TileDef
} from './map-model/types';
import { createTileDef } from './map-model/factory';

export function findLayer(map: GameMap, layerID: string): LayerData | undefined {
	return map.layers.find((layer) => layer.id === layerID);
}

export function inBounds(map: GameMap, x: number, y: number): boolean {
	return (
		Number.isInteger(x) &&
		Number.isInteger(y) &&
		x >= 0 &&
		y >= 0 &&
		x < map.width &&
		y < map.height
	);
}

/** Returns false when the layer or cell does not exist. */
export function setCell(
	map: GameMap,
	layerID: string,
	x: number,
	y: number,
	tileID: Cell
): boolean {
	const layer = findLayer(map, layerID);
	if (!layer || !inBounds(map, x, y)) return false;
	layer.grid[y][x] = tileID;
	return true;
}

/** Integer grid line between two cells, used for gap-free drag painting. */
export function lineBetween(from: CellPosition, to: CellPosition): CellPosition[] {
	const cells: CellPosition[] = [];
	let x = from.x;
	let y = from.y;
	const dx = Math.abs(to.x - from.x);
	const dy = Math.abs(to.y - from.y);
	const sx = from.x < to.x ? 1 : -1;
	const sy = from.y < to.y ? 1 : -1;
	let err = dx - dy;

	for (;;) {
		cells.push({ x, y });
		if (x === to.x && y === to.y) break;
		const e2 = 2 * err;
		if (e2 > -dy) {
			err -= dy;
			x += sx;
		}
		if (e2 < dx) {
			err += dx;
			y += sy;
		}
	}

	return cells;
}

export function countPopulatedCells(layer: LayerData): number {
	let count = 0;
	for (const row of layer.grid) {
		for (const cell of row) {
			if (cell !== '') count += 1;
		}
	}
	return count;
}

export function uniqueLayerId(map: GameMap, base = 'layer'): string {
	const ids = new Set(map.layers.map((layer) => layer.id));
	if (!ids.has(base)) return base;
	let suffix = 2;
	while (ids.has(`${base}_${suffix}`)) suffix += 1;
	return `${base}_${suffix}`;
}

export function uniqueTileId(map: GameMap, base = 'tile'): string {
	if (!Object.prototype.hasOwnProperty.call(map.tiles, base)) return base;
	let suffix = 2;
	while (Object.prototype.hasOwnProperty.call(map.tiles, `${base}_${suffix}`)) suffix += 1;
	return `${base}_${suffix}`;
}

export function addLayer(map: GameMap, id: string, collision: CollisionName): LayerData | null {
	if (map.layers.some((layer) => layer.id === id)) return null;
	const maxZ = map.layers.reduce((max, layer) => Math.max(max, layer.z_index), -1);
	const layer: LayerData = {
		id,
		type: 'tile',
		z_index: maxZ + 1,
		collision,
		grid: Array.from({ length: map.height }, () => Array.from({ length: map.width }, () => ''))
	};
	map.layers = [...map.layers, layer];
	return layer;
}

export function removeLayer(map: GameMap, layerID: string): boolean {
	if (map.layers.length <= 1) return false;
	const next = map.layers.filter((layer) => layer.id !== layerID);
	if (next.length === map.layers.length) return false;
	map.layers = next;
	normalizeLayerOrder(map);
	return true;
}

export function renameLayer(map: GameMap, oldID: string, newID: string): boolean {
	if (oldID === newID) return true;
	const trimmed = newID.trim();
	if (trimmed === '') return false;
	if (map.layers.some((layer) => layer.id === trimmed)) return false;
	const layer = findLayer(map, oldID);
	if (!layer) return false;

	layer.id = trimmed;
	map.layers = [...map.layers];
	for (const object of map.objects) {
		if (object.layer === oldID) object.layer = trimmed;
	}
	return true;
}

export function setLayerCollision(
	map: GameMap,
	layerID: string,
	collision: CollisionName
): boolean {
	const layer = findLayer(map, layerID);
	if (!layer) return false;
	layer.collision = collision;
	map.layers = [...map.layers];
	return true;
}

/** direction > 0 moves the layer toward the top (higher z-index). */
export function moveLayer(map: GameMap, layerID: string, direction: number): boolean {
	const ordered = [...map.layers].sort((a, b) => a.z_index - b.z_index);
	const index = ordered.findIndex((layer) => layer.id === layerID);
	if (index === -1) return false;
	const target = index + Math.sign(direction);
	if (target < 0 || target >= ordered.length) return false;

	const [layer] = ordered.splice(index, 1);
	ordered.splice(target, 0, layer);
	ordered.forEach((item, zIndex) => {
		item.z_index = zIndex;
	});
	map.layers = ordered;
	return true;
}

export function normalizeLayerOrder(map: GameMap): void {
	const ordered = [...map.layers].sort((a, b) => a.z_index - b.z_index);
	ordered.forEach((layer, index) => {
		layer.z_index = index;
	});
	map.layers = ordered;
}

export function addTile(map: GameMap, id: string, color: string, solid = false): boolean {
	const trimmed = id.trim();
	if (trimmed === '') return false;
	if (Object.prototype.hasOwnProperty.call(map.tiles, trimmed)) return false;
	map.tiles = { ...map.tiles, [trimmed]: createTileDef(color, solid) };
	return true;
}

export function updateTile(map: GameMap, id: string, patch: Partial<TileDef>): boolean {
	const tile = map.tiles[id];
	if (!tile) return false;
	map.tiles = {
		...map.tiles,
		[id]: {
			solid: patch.solid ?? tile.solid,
			visual: { color: patch.visual?.color ?? tile.visual.color }
		}
	};
	return true;
}

export function renameTile(map: GameMap, oldID: string, newID: string): boolean {
	const trimmed = newID.trim();
	if (trimmed === '' || oldID === trimmed) return false;
	if (Object.prototype.hasOwnProperty.call(map.tiles, trimmed)) return false;
	const tile = map.tiles[oldID];
	if (!tile) return false;

	const nextTiles = { ...map.tiles };
	delete nextTiles[oldID];
	nextTiles[trimmed] = tile;
	map.tiles = nextTiles;

	for (const layer of map.layers) {
		for (const row of layer.grid) {
			for (let x = 0; x < row.length; x += 1) {
				if (row[x] === oldID) row[x] = trimmed;
			}
		}
	}
	return true;
}

export function countTileReferences(map: GameMap, tileID: string): number {
	let count = 0;
	for (const layer of map.layers) {
		for (const row of layer.grid) {
			for (const cell of row) {
				if (cell === tileID) count += 1;
			}
		}
	}
	return count;
}

export function tileIsReferencedByObject(map: GameMap, tileID: string): boolean {
	return map.objects.some((object) => object.type === tileID);
}

export function removeTile(map: GameMap, tileID: string): boolean {
	if (countTileReferences(map, tileID) > 0 || tileIsReferencedByObject(map, tileID)) {
		return false;
	}
	if (!Object.prototype.hasOwnProperty.call(map.tiles, tileID)) return false;
	const nextTiles = { ...map.tiles };
	delete nextTiles[tileID];
	map.tiles = nextTiles;
	return true;
}

export interface ResizeResult {
	removedCells: number;
	removedRooms: number;
	removedObjects: number;
	spawnMoved: boolean;
}

export function describeResizeLoss(map: GameMap, width: number, height: number): ResizeResult {
	let removedCells = 0;
	for (const layer of map.layers) {
		for (let y = 0; y < layer.grid.length; y += 1) {
			for (let x = 0; x < layer.grid[y].length; x += 1) {
				if (x >= width || y >= height) {
					if (layer.grid[y][x] !== '') removedCells += 1;
				}
			}
		}
	}

	const removedRooms = map.rooms.filter((room) => !roomFits(room, width, height)).length;
	const removedObjects = map.objects.filter(
		(object) => object.x >= width || object.y >= height
	).length;
	const spawnMoved = map.spawn.x >= width || map.spawn.y >= height;

	return { removedCells, removedRooms, removedObjects, spawnMoved };
}

export function resizeMap(map: GameMap, width: number, height: number): ResizeResult {
	const result = describeResizeLoss(map, width, height);

	for (const layer of map.layers) {
		const next: Cell[][] = Array.from({ length: height }, () =>
			Array.from({ length: width }, () => '')
		);
		for (let y = 0; y < Math.min(height, layer.grid.length); y += 1) {
			for (let x = 0; x < Math.min(width, layer.grid[y].length); x += 1) {
				next[y][x] = layer.grid[y][x];
			}
		}
		layer.grid = next;
	}

	map.rooms = map.rooms.filter((room) => roomFits(room, width, height));
	map.objects = map.objects.filter((object) => object.x < width && object.y < height);
	if (result.spawnMoved) {
		map.spawn = {
			x: Math.min(Math.max(0, map.spawn.x), width - 1),
			y: Math.min(Math.max(0, map.spawn.y), height - 1)
		};
	}

	map.width = width;
	map.height = height;
	map.layers = [...map.layers];
	return result;
}

function roomFits(room: RoomData, width: number, height: number): boolean {
	return room.bounds.x + room.bounds.width <= width && room.bounds.y + room.bounds.height <= height;
}

export function resizeLossLabels(result: ResizeResult): string[] {
	const labels: string[] = [];
	if (result.removedCells > 0) {
		labels.push(`${result.removedCells} painted cell${result.removedCells === 1 ? '' : 's'}`);
	}
	if (result.removedRooms > 0) {
		labels.push(`${result.removedRooms} room${result.removedRooms === 1 ? '' : 's'}`);
	}
	if (result.removedObjects > 0) {
		labels.push(`${result.removedObjects} object${result.removedObjects === 1 ? '' : 's'}`);
	}
	if (result.spawnMoved) labels.push('the spawn point will be clamped');
	return labels;
}
