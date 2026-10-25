import type { GameMap } from './map-model/types';
import { FORMAT_VERSION } from './map-model/types';

/**
 * Convert the in-memory map into a fresh object containing only the fields the
 * Go loader understands. Editor state is never included.
 */
export function serializeMap(map: GameMap): GameMap {
	return {
		format_version: FORMAT_VERSION,
		id: map.id,
		name: map.name,
		tile_size: map.tile_size,
		width: map.width,
		height: map.height,
		tiles: Object.fromEntries(
			Object.entries(map.tiles).map(([id, tile]) => [
				id,
				{ solid: tile.solid, visual: { color: tile.visual.color } }
			])
		),
		layers: map.layers.map((layer) => ({
			id: layer.id,
			type: layer.type,
			z_index: layer.z_index,
			collision: layer.collision,
			grid: layer.grid.map((row) => [...row])
		})),
		objects: map.objects.map((object) => ({
			id: object.id,
			type: object.type,
			layer: object.layer,
			x: object.x,
			y: object.y,
			properties: { ...object.properties }
		})),
		spawn: { x: map.spawn.x, y: map.spawn.y }
	};
}

export function mapToJson(map: GameMap): string {
	return JSON.stringify(serializeMap(map), null, 2);
}

export function normalizeFilename(filename: string, fallback: string): string {
	const trimmed = filename.trim();
	const base = trimmed === '' ? fallback : trimmed;
	const safe = base.replace(/[\\/:*?"<>|]/g, '-');
	return safe.toLowerCase().endsWith('.json') ? safe : `${safe}.json`;
}
