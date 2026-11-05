import type { CellPosition, GameMap, ObjectData } from '../map-model/types';
import { cellCoveredByObject, objectFitsInMap } from '../map-model/object-geometry';
import type { TilesetSlice } from '../tileset/slicing';
import type { Viewport } from './viewport';

export type EditorTool = 'select' | 'paint' | 'erase' | 'object' | 'spawn';

export interface TilesetRenderData {
	image: HTMLImageElement | null;
	slices: TilesetSlice[];
	bindings: Record<string, number>;
}

export interface MapRenderOptions {
	ctx: CanvasRenderingContext2D;
	width: number;
	height: number;
	map: GameMap;
	view: Viewport;
	hiddenLayers: readonly string[];
	showGrid: boolean;
	showTilesetArt: boolean;
	tileset: TilesetRenderData | null;
	hover: { x: number; y: number } | null;
	activeTileId: string | null;
	activeLayerId: string | null;
	activeObjectTypeId: string | null;
	objectColors: Record<string, string>;
	onTextureLoad?: () => void;
	selectedCell: CellPosition | null;
	draggingObjectId: string | null;
	dragTarget: CellPosition | null;
	placementBrushWidth: number;
	placementBrushHeight: number;
	tool: EditorTool;
}

const BACKGROUND = '#171717';
const MAP_BACKGROUND = '#2b2b2b';
const MAP_BORDER = '#0f0f0f';
const GRID = 'rgba(255, 255, 255, 0.05)';
const GRID_STRONG = 'rgba(255, 255, 255, 0.11)';
const ACCENT = '#6f9fd8';
const ERASE = '#e07b72';
const SPAWN = '#d9a441';
const OBJECT = '#c98bb9';
const OBJECT_INK = '#141414';

const textureCache = new Map<string, HTMLImageElement>();

function textureImage(name: string): HTMLImageElement {
	let img = textureCache.get(name);
	if (!img) {
		img = new Image();
		img.src = name;
		textureCache.set(name, img);
	}
	return img;
}

export function drawMap(options: MapRenderOptions): void {
	const { ctx, width, height, map, view } = options;
	ctx.clearRect(0, 0, width, height);
	ctx.fillStyle = BACKGROUND;
	ctx.fillRect(0, 0, width, height);

	const tilePx = map.tile_size * view.scale;
	if (!Number.isFinite(tilePx) || tilePx <= 0) return;

	const mapX = view.offsetX;
	const mapY = view.offsetY;
	const mapWidth = map.width * tilePx;
	const mapHeight = map.height * tilePx;

	// The map plate floats above the canvas bed, so artwork reads as separate
	// from the chrome even when a layer is mostly empty.
	ctx.save();
	ctx.shadowColor = 'rgba(0, 0, 0, 0.55)';
	ctx.shadowBlur = 20;
	ctx.shadowOffsetY = 6;
	ctx.fillStyle = MAP_BACKGROUND;
	ctx.fillRect(mapX, mapY, mapWidth, mapHeight);
	ctx.restore();

	const startX = Math.max(0, Math.floor(-view.offsetX / tilePx));
	const startY = Math.max(0, Math.floor(-view.offsetY / tilePx));
	const endX = Math.min(map.width - 1, Math.floor((width - view.offsetX) / tilePx));
	const endY = Math.min(map.height - 1, Math.floor((height - view.offsetY) / tilePx));

	if (endX >= startX && endY >= startY) {
		drawLayers(options, startX, startY, endX, endY, tilePx);
		if (options.showGrid) {
			const gridAlpha = Math.min(1, Math.max(0, (tilePx - 7) / 18));
			if (gridAlpha > 0.02) {
				drawGrid(options, startX, startY, endX, endY, tilePx, gridAlpha);
			}
		}
	}

	drawSelection(options, tilePx);

	ctx.strokeStyle = MAP_BORDER;
	ctx.lineWidth = 1;
	ctx.strokeRect(mapX + 0.5, mapY + 0.5, mapWidth - 1, mapHeight - 1);
	ctx.strokeStyle = 'rgba(255, 255, 255, 0.045)';
	ctx.strokeRect(mapX + 1.5, mapY + 1.5, mapWidth - 3, mapHeight - 3);

	drawObjects(options, tilePx);
	drawObjectDrag(options, tilePx);
	drawSpawn(options, tilePx);
	drawHover(options, tilePx);
}

function drawLayers(
	options: MapRenderOptions,
	startX: number,
	startY: number,
	endX: number,
	endY: number,
	tilePx: number
): void {
	const { ctx, map, view, hiddenLayers, showTilesetArt, tileset } = options;
	const hidden = new Set(hiddenLayers);
	const ordered = [...map.layers].sort((a, b) => a.z_index - b.z_index);
	const canDrawArt =
		showTilesetArt && tileset !== null && tileset.image !== null && tileset.image.complete;

	for (const layer of ordered) {
		if (hidden.has(layer.id)) continue;

		for (let y = startY; y <= endY; y += 1) {
			const row = layer.grid[y];
			if (!row) continue;
			for (let x = startX; x <= endX; x += 1) {
				const tileID = row[x];
				if (!tileID) continue;
				const tile = map.tiles[tileID];
				if (!tile) continue;

				const px = view.offsetX + x * tilePx;
				const py = view.offsetY + y * tilePx;
				const drawWidth = Math.ceil(tilePx);
				const drawHeight = Math.ceil(tilePx);

				// If the tile has a texture image, draw it
				if (tile.visual.texture) {
					const img = textureImage(tile.visual.texture);
					if (img.complete) {
						ctx.imageSmoothingEnabled = false;
						ctx.drawImage(img, px, py, drawWidth, drawHeight);
					} else if (options.onTextureLoad) {
						if (!img.dataset.ready) {
							img.dataset.ready = '1';
							img.onload = () => options.onTextureLoad?.();
						}
					}
				} else {
					ctx.fillStyle = tile.visual.color;
					ctx.fillRect(px, py, drawWidth, drawHeight);
				}

				if (canDrawArt && tileset && !tile.visual.texture) {
					const sliceIndex = tileset.bindings[tileID];
					const slice = sliceIndex === undefined ? undefined : tileset.slices[sliceIndex];
					if (slice) {
						ctx.drawImage(
							tileset.image as HTMLImageElement,
							slice.sx,
							slice.sy,
							slice.sw,
							slice.sh,
							px,
							py,
							drawWidth,
							drawHeight
						);
					}
				}
			}
		}
	}
}

function drawTileArt(
	ctx: CanvasRenderingContext2D,
	tileID: string,
	px: number,
	py: number,
	width: number,
	height: number,
	tileset: TilesetRenderData
): void {
	const image = tileset.image;
	if (!image) return;
	const index = tileset.bindings[tileID];
	const slice = index === undefined ? undefined : tileset.slices[index];
	if (!slice) return;
	ctx.drawImage(image, slice.sx, slice.sy, slice.sw, slice.sh, px, py, width, height);
}

function drawGrid(
	options: MapRenderOptions,
	startX: number,
	startY: number,
	endX: number,
	endY: number,
	tilePx: number,
	alpha: number
): void {
	const { ctx, view } = options;
	ctx.strokeStyle = tilePx >= 32 ? GRID_STRONG : GRID;
	ctx.globalAlpha = alpha;
	ctx.lineWidth = 1;
	ctx.beginPath();
	for (let x = startX; x <= endX + 1; x += 1) {
		const px = Math.round(view.offsetX + x * tilePx) + 0.5;
		ctx.moveTo(px, view.offsetY + startY * tilePx);
		ctx.lineTo(px, view.offsetY + (endY + 1) * tilePx);
	}
	for (let y = startY; y <= endY + 1; y += 1) {
		const py = Math.round(view.offsetY + y * tilePx) + 0.5;
		ctx.moveTo(view.offsetX + startX * tilePx, py);
		ctx.lineTo(view.offsetX + (endX + 1) * tilePx, py);
	}
	ctx.stroke();
	ctx.globalAlpha = 1;
}

function drawObjects(options: MapRenderOptions, tilePx: number): void {
	const { ctx, map, view, objectColors, hover, tool, draggingObjectId } = options;
	let hoveredId: string | null = null;
	if (hover) {
		for (let index = map.objects.length - 1; index >= 0; index -= 1) {
			const object = map.objects[index];
			if (cellCoveredByObject(object, hover.x, hover.y)) {
				hoveredId = object.id;
				break;
			}
		}
	}

	for (const object of map.objects) {
		const { px, py, width, height } = footprintPx(view, object, tilePx);
		const color = objectColors[object.type] ?? OBJECT;
		const isDragging = object.id === draggingObjectId;
		const multiTile = object.width > 1 || object.height > 1;
		drawObjectFootprint(ctx, px, py, width, height, color, multiTile, isDragging ? 0.08 : 0.16);
		drawObjectMarker(ctx, px, py, width, height, color, isDragging ? 0.25 : 0.92);
		if (!isDragging && tool === 'select' && object.id === hoveredId) {
			drawObjectFocus(ctx, px, py, width, height);
		}
	}
}

interface FootprintPx {
	px: number;
	py: number;
	width: number;
	height: number;
}

/** Pixel rectangle for an object's footprint at the current zoom. */
function footprintPx(
	view: Viewport,
	object: Pick<ObjectData, 'x' | 'y' | 'width' | 'height'> & { properties: Record<string, unknown> },
	tilePx: number
): FootprintPx {
	const interactionPadding = (object.properties['interactionPadding'] as number) ?? 0;
	const pad = interactionPadding * tilePx;
	return {
		px: view.offsetX + object.x * tilePx - pad,
		py: view.offsetY + object.y * tilePx - pad,
		width: Math.ceil(object.width * tilePx + pad * 2),
		height: Math.ceil(object.height * tilePx + pad * 2)
	};
}

/** Tinted rectangle behind a multi-tile object, so its full footprint reads at a glance. */
function drawObjectFootprint(
	ctx: CanvasRenderingContext2D,
	px: number,
	py: number,
	width: number,
	height: number,
	color: string,
	visible: boolean,
	alpha: number
): void {
	if (!visible) return;
	ctx.save();
	ctx.globalAlpha = alpha;
	ctx.fillStyle = color;
	ctx.fillRect(px, py, width, height);
	ctx.globalAlpha = Math.min(1, alpha + 0.35);
	ctx.strokeStyle = color;
	ctx.lineWidth = 1;
	ctx.strokeRect(px + 0.5, py + 0.5, width - 1, height - 1);
	ctx.restore();
}

/** Dashed snapped-footprint preview and ghost marker while an object is dragged. */
function drawObjectDrag(options: MapRenderOptions, tilePx: number): void {
	const { ctx, map, view, objectColors, draggingObjectId, dragTarget } = options;
	if (draggingObjectId === null || dragTarget === null) return;

	const object = map.objects.find((item) => item.id === draggingObjectId);
	if (!object) return;
	const target = {
		x: dragTarget.x,
		y: dragTarget.y,
		width: object.width,
		height: object.height
	};
	if (!objectFitsInMap(target, map.width, map.height)) return;

	const { px, py, width, height } = footprintPx(view, target, tilePx);

	ctx.save();
	ctx.fillStyle = 'rgba(111, 159, 216, 0.16)';
	ctx.fillRect(px, py, width, height);
	ctx.strokeStyle = ACCENT;
	ctx.lineWidth = 2;
	ctx.setLineDash([4, 3]);
	ctx.strokeRect(px + 1, py + 1, width - 2, height - 2);
	ctx.restore();

	drawObjectMarker(ctx, px, py, width, height, objectColors[object.type] ?? OBJECT, 0.92);
	drawObjectFocus(ctx, px, py, width, height);
}

function drawSelection(options: MapRenderOptions, tilePx: number): void {
	const { ctx, map, view, selectedCell, tool } = options;
	if (tool !== 'select' || !selectedCell) return;
	if (
		selectedCell.x < 0 ||
		selectedCell.y < 0 ||
		selectedCell.x >= map.width ||
		selectedCell.y >= map.height
	) {
		return;
	}

	let selectedObject: ObjectData | undefined;
	for (let index = map.objects.length - 1; index >= 0; index -= 1) {
		const object = map.objects[index];
		if (cellCoveredByObject(object, selectedCell.x, selectedCell.y)) {
			selectedObject = object;
			break;
		}
	}

	const { px, py, width, height } = selectedObject
		? footprintPx(view, selectedObject, tilePx)
		: {
				px: view.offsetX + selectedCell.x * tilePx,
				py: view.offsetY + selectedCell.y * tilePx,
				width: Math.ceil(tilePx),
				height: Math.ceil(tilePx)
			};

	ctx.save();
	ctx.fillStyle = 'rgba(111, 159, 216, 0.16)';
	ctx.fillRect(px, py, width, height);
	ctx.strokeStyle = ACCENT;
	ctx.lineWidth = 2;
	ctx.strokeRect(px + 1, py + 1, width - 2, height - 2);
	ctx.restore();
}

function drawObjectFocus(
	ctx: CanvasRenderingContext2D,
	px: number,
	py: number,
	width: number,
	height: number
): void {
	const cx = px + width / 2;
	const cy = py + height / 2;
	const radius = Math.max(3, Math.min(Math.min(width, height) * 0.3, 11));

	ctx.save();
	ctx.strokeStyle = ACCENT;
	ctx.lineWidth = 1.5;
	ctx.beginPath();
	ctx.arc(cx, cy, radius + 3, 0, Math.PI * 2);
	ctx.stroke();
	ctx.restore();
}

function drawObjectMarker(
	ctx: CanvasRenderingContext2D,
	px: number,
	py: number,
	width: number,
	height: number,
	color: string,
	alpha: number
): void {
	const cx = px + width / 2;
	const cy = py + height / 2;
	const radius = Math.max(3, Math.min(Math.min(width, height) * 0.3, 11));

	ctx.save();
	ctx.globalAlpha = alpha;
	ctx.beginPath();
	ctx.moveTo(cx, cy - radius);
	ctx.lineTo(cx + radius, cy);
	ctx.lineTo(cx, cy + radius);
	ctx.lineTo(cx - radius, cy);
	ctx.closePath();
	ctx.fillStyle = color;
	ctx.fill();
	ctx.strokeStyle = OBJECT_INK;
	ctx.lineWidth = 1.5;
	ctx.stroke();
	ctx.restore();
}

function drawSpawn(options: MapRenderOptions, tilePx: number): void {
	const { ctx, map, view } = options;
	const px = view.offsetX + (map.spawn.x + 0.5) * tilePx;
	const py = view.offsetY + (map.spawn.y + 0.5) * tilePx;
	const radius = Math.max(4, Math.min(tilePx * 0.32, 14));

	ctx.save();
	ctx.strokeStyle = SPAWN;
	ctx.fillStyle = 'rgba(217, 164, 65, 0.16)';
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.arc(px, py, radius, 0, Math.PI * 2);
	ctx.fill();
	ctx.stroke();
	ctx.beginPath();
	ctx.moveTo(px - radius * 1.6, py);
	ctx.lineTo(px + radius * 1.6, py);
	ctx.moveTo(px, py - radius * 1.6);
	ctx.lineTo(px, py + radius * 1.6);
	ctx.stroke();
	ctx.restore();
}

function drawHover(options: MapRenderOptions, tilePx: number): void {
	const {
		ctx,
		map,
		view,
		hover,
		tool,
		activeTileId,
		activeLayerId,
		activeObjectTypeId,
		objectColors,
		hiddenLayers,
		showTilesetArt,
		tileset,
		draggingObjectId,
		placementBrushWidth,
		placementBrushHeight
	} = options;
	if (!hover) return;
	if (draggingObjectId !== null) return;
	if (hover.x < 0 || hover.y < 0 || hover.x >= map.width || hover.y >= map.height) return;

	const px = view.offsetX + hover.x * tilePx;
	const py = view.offsetY + hover.y * tilePx;
	const drawWidth = Math.ceil(tilePx);
	const drawHeight = Math.ceil(tilePx);

	ctx.save();
	if (tool === 'spawn') {
		ctx.fillStyle = 'rgba(217, 164, 65, 0.2)';
		ctx.fillRect(px, py, drawWidth, drawHeight);
		ctx.strokeStyle = SPAWN;
		ctx.lineWidth = 2;
		ctx.strokeRect(px + 1, py + 1, drawWidth - 2, drawHeight - 2);
	} else if (tool === 'erase') {
		ctx.fillStyle = 'rgba(224, 123, 114, 0.24)';
		ctx.fillRect(px, py, drawWidth, drawHeight);
		ctx.strokeStyle = ERASE;
		ctx.lineWidth = 1.5;
		ctx.strokeRect(px + 0.5, py + 0.5, drawWidth - 1, drawHeight - 1);
	} else if (tool === 'select') {
		ctx.strokeStyle = 'rgba(255, 255, 255, 0.55)';
		ctx.lineWidth = 1;
		ctx.strokeRect(px + 0.5, py + 0.5, drawWidth - 1, drawHeight - 1);
	} else if (tool === 'object') {
		const color = activeObjectTypeId ? objectColors[activeObjectTypeId] : undefined;
		const previewWidth = Math.ceil(placementBrushWidth * tilePx);
		const previewHeight = Math.ceil(placementBrushHeight * tilePx);
		ctx.strokeStyle = ACCENT;
		ctx.lineWidth = 1.5;
		ctx.strokeRect(px + 0.5, py + 0.5, previewWidth - 1, previewHeight - 1);
		if (color) {
			drawObjectMarker(ctx, px, py, previewWidth, previewHeight, color, 0.55);
		}
	} else if (activeTileId) {
		const activeTile = map.tiles[activeTileId];
		const activeLayerVisible = activeLayerId !== null && !hiddenLayers.includes(activeLayerId);
		if (activeLayerVisible && activeTile) {
			ctx.globalAlpha = 0.55;
			// If the tile has a texture image, draw it; otherwise use solid color
			if (activeTile.visual.texture) {
				const img = textureImage(activeTile.visual.texture);
				if (img.complete) {
					ctx.imageSmoothingEnabled = false;
					ctx.globalAlpha = 0.55;
					ctx.drawImage(img, px, py, drawWidth, drawHeight);
					ctx.globalAlpha = 1;
				} else if (options.onTextureLoad) {
					if (!img.dataset.ready) {
						img.dataset.ready = '1';
						img.onload = () => options.onTextureLoad?.();
					}
				}
			} else {
				ctx.fillStyle = activeTile.visual.color;
				ctx.fillRect(px, py, drawWidth, drawHeight);

				if (showTilesetArt && tileset?.image?.complete) {
					drawTileArt(ctx, activeTileId, px, py, drawWidth, drawHeight, tileset);
				}
				ctx.globalAlpha = 1;
			}
			ctx.globalAlpha = 1;
			ctx.strokeStyle = ACCENT;
			ctx.lineWidth = 1.5;
			ctx.strokeRect(px + 0.5, py + 0.5, drawWidth - 1, drawHeight - 1);
		}
	} else {
		ctx.strokeStyle = 'rgba(255, 255, 255, 0.35)';
		ctx.lineWidth = 1;
		ctx.strokeRect(px + 0.5, py + 0.5, drawWidth - 1, drawHeight - 1);
	}
	ctx.restore();
}
