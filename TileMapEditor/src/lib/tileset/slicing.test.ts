import { describe, expect, it } from 'vitest';
import {
	computeTilesetGrid,
	computeTilesetSlices,
	findSliceAtPoint,
	sliceIndexAtPoint,
	type TilesetConfig
} from './slicing';

function config(overrides: Partial<TilesetConfig> = {}): TilesetConfig {
	return {
		imageWidth: 64,
		imageHeight: 32,
		tileWidth: 16,
		tileHeight: 16,
		spacing: 0,
		marginX: 0,
		marginY: 0,
		...overrides
	};
}

describe('computeTilesetGrid', () => {
	it('counts tiles across a plain sheet', () => {
		expect(computeTilesetGrid(config())).toEqual({ columns: 4, rows: 2 });
	});

	it('accounts for margin and spacing', () => {
		const grid = computeTilesetGrid(
			config({ imageWidth: 70, imageHeight: 38, spacing: 2, marginX: 2, marginY: 2 })
		);
		expect(grid).toEqual({ columns: 3, rows: 2 });
	});

	it('accounts for different x and y margins', () => {
		const grid = computeTilesetGrid(
			config({ imageWidth: 70, imageHeight: 38, spacing: 2, marginX: 2, marginY: 4 })
		);
		expect(grid).toEqual({ columns: 3, rows: 1 });
	});

	it('returns no tiles when the image is too small', () => {
		expect(computeTilesetGrid(config({ imageWidth: 10, imageHeight: 10 }))).toEqual({
			columns: 0,
			rows: 0
		});
	});

	it('rejects invalid dimensions', () => {
		expect(computeTilesetGrid(config({ tileWidth: 0 }))).toEqual({ columns: 0, rows: 0 });
		expect(computeTilesetGrid(config({ spacing: -1 }))).toEqual({ columns: 0, rows: 0 });
		expect(computeTilesetGrid(config({ marginX: -1 }))).toEqual({ columns: 0, rows: 0 });
		expect(computeTilesetGrid(config({ marginY: -1 }))).toEqual({ columns: 0, rows: 0 });
	});
});

describe('computeTilesetSlices', () => {
	it('builds row-major source rectangles', () => {
		const slices = computeTilesetSlices(config());
		expect(slices).toHaveLength(8);
		expect(slices[0]).toMatchObject({ index: 0, row: 0, column: 0, sx: 0, sy: 0 });
		expect(slices[3]).toMatchObject({ index: 3, row: 0, column: 3, sx: 48, sy: 0 });
		expect(slices[4]).toMatchObject({ index: 4, row: 1, column: 0, sx: 0, sy: 16 });
		expect(slices[7]).toMatchObject({ index: 7, row: 1, column: 3, sx: 48, sy: 16 });
	});

	it('offsets slices by margin and spacing', () => {
		const slices = computeTilesetSlices(
			config({ imageWidth: 70, imageHeight: 38, spacing: 2, marginX: 2, marginY: 2 })
		);
		expect(slices[0]).toMatchObject({ sx: 2, sy: 2 });
		expect(slices[1]).toMatchObject({ sx: 20, sy: 2 });
		expect(slices[3]).toMatchObject({ sx: 2, sy: 20 });
	});

	it('offsets slices by different x and y margins', () => {
		const slices = computeTilesetSlices(
			config({ imageWidth: 70, imageHeight: 46, spacing: 2, marginX: 4, marginY: 6 })
		);
		expect(slices[0]).toMatchObject({ sx: 4, sy: 6 });
		expect(slices[1]).toMatchObject({ sx: 22, sy: 6 });
		expect(slices[3]).toMatchObject({ sx: 4, sy: 24 });
	});
});

describe('slice hit testing', () => {
	const plain = config();

	it('maps image points to slice indices', () => {
		expect(sliceIndexAtPoint(plain, 0, 0)).toBe(0);
		expect(sliceIndexAtPoint(plain, 15, 15)).toBe(0);
		expect(sliceIndexAtPoint(plain, 16, 0)).toBe(1);
		expect(sliceIndexAtPoint(plain, 48, 16)).toBe(7);
	});

	it('ignores spacing and margin gaps', () => {
		const spaced = config({ imageWidth: 70, imageHeight: 38, spacing: 2, marginX: 2, marginY: 2 });
		expect(sliceIndexAtPoint(spaced, 18, 2)).toBeNull();
		expect(sliceIndexAtPoint(spaced, 2, 2)).toBe(0);
		expect(sliceIndexAtPoint(spaced, 1, 1)).toBeNull();
	});

	it('ignores asymmetric margin gaps', () => {
		const spaced = config({ imageWidth: 70, imageHeight: 38, spacing: 2, marginX: 4, marginY: 6 });
		expect(sliceIndexAtPoint(spaced, 4, 6)).toBe(0);
		expect(sliceIndexAtPoint(spaced, 3, 6)).toBeNull();
		expect(sliceIndexAtPoint(spaced, 4, 5)).toBeNull();
	});

	it('finds slices from applied geometry', () => {
		const slices = computeTilesetSlices(plain);
		expect(findSliceAtPoint(slices, 20, 20)?.index).toBe(5);
		expect(findSliceAtPoint(slices, 200, 200)).toBeNull();
	});
});
