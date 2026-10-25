export interface TilesetConfig {
	imageWidth: number;
	imageHeight: number;
	tileWidth: number;
	tileHeight: number;
	spacing: number;
	marginX: number;
	marginY: number;
}

export interface TilesetSlice {
	index: number;
	row: number;
	column: number;
	sx: number;
	sy: number;
	sw: number;
	sh: number;
}

export const MAX_TILESET_SLICES = 4096;

/**
 * Number of tile columns and rows that fit inside a tileset image. The formula
 * mirrors common tileset layouts: margin around the sheet, spacing between
 * tiles, and no trailing spacing required after the last tile.
 */
export function computeTilesetGrid(config: TilesetConfig): { columns: number; rows: number } {
	const { imageWidth, imageHeight, tileWidth, tileHeight, spacing, marginX, marginY } = config;

	if (
		![imageWidth, imageHeight, tileWidth, tileHeight].every((value) => Number.isFinite(value)) ||
		imageWidth <= 0 ||
		imageHeight <= 0 ||
		tileWidth <= 0 ||
		tileHeight <= 0 ||
		spacing < 0 ||
		marginX < 0 ||
		marginY < 0
	) {
		return { columns: 0, rows: 0 };
	}

	const usableWidth = imageWidth - marginX * 2;
	const usableHeight = imageHeight - marginY * 2;
	if (usableWidth < tileWidth || usableHeight < tileHeight) {
		return { columns: 0, rows: 0 };
	}

	const columns = Math.floor((usableWidth + spacing) / (tileWidth + spacing));
	const rows = Math.floor((usableHeight + spacing) / (tileHeight + spacing));

	return { columns: Math.max(columns, 0), rows: Math.max(rows, 0) };
}

export function computeTilesetSlices(config: TilesetConfig): TilesetSlice[] {
	const { columns, rows } = computeTilesetGrid(config);
	const slices: TilesetSlice[] = [];

	for (let row = 0; row < rows; row += 1) {
		for (let column = 0; column < columns; column += 1) {
			slices.push({
				index: row * columns + column,
				row,
				column,
				sx: config.marginX + column * (config.tileWidth + config.spacing),
				sy: config.marginY + row * (config.tileHeight + config.spacing),
				sw: config.tileWidth,
				sh: config.tileHeight
			});
		}
	}

	return slices;
}

/** Which applied slice, if any, contains a point in source-image coordinates. */
export function findSliceAtPoint(
	slices: TilesetSlice[],
	x: number,
	y: number
): TilesetSlice | null {
	for (const slice of slices) {
		if (x >= slice.sx && x < slice.sx + slice.sw && y >= slice.sy && y < slice.sy + slice.sh) {
			return slice;
		}
	}
	return null;
}

/** Which slice, if any, sits under a point in source-image pixel coordinates. */
export function sliceIndexAtPoint(config: TilesetConfig, x: number, y: number): number | null {
	const { columns, rows } = computeTilesetGrid(config);
	if (columns === 0 || rows === 0) return null;

	const localX = x - config.marginX;
	const localY = y - config.marginY;
	if (localX < 0 || localY < 0) return null;

	const column = Math.floor(localX / (config.tileWidth + config.spacing));
	const row = Math.floor(localY / (config.tileHeight + config.spacing));
	if (column < 0 || row < 0 || column >= columns || row >= rows) return null;

	const withinX = localX - column * (config.tileWidth + config.spacing);
	const withinY = localY - row * (config.tileHeight + config.spacing);
	if (withinX >= config.tileWidth || withinY >= config.tileHeight) return null;

	return row * columns + column;
}
