import type { ObjectData } from './types';

/**
 * An object's origin and footprint in tile cells. The rectangle covers the
 * half-open range [x, x + width) x [y, y + height), so touching edges do not
 * overlap.
 */
export interface ObjectRect {
	x: number;
	y: number;
	width: number;
	height: number;
}

type FootprintSource = Pick<ObjectData, 'x' | 'y' | 'width' | 'height'>;

/** Returns the object's footprint as a rectangle in tile cells. */
export function objectRect(object: FootprintSource): ObjectRect {
	return {
		x: object.x,
		y: object.y,
		width: object.width,
		height: object.height
	};
}

/** Reports whether a cell is covered by the object's full footprint. */
export function cellCoveredByObject(object: FootprintSource, x: number, y: number): boolean {
	const rect = objectRect(object);
	return x >= rect.x && x < rect.x + rect.width && y >= rect.y && y < rect.y + rect.height;
}

/**
 * Reports whether the full footprint fits inside a map of the given tile
 * dimensions. The origin must be non-negative and the whole rectangle must end
 * at or before the map edge.
 */
export function objectFitsInMap(
	object: FootprintSource,
	mapWidth: number,
	mapHeight: number
): boolean {
	const rect = objectRect(object);
	return (
		rect.x >= 0 &&
		rect.y >= 0 &&
		rect.width >= 1 &&
		rect.height >= 1 &&
		rect.x + rect.width <= mapWidth &&
		rect.y + rect.height <= mapHeight
	);
}

/**
 * Reports whether two footprints share at least one cell. Touching edges are
 * not an overlap.
 */
export function objectFootprintsOverlap(a: FootprintSource, b: FootprintSource): boolean {
	const first = objectRect(a);
	const second = objectRect(b);
	return (
		first.x < second.x + second.width &&
		first.x + first.width > second.x &&
		first.y < second.y + second.height &&
		first.y + first.height > second.y
	);
}

/**
 * Coerces an unknown width or height value into an integer of at least 1 for
 * import and form use. This intentionally clamps invalid values; validation
 * must inspect the raw value so invalid input can be rejected instead of
 * silently repaired.
 */
export function normalizeObjectDimension(value: unknown): number {
	if (typeof value !== 'number' || !Number.isFinite(value) || !Number.isInteger(value)) {
		return 1;
	}
	return value >= 1 ? value : 1;
}
