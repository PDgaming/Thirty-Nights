import type { CellPosition } from '../map-model/types';

export interface Viewport {
	/** Screen pixels per map pixel. */
	scale: number;
	/** Screen-space offset in CSS pixels. */
	offsetX: number;
	offsetY: number;
}

export const MIN_SCALE = 0.02;
export const MAX_SCALE = 64;

export function clampScale(scale: number): number {
	return Math.min(MAX_SCALE, Math.max(MIN_SCALE, scale));
}

export function screenToWorld(
	screenX: number,
	screenY: number,
	view: Viewport
): { x: number; y: number } {
	return {
		x: (screenX - view.offsetX) / view.scale,
		y: (screenY - view.offsetY) / view.scale
	};
}

export function screenToCell(
	screenX: number,
	screenY: number,
	view: Viewport,
	tileSize: number
): CellPosition {
	const world = screenToWorld(screenX, screenY, view);
	return {
		x: Math.floor(world.x / tileSize),
		y: Math.floor(world.y / tileSize)
	};
}

export function cellToScreen(
	cell: CellPosition,
	view: Viewport,
	tileSize: number
): { x: number; y: number } {
	return {
		x: view.offsetX + cell.x * tileSize * view.scale,
		y: view.offsetY + cell.y * tileSize * view.scale
	};
}

/** Zoom while keeping the map point under the pointer stationary. */
export function zoomAt(view: Viewport, factor: number, screenX: number, screenY: number): Viewport {
	const scale = clampScale(view.scale * factor);
	const worldX = (screenX - view.offsetX) / view.scale;
	const worldY = (screenY - view.offsetY) / view.scale;
	return {
		scale,
		offsetX: screenX - worldX * scale,
		offsetY: screenY - worldY * scale
	};
}

export function fitViewport(
	mapWidth: number,
	mapHeight: number,
	tileSize: number,
	viewportWidth: number,
	viewportHeight: number,
	padding = 32
): Viewport {
	const contentWidth = Math.max(1, mapWidth * tileSize);
	const contentHeight = Math.max(1, mapHeight * tileSize);
	const availableWidth = Math.max(1, viewportWidth - padding * 2);
	const availableHeight = Math.max(1, viewportHeight - padding * 2);
	const scale = clampScale(
		Math.min(availableWidth / contentWidth, availableHeight / contentHeight)
	);

	return {
		scale,
		offsetX: (viewportWidth - contentWidth * scale) / 2,
		offsetY: (viewportHeight - contentHeight * scale) / 2
	};
}
