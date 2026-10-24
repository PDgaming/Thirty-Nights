import { createEmptyMap, createTileDef, type NewMapOptions } from './map-model/factory';
import type { CollisionName, GameMap, LayerData, TileDef } from './map-model/types';
import {
	addLayer,
	addTile,
	countPopulatedCells,
	countTileReferences,
	describeResizeLoss,
	findLayer,
	inBounds,
	lineBetween,
	moveLayer as moveLayerOperation,
	removeLayer as removeLayerOperation,
	removeTile,
	renameLayer as renameLayerOperation,
	renameTile,
	resizeLossLabels,
	resizeMap as resizeMapOperation,
	setCell,
	setLayerCollision,
	tileIsReferencedByObject,
	uniqueLayerId,
	uniqueTileId,
	updateTile
} from './map-operations';
import { hasBlockingErrors, validateMap, type ValidationIssue } from './map-validation';
import { mapToJson, normalizeFilename } from './map-serializer';
import { decodeImageFile, downloadJson, extractSliceColors } from './file-io';
import {
	computeTilesetSlices,
	MAX_TILESET_SLICES,
	type TilesetConfig,
	type TilesetSlice
} from './tileset/slicing';
import type { EditorTool, TilesetRenderData } from './renderer/draw-map';
import { fitViewport, zoomAt as zoomAtOperation, type Viewport } from './renderer/viewport';

export type EditorTab = 'tileset' | 'inspector';
export type NoticeKind = 'info' | 'success' | 'error';

export interface Notice {
	kind: NoticeKind;
	text: string;
}

export interface TilesetImageState {
	image: HTMLImageElement | null;
	objectUrl: string | null;
	fileName: string;
	width: number;
	height: number;
}

const DEFAULT_SLICING = {
	tileWidth: 16,
	tileHeight: 16,
	spacing: 0,
	margin: 0
};

export class EditorState {
	/** Authoritative map data. Mutated through operations, then committed. */
	map = $state.raw<GameMap>(createEmptyMap());
	mapRevision = $state(0);
	/** Bumped when the map is replaced or resized, so forms can resync. */
	mapEpoch = $state(0);

	activeLayerId = $state('floor');
	activeTileId = $state<string | null>(null);
	tool = $state<EditorTool>('paint');
	tab = $state<EditorTab>('tileset');
	hiddenLayers = $state<string[]>([]);

	view = $state<Viewport>({ scale: 1, offsetX: 0, offsetY: 0 });
	viewportWidth = $state(0);
	viewportHeight = $state(0);
	hover = $state<{ x: number; y: number } | null>(null);
	showGrid = $state(true);
	showTilesetArt = $state(true);

	tileset = $state<TilesetImageState>({
		image: null,
		objectUrl: null,
		fileName: '',
		width: 0,
		height: 0
	});
	slicing = $state({ ...DEFAULT_SLICING });
	tilesetSlices = $state.raw<TilesetSlice[]>([]);
	tileBindings = $state.raw<Record<string, number>>({});
	/** Last colour sampled from the tileset for auto-generated definitions. */
	autoTileColors = $state.raw<Record<string, string>>({});

	notice = $state<Notice | null>(null);
	private noticeTimer: ReturnType<typeof setTimeout> | null = null;
	private hasFitted = false;

	get activeLayer(): LayerData | undefined {
		return findLayer(this.map, this.activeLayerId) ?? this.map.layers[0];
	}

	get activeTile(): TileDef | undefined {
		return this.activeTileId ? this.map.tiles[this.activeTileId] : undefined;
	}

	get activeLayerIsVisible(): boolean {
		return !this.hiddenLayers.includes(this.activeLayerId);
	}

	get tilesetRenderData(): TilesetRenderData | null {
		if (!this.tileset.image) return null;
		return {
			image: this.tileset.image,
			slices: this.tilesetSlices,
			bindings: this.tileBindings
		};
	}

	get issues(): ValidationIssue[] {
		return validateMap(this.map);
	}

	get canExport(): boolean {
		return !hasBlockingErrors(this.issues);
	}

	get suggestedFilename(): string {
		return normalizeFilename(this.map.id, 'map');
	}

	get sliceForActiveTile(): TilesetSlice | null {
		return this.sliceForTile(this.activeTileId);
	}

	sliceForTile(tileID: string | null): TilesetSlice | null {
		if (!tileID) return null;
		const index = this.tileBindings[tileID];
		if (index === undefined) return null;
		return this.tilesetSlices[index] ?? null;
	}

	get activeLayerCellCount(): number {
		return this.activeLayer ? countPopulatedCells(this.activeLayer) : 0;
	}

	notify(text: string, kind: NoticeKind = 'info'): void {
		this.notice = { kind, text };
		if (this.noticeTimer) clearTimeout(this.noticeTimer);
		this.noticeTimer = setTimeout(
			() => {
				this.notice = null;
				this.noticeTimer = null;
			},
			kind === 'error' ? 8000 : 4000
		);
	}

	dismissNotice(): void {
		if (this.noticeTimer) clearTimeout(this.noticeTimer);
		this.noticeTimer = null;
		this.notice = null;
	}

	private commit(): void {
		this.map = { ...this.map };
		this.mapRevision += 1;
	}

	// ---------------------------------------------------------------- viewport

	setViewportSize(width: number, height: number): void {
		this.viewportWidth = width;
		this.viewportHeight = height;
		if (!this.hasFitted && width > 0 && height > 0) {
			this.hasFitted = true;
			this.fitView();
		}
	}

	fitView(): void {
		if (this.viewportWidth <= 0 || this.viewportHeight <= 0) return;
		this.view = fitViewport(
			this.map.width,
			this.map.height,
			this.map.tile_size,
			this.viewportWidth,
			this.viewportHeight
		);
	}

	zoomAt(factor: number, screenX: number, screenY: number): void {
		this.view = zoomAtOperation(this.view, factor, screenX, screenY);
	}

	zoomBy(factor: number): void {
		this.zoomAt(factor, this.viewportWidth / 2, this.viewportHeight / 2);
	}

	setZoomPercent(percent: number): void {
		if (!Number.isFinite(percent) || percent <= 0) return;
		const target = percent / 100;
		const factor = target / this.view.scale;
		this.zoomBy(factor);
	}

	panBy(dx: number, dy: number): void {
		this.view = {
			...this.view,
			offsetX: this.view.offsetX + dx,
			offsetY: this.view.offsetY + dy
		};
	}

	// -------------------------------------------------------------- map fields

	setMapId(id: string): void {
		this.map.id = id;
		this.commit();
	}

	setMapName(name: string): void {
		this.map.name = name;
		this.commit();
	}

	setTileSize(tileSize: number): void {
		if (!Number.isInteger(tileSize) || tileSize <= 0) {
			this.notify('Tile size must be a positive integer', 'error');
			return;
		}
		this.map.tile_size = tileSize;
		this.commit();
	}

	planResize(width: number, height: number) {
		return describeResizeLoss(this.map, width, height);
	}

	applyResize(width: number, height: number): boolean {
		if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0) {
			this.notify('Map dimensions must be positive integers', 'error');
			return false;
		}
		const result = resizeMapOperation(this.map, width, height);
		this.mapEpoch += 1;
		this.commit();
		if (this.activeLayerId && !findLayer(this.map, this.activeLayerId)) {
			this.activeLayerId = this.map.layers[0]?.id ?? '';
		}
		const losses = resizeLossLabels(result);
		this.notify(
			losses.length > 0
				? `Resized map. ${losses.join(', ')}.`
				: `Resized map to ${width}x${height}.`,
			losses.length > 0 ? 'info' : 'success'
		);
		this.fitView();
		return true;
	}

	setSpawn(x: number, y: number): void {
		if (!Number.isFinite(x) || !Number.isFinite(y)) return;
		this.map.spawn = { x: Math.trunc(x), y: Math.trunc(y) };
		this.commit();
	}

	setSpawnFromCell(x: number, y: number): void {
		if (!inBounds(this.map, x, y)) return;
		this.map.spawn = { x, y };
		this.commit();
	}

	newMap(options: NewMapOptions = {}): void {
		this.map = createEmptyMap(options);
		this.mapEpoch += 1;
		this.mapRevision += 1;
		this.activeLayerId = this.map.layers[0].id;
		this.activeTileId = null;
		this.autoTileColors = {};
		this.hiddenLayers = [];
		this.hover = null;
		this.tool = 'paint';
		this.hasFitted = false;
		this.fitView();
		if (this.tileset.image) {
			this.applyTilesetSlicing(false);
		}
		this.notify('Created a new map', 'success');
	}

	// ------------------------------------------------------------------ layers

	selectLayer(id: string): void {
		if (!findLayer(this.map, id)) return;
		this.activeLayerId = id;
	}

	isLayerHidden(id: string): boolean {
		return this.hiddenLayers.includes(id);
	}

	toggleLayerVisibility(id: string): void {
		if (!findLayer(this.map, id)) return;
		this.hiddenLayers = this.isLayerHidden(id)
			? this.hiddenLayers.filter((layerId) => layerId !== id)
			: [...this.hiddenLayers, id];
	}

	createLayer(id: string, collision: CollisionName): boolean {
		const trimmed = id.trim();
		if (trimmed === '') {
			this.notify('Layer ID is required', 'error');
			return false;
		}
		const created = addLayer(this.map, trimmed, collision);
		if (!created) {
			this.notify(`A layer with ID "${trimmed}" already exists`, 'error');
			return false;
		}
		this.activeLayerId = created.id;
		this.commit();
		this.notify(`Added layer "${created.id}"`, 'success');
		return true;
	}

	suggestLayerId(): string {
		return uniqueLayerId(this.map);
	}

	renameLayer(oldID: string, newID: string): boolean {
		const trimmed = newID.trim();
		if (trimmed === oldID) return true;
		if (!renameLayerOperation(this.map, oldID, trimmed)) {
			this.notify('Layer ID is empty or already in use', 'error');
			return false;
		}
		if (this.activeLayerId === oldID) this.activeLayerId = trimmed;
		this.commit();
		return true;
	}

	deleteLayer(id: string): boolean {
		const layer = findLayer(this.map, id);
		if (!layer) return false;
		if (this.map.layers.length <= 1) {
			this.notify('A map must keep at least one layer', 'error');
			return false;
		}
		if (!removeLayerOperation(this.map, id)) return false;
		this.hiddenLayers = this.hiddenLayers.filter((layerId) => layerId !== id);
		if (this.activeLayerId === id) {
			this.activeLayerId = this.map.layers[0].id;
		}
		this.commit();
		return true;
	}

	moveLayer(id: string, direction: number): void {
		if (moveLayerOperation(this.map, id, direction)) {
			this.commit();
		}
	}

	setLayerCollision(id: string, collision: CollisionName): void {
		if (setLayerCollision(this.map, id, collision)) {
			this.commit();
		}
	}

	layerCellCount(id: string): number {
		const layer = findLayer(this.map, id);
		return layer ? countPopulatedCells(layer) : 0;
	}

	// ------------------------------------------------------------------- tiles

	selectTile(id: string): void {
		if (!this.map.tiles[id]) return;
		this.activeTileId = id;
		this.tool = 'paint';
	}

	findTileForSlice(index: number): string | null {
		for (const [tileID, sliceIndex] of Object.entries(this.tileBindings)) {
			if (sliceIndex === index) return tileID;
		}
		return null;
	}

	selectSlice(index: number): void {
		const slice = this.tilesetSlices[index];
		if (!slice) return;

		const existing = this.findTileForSlice(index);
		if (existing) {
			this.activeTileId = existing;
			this.tool = 'paint';
			return;
		}

		const tileID = this.createTileForSlice(slice);
		if (tileID) {
			this.activeTileId = tileID;
			this.tool = 'paint';
		}
	}

	private createTileForSlice(slice: TilesetSlice): string | null {
		if (!this.tileset.image) return null;
		const tileID = uniqueTileId(this.map, `tile_${slice.row}_${slice.column}`);
		const colors = extractSliceColors(this.tileset.image, [slice]);
		const color = colors[slice.index] ?? '#7a828c';
		this.map.tiles = {
			...this.map.tiles,
			[tileID]: createTileDef(color)
		};
		this.tileBindings = { ...this.tileBindings, [tileID]: slice.index };
		this.autoTileColors = { ...this.autoTileColors, [tileID]: color };
		this.commit();
		return tileID;
	}

	addTileDefinition(id: string, color: string, solid = false): boolean {
		const trimmed = id.trim();
		if (!addTile(this.map, trimmed, color, solid)) {
			this.notify(
				trimmed === '' ? 'Tile ID is required' : `Tile "${trimmed}" already exists`,
				'error'
			);
			return false;
		}
		this.activeTileId = trimmed;
		this.tool = 'paint';
		this.commit();
		this.notify(`Added tile "${trimmed}"`, 'success');
		return true;
	}

	setTileSolid(id: string, solid: boolean): void {
		if (updateTile(this.map, id, { solid })) this.commit();
	}

	setTileColor(id: string, color: string): void {
		if (!updateTile(this.map, id, { visual: { color } })) return;
		if (this.autoTileColors[id] !== undefined) {
			const next = { ...this.autoTileColors };
			delete next[id];
			this.autoTileColors = next;
		}
		this.commit();
	}

	renameActiveTile(newID: string): boolean {
		const oldID = this.activeTileId;
		if (!oldID) return false;
		const trimmed = newID.trim();
		if (trimmed === oldID) return true;
		if (trimmed === '') {
			this.notify('Tile ID is required', 'error');
			return false;
		}
		if (this.map.tiles[trimmed]) {
			this.notify(`Tile "${trimmed}" already exists`, 'error');
			return false;
		}
		if (tileIsReferencedByObject(this.map, oldID)) {
			this.notify(
				`Tile "${oldID}" is the type of one or more objects and cannot be renamed`,
				'error'
			);
			return false;
		}
		if (!renameTile(this.map, oldID, trimmed)) return false;

		const index = this.tileBindings[oldID];
		if (index !== undefined) {
			const next = { ...this.tileBindings };
			delete next[oldID];
			next[trimmed] = index;
			this.tileBindings = next;
		}
		if (this.autoTileColors[oldID] !== undefined) {
			const nextAuto = { ...this.autoTileColors };
			nextAuto[trimmed] = nextAuto[oldID];
			delete nextAuto[oldID];
			this.autoTileColors = nextAuto;
		}
		this.activeTileId = trimmed;
		this.commit();
		return true;
	}

	deleteActiveTile(): boolean {
		const id = this.activeTileId;
		if (!id) return false;
		const references = countTileReferences(this.map, id);
		if (references > 0) {
			this.notify(
				`Cannot delete "${id}": ${references} map cell${references === 1 ? '' : 's'} use it`,
				'error'
			);
			return false;
		}
		if (tileIsReferencedByObject(this.map, id)) {
			this.notify(`Cannot delete "${id}": it is used by an object type`, 'error');
			return false;
		}
		if (!removeTile(this.map, id)) return false;
		const next = { ...this.tileBindings };
		delete next[id];
		this.tileBindings = next;
		if (this.autoTileColors[id] !== undefined) {
			const nextAuto = { ...this.autoTileColors };
			delete nextAuto[id];
			this.autoTileColors = nextAuto;
		}
		this.activeTileId = null;
		this.commit();
		this.notify(`Deleted tile "${id}"`, 'success');
		return true;
	}

	tileReferenceCount(id: string): number {
		return countTileReferences(this.map, id);
	}

	// ----------------------------------------------------------------- tileset

	async importTileset(file: File): Promise<void> {
		try {
			const loaded = await decodeImageFile(file);
			if (this.tileset.objectUrl) URL.revokeObjectURL(this.tileset.objectUrl);
			this.tileset = {
				image: loaded.image,
				objectUrl: loaded.objectUrl,
				fileName: file.name,
				width: loaded.width,
				height: loaded.height
			};
			this.slicing.tileWidth = Math.min(Math.max(1, this.slicing.tileWidth), loaded.width);
			this.slicing.tileHeight = Math.min(Math.max(1, this.slicing.tileHeight), loaded.height);
			this.applyTilesetSlicing(false);
			this.notify(`Imported ${file.name} (${loaded.width}x${loaded.height})`, 'success');
		} catch (error) {
			this.notify(error instanceof Error ? error.message : 'Could not import image', 'error');
		}
	}

	get tilesetConfig(): TilesetConfig {
		return {
			imageWidth: this.tileset.width,
			imageHeight: this.tileset.height,
			tileWidth: this.slicing.tileWidth,
			tileHeight: this.slicing.tileHeight,
			spacing: this.slicing.spacing,
			margin: this.slicing.margin
		};
	}

	getPreviewSlices(): TilesetSlice[] {
		return computeTilesetSlices(this.tilesetConfig);
	}

	applyTilesetSlicing(announce = true): boolean {
		const image = this.tileset.image;
		if (!image) {
			this.notify('Import a tileset image first', 'error');
			return false;
		}
		const config = this.tilesetConfig;
		const slices = computeTilesetSlices(config);
		if (slices.length === 0) {
			this.notify('No tiles fit inside the image with these slicing settings', 'error');
			return false;
		}
		if (slices.length > MAX_TILESET_SLICES) {
			this.notify(
				`This slicing produces ${slices.length} tiles; the limit is ${MAX_TILESET_SLICES}`,
				'error'
			);
			return false;
		}

		const colors = extractSliceColors(image, slices);
		const reverse: Record<number, string> = {};
		for (const [tileID, index] of Object.entries(this.tileBindings)) {
			if (this.map.tiles[tileID]) reverse[index] = tileID;
		}

		const nextTiles = { ...this.map.tiles };
		const nextBindings: Record<string, number> = {};
		const nextAutoColors = { ...this.autoTileColors };

		for (const slice of slices) {
			const sampled = colors[slice.index] ?? '#7a828c';
			let tileID = reverse[slice.index];
			if (!tileID || !nextTiles[tileID]) {
				tileID = uniqueTileId(
					{ ...this.map, tiles: nextTiles },
					`tile_${slice.row}_${slice.column}`
				);
				nextTiles[tileID] = createTileDef(sampled);
				nextAutoColors[tileID] = sampled;
			} else {
				const previousAuto = nextAutoColors[tileID];
				const current = nextTiles[tileID];
				if (previousAuto !== undefined && current.visual.color === previousAuto) {
					nextTiles[tileID] = { ...current, visual: { color: sampled } };
					nextAutoColors[tileID] = sampled;
				}
			}
			nextBindings[tileID] = slice.index;
		}

		// Drop auto-generated tiles that no longer match a slice and are not in
		// use, so repeated slicing does not accumulate dead definitions.
		for (const [tileID] of Object.entries(nextAutoColors)) {
			if (nextBindings[tileID] !== undefined) continue;
			if (countTileReferences(this.map, tileID) > 0) continue;
			if (tileIsReferencedByObject(this.map, tileID)) continue;
			delete nextTiles[tileID];
			delete nextAutoColors[tileID];
		}

		this.map.tiles = nextTiles;
		this.tileBindings = nextBindings;
		this.autoTileColors = nextAutoColors;
		this.tilesetSlices = slices;
		if (this.activeTileId && !nextTiles[this.activeTileId]) {
			this.activeTileId = null;
		}
		this.commit();
		if (announce) {
			this.notify(`Sliced tileset into ${slices.length} tiles`, 'success');
		}
		return true;
	}

	// ----------------------------------------------------------------- painting

	paintLine(from: { x: number; y: number }, to: { x: number; y: number }, erase: boolean): void {
		const layer = findLayer(this.map, this.activeLayerId);
		if (!layer) return;
		if (!erase && (!this.activeTileId || !this.map.tiles[this.activeTileId])) return;

		let changed = false;
		for (const cell of lineBetween(from, to)) {
			if (!inBounds(this.map, cell.x, cell.y)) continue;
			const tileID = erase ? '' : (this.activeTileId as string);
			if (layer.grid[cell.y][cell.x] !== tileID) {
				setCell(this.map, layer.id, cell.x, cell.y, tileID);
				changed = true;
			}
		}
		if (changed) this.commit();
	}

	canPaint(): boolean {
		if (!this.activeLayer) return false;
		if (!this.activeLayerIsVisible) {
			this.notify('The active layer is hidden. Show it before painting.', 'error');
			return false;
		}
		return true;
	}

	// ------------------------------------------------------------------- export

	exportMap(filename: string): boolean {
		const issues = this.issues;
		if (hasBlockingErrors(issues)) {
			const first = issues.find((issue) => issue.severity === 'error');
			this.notify(
				`Cannot export: ${first ? `${first.path} ${first.message}` : 'map is invalid'}`,
				'error'
			);
			return false;
		}

		const safeName = normalizeFilename(filename, this.suggestedFilename);
		downloadJson(safeName, mapToJson(this.map));
		this.notify(`Exported ${safeName}`, 'success');
		return true;
	}
}
