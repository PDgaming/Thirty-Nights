import { describe, expect, it } from 'vitest';
import {
	cellCoveredByObject,
	normalizeObjectDimension,
	objectFitsInMap,
	objectFootprintsOverlap,
	objectRect
} from './object-geometry';

describe('objectRect', () => {
	it.each([
		{ width: 1, height: 1 },
		{ width: 2, height: 1 },
		{ width: 1, height: 2 },
		{ width: 3, height: 3 }
	])('builds a $width x $height rectangle at the origin', ({ width, height }) => {
		expect(objectRect({ x: 2, y: 5, width, height })).toEqual({
			x: 2,
			y: 5,
			width,
			height
		});
	});
});

describe('cellCoveredByObject', () => {
	const object = { x: 2, y: 4, width: 3, height: 3 };

	it('covers every corner of the footprint', () => {
		expect(cellCoveredByObject(object, 2, 4)).toBe(true);
		expect(cellCoveredByObject(object, 4, 4)).toBe(true);
		expect(cellCoveredByObject(object, 2, 6)).toBe(true);
		expect(cellCoveredByObject(object, 4, 6)).toBe(true);
	});

	it('covers the origin cell of a 1x1 object', () => {
		const single = { x: 2, y: 4, width: 1, height: 1 };
		expect(cellCoveredByObject(single, 2, 4)).toBe(true);
		expect(cellCoveredByObject(single, 3, 4)).toBe(false);
	});

	it('does not cover cells just outside the footprint', () => {
		expect(cellCoveredByObject(object, 1, 4)).toBe(false);
		expect(cellCoveredByObject(object, 5, 4)).toBe(false);
		expect(cellCoveredByObject(object, 2, 3)).toBe(false);
		expect(cellCoveredByObject(object, 2, 7)).toBe(false);
	});
});

describe('objectFitsInMap', () => {
	const mapWidth = 5;
	const mapHeight = 4;

	it('accepts footprints inside the map', () => {
		expect(objectFitsInMap({ x: 0, y: 0, width: 2, height: 2 }, mapWidth, mapHeight)).toBe(true);
		expect(objectFitsInMap({ x: 3, y: 2, width: 2, height: 2 }, mapWidth, mapHeight)).toBe(true);
	});

	it('accepts an exact fit at the right and bottom edges', () => {
		expect(objectFitsInMap({ x: 3, y: 0, width: 2, height: 1 }, mapWidth, mapHeight)).toBe(true);
		expect(objectFitsInMap({ x: 0, y: 2, width: 1, height: 2 }, mapWidth, mapHeight)).toBe(true);
	});

	it('rejects overflow on the right and bottom', () => {
		expect(objectFitsInMap({ x: 4, y: 0, width: 2, height: 1 }, mapWidth, mapHeight)).toBe(false);
		expect(objectFitsInMap({ x: 0, y: 3, width: 1, height: 2 }, mapWidth, mapHeight)).toBe(false);
	});

	it('rejects negative origins and oversized footprints', () => {
		expect(objectFitsInMap({ x: -1, y: 0, width: 1, height: 1 }, mapWidth, mapHeight)).toBe(false);
		expect(objectFitsInMap({ x: 0, y: -1, width: 1, height: 1 }, mapWidth, mapHeight)).toBe(false);
		expect(objectFitsInMap({ x: 0, y: 0, width: 6, height: 1 }, mapWidth, mapHeight)).toBe(false);
		expect(objectFitsInMap({ x: 0, y: 0, width: 1, height: 5 }, mapWidth, mapHeight)).toBe(false);
	});
});

describe('objectFootprintsOverlap', () => {
	it('treats touching edges as no overlap', () => {
		const left = { x: 0, y: 0, width: 2, height: 1 };
		const right = { x: 2, y: 0, width: 1, height: 1 };
		const below = { x: 0, y: 1, width: 2, height: 1 };
		expect(objectFootprintsOverlap(left, right)).toBe(false);
		expect(objectFootprintsOverlap(left, below)).toBe(false);
	});

	it('detects partial overlap', () => {
		expect(
			objectFootprintsOverlap(
				{ x: 0, y: 0, width: 2, height: 2 },
				{ x: 1, y: 1, width: 2, height: 2 }
			)
		).toBe(true);
	});

	it('detects containment in both directions', () => {
		expect(
			objectFootprintsOverlap(
				{ x: 0, y: 0, width: 3, height: 2 },
				{ x: 1, y: 1, width: 1, height: 1 }
			)
		).toBe(true);
		expect(
			objectFootprintsOverlap(
				{ x: 1, y: 1, width: 1, height: 1 },
				{ x: 0, y: 0, width: 3, height: 2 }
			)
		).toBe(true);
	});

	it('reports disjoint rectangles', () => {
		expect(
			objectFootprintsOverlap(
				{ x: 0, y: 0, width: 1, height: 1 },
				{ x: 3, y: 3, width: 1, height: 1 }
			)
		).toBe(false);
	});
});

describe('normalizeObjectDimension', () => {
	it('defaults missing and non-numeric values to one', () => {
		expect(normalizeObjectDimension(undefined)).toBe(1);
		expect(normalizeObjectDimension(null)).toBe(1);
		expect(normalizeObjectDimension('2')).toBe(1);
	});

	it('clamps zero and negative values to one', () => {
		expect(normalizeObjectDimension(0)).toBe(1);
		expect(normalizeObjectDimension(-3)).toBe(1);
	});

	it('clamps fractional values to one', () => {
		expect(normalizeObjectDimension(1.5)).toBe(1);
		expect(normalizeObjectDimension(-0.5)).toBe(1);
	});

	it('keeps valid integer dimensions', () => {
		expect(normalizeObjectDimension(1)).toBe(1);
		expect(normalizeObjectDimension(4)).toBe(4);
	});
});
