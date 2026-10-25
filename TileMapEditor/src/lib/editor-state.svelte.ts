import { createEmptyMap, createTileDef, type NewMapOptions } from './map-model/factory';
import type { CollisionName, GameMap, LayerData, TileDef } from './map-model/types';
import {
	addLayer,
	addObject,
	addTile,
	countObjectsOfType,
	countPopulatedCells,
	countTileReferences,
	describeResizeLoss,
	findLayer,
	inBounds,
	lineBetween,
	moveLayer as moveLayerOperation,
	removeLayer as removeLayerOperation,
	removeObjectsAt,
	removeTile,
	reorderLayers as reorderLayersOperation,
	renameLayer as renameLayerOperation,
	renameObjectsOfType,
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
import {
	hasBlockingErrors,
	isValidHexColor,
	validateMap,
	type ValidationIssue
} from './map-validation';
import { mapToJson, normalizeFilename } from './map-serializer';
import {
	DEFAULT_MAP_SCHEMA,
	mapToJsonWithSchema,
	parseMapSchema,
	validateMapSchema,
	type MapSchema
} from './map-schema';
import { parseMapJson } from './map-importer';
import { decodeImageFile, downloadJson, extractSliceColors } from './file-io';
import {
	computeTilesetSlices,
	MAX_TILESET_SLICES,
	type TilesetConfig,
	type TilesetSlice
} from './tileset/slicing';
import type { EditorTool, TilesetRenderData } from './renderer/draw-map';
import { fitViewport, zoomAt as zoomAtOperation, type Viewport } from './renderer/viewport';

export type EditorTab = 'tileset' | 'inspector' | 'objects';

const EDITOR_TABS: readonly EditorTab[] = ['tileset', 'inspector', 'objects'];
export const TAB_STORAGE_KEY = 'tilemap-editor.tab';
export const SELECTION_STORAGE_KEY = 'tilemap-editor.selection';
export const MAP_SCHEMA_STORAGE_KEY = 'tilemap-editor.map-schema';

function isEditorTab(value: unknown): value is EditorTab {
	return typeof value === 'string' && (EDITOR_TABS as readonly string[]).includes(value);
}

interface SavedSelection {
	kind: 'tile' | 'object';
	id: string;
}

function parseSavedSelection(raw: string): SavedSelection | null {
	try {
		const value: unknown = JSON.parse(raw);
		if (typeof value !== 'object' || value === null) return null;
		const { kind, id } = value as Record<string, unknown>;
		if ((kind === 'tile' || kind === 'object') && typeof id === 'string' && id !== '') {
			return { kind, id };
		}
	} catch {
		// Malformed JSON is treated as no saved selection.
	}
	return null;
}
export type NoticeKind = 'info' | 'success' | 'error';

export interface Notice {
	kind: NoticeKind;
	text: string;
}

/** Editor-side object palette entry. Object types are not exported. */
export interface ObjectTypeDef {
	id: string;
	color: string;
}

const DEFAULT_OBJECT_TYPES: ObjectTypeDef[] = [{ id: 'door', color: '#9a7047' }];
const OBJECT_MARKER_COLORS = ['#f472b6', '#60a5fa', '#4ade80', '#c084fc', '#facc15', '#fb923c'];

/** Rebuild the object palette from the types used by an imported map. */
function objectTypesForMap(map: GameMap): ObjectTypeDef[] {
	const ids: string[] = [];
	for (const object of map.objects) {
		if (!ids.includes(object.type)) ids.push(object.type);
	}
	if (ids.length === 0) ids.push('door');
	return ids.map((id, index) => ({
		id,
		color: map.tiles[id]?.visual.color ?? OBJECT_MARKER_COLORS[index % OBJECT_MARKER_COLORS.length]
	}));
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
	marginX: 0,
	marginY: 0
};

const MAX_UNDO_STEPS = 100;

/** Survives reloads and browser restarts, so work outlives a single tab. */
export const MAP_STORAGE_KEY = 'thirty-nights.tilemap-editor.current-map';

/** Edits are batched into at most one storage write per window. */
export const AUTOSAVE_DELAY_MS = 5000;
export const MIN_AUTOSAVE_DELAY_MS = 1000;
export const MAX_AUTOSAVE_DELAY_MS = 300000;
const AUTOSAVE_PREFERENCE_KEY = 'thirty-nights.tilemap-editor.autosave-delay';

/** Everything that should move together when undoing an edit. */
interface UndoSnapshot {
	map: GameMap;
	tileBindings: Record<string, number>;
	autoTileColors: Record<string, string>;
	objectTypes: ObjectTypeDef[];
}

export class EditorState {
	/** Authoritative map data. Mutated through operations, then committed. */
	map = $state.raw<GameMap>(createEmptyMap());
	mapRevision = $state(0);
	/** Bumped when the map is replaced or resized, so forms can resync. */
	mapEpoch = $state(0);

	activeLayerId = $state('floor');
	#activeTileId = $state<string | null>(null);
	#activeObjectTypeId = $state<string | null>(null);
	objectTypes = $state.raw<ObjectTypeDef[]>(DEFAULT_OBJECT_TYPES.map((type) => ({ ...type })));
	tool = $state<EditorTool>('paint');
	tab = $state<EditorTab>('tileset');
	hiddenLayers = $state<string[]>([]);

	/** Assigning a tile clears any object selection, and vice versa. */
	get activeTileId(): string | null {
		return this.#activeTileId;
	}

	set activeTileId(id: string | null) {
		this.#activeTileId = id;
		if (id !== null) this.#activeObjectTypeId = null;
		this.persistSelection();
	}

	get activeObjectTypeId(): string | null {
		return this.#activeObjectTypeId;
	}

	set activeObjectTypeId(id: string | null) {
		this.#activeObjectTypeId = id;
		if (id !== null) this.#activeTileId = null;
		this.persistSelection();
	}

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
	autosaveDelayMs = $state(AUTOSAVE_DELAY_MS);
	mapSchema = $state.raw<MapSchema>(DEFAULT_MAP_SCHEMA);
	private noticeTimer: ReturnType<typeof setTimeout> | null = null;
	private autosaveTimer: ReturnType<typeof setTimeout> | null = null;
	private autosaveDirty = false;
	private hasFitted = false;

	private undoStack: UndoSnapshot[] = [];
	private redoStack: UndoSnapshot[] = [];
	private undoGroupActive = false;
	private undoGroupCaptured = false;
	/** Number of undoable steps, kept reactive for the toolbar button. */
	undoDepth = $state(0);
	/** Number of redoable steps, kept reactive for the toolbar button. */
	redoDepth = $state(0);

	constructor() {
		this.autosaveDelayMs = this.loadAutosaveDelay();
		this.mapSchema = this.loadMapSchema();
		const restored = this.restoreMap();
		if (restored) {
			this.map = restored;
			this.mapEpoch += 1;
			this.mapRevision += 1;
			this.activeLayerId = restored.layers[0]?.id ?? 'floor';
			this.objectTypes = objectTypesForMap(restored);
		}
		this.tab = this.restoreTab() ?? this.tab;
		this.restoreSelection();
		this.resetUndoHistory();
	}

	private localStorage(): Storage | null {
		try {
			if (typeof window === 'undefined') return null;
			return window.localStorage;
		} catch {
			// Storage can be unavailable in privacy-restricted contexts.
			return null;
		}
	}

	/** Returns null when storage is unavailable or holds an unknown tab. */
	private restoreTab(): EditorTab | null {
		try {
			const value = this.localStorage()?.getItem(TAB_STORAGE_KEY);
			return isEditorTab(value) ? value : null;
		} catch {
			return null;
		}
	}

	/** Best-effort persistence; blocked storage should never break tab switching. */
	private persistTab(tab: EditorTab): void {
		try {
			this.localStorage()?.setItem(TAB_STORAGE_KEY, tab);
		} catch {
			// Ignore quota and privacy errors.
		}
	}

	private loadAutosaveDelay(): number {
		try {
			const raw = this.localStorage()?.getItem(AUTOSAVE_PREFERENCE_KEY);
			const value = raw == null ? NaN : Number(raw);
			if (!Number.isFinite(value)) return AUTOSAVE_DELAY_MS;
			return Math.min(MAX_AUTOSAVE_DELAY_MS, Math.max(MIN_AUTOSAVE_DELAY_MS, Math.round(value)));
		} catch {
			return AUTOSAVE_DELAY_MS;
		}
	}

	/** Change how often edits are written to browser storage. */
	setAutosaveDelay(delayMs: number): void {
		if (!Number.isFinite(delayMs)) return;
		const clamped = Math.min(
			MAX_AUTOSAVE_DELAY_MS,
			Math.max(MIN_AUTOSAVE_DELAY_MS, Math.round(delayMs))
		);
		if (clamped === this.autosaveDelayMs) return;

		this.autosaveDelayMs = clamped;
		try {
			this.localStorage()?.setItem(AUTOSAVE_PREFERENCE_KEY, String(clamped));
		} catch {
			// Ignore quota and privacy errors.
		}

		// Apply the new delay to a pending save instead of waiting out the old one.
		if (this.autosaveTimer !== null) {
			clearTimeout(this.autosaveTimer);
			this.autosaveTimer = null;
			if (this.autosaveDirty) this.schedulePersist();
		}
	}

	private loadMapSchema(): MapSchema {
		try {
			const raw = this.localStorage()?.getItem(MAP_SCHEMA_STORAGE_KEY);
			if (!raw) return DEFAULT_MAP_SCHEMA;
			return parseMapSchema(raw);
		} catch {
			return DEFAULT_MAP_SCHEMA;
		}
	}

	/** Replace the export schema. Throws if compulsory fields are removed. */
	setMapSchema(schema: MapSchema): void {
		validateMapSchema(schema);
		this.mapSchema = schema;
		try {
			this.localStorage()?.setItem(MAP_SCHEMA_STORAGE_KEY, JSON.stringify(schema));
		} catch {
			// Ignore quota and privacy errors.
		}
	}

	resetMapSchema(): void {
		this.setMapSchema(DEFAULT_MAP_SCHEMA);
	}

	/** Switches the right-hand panel and remembers the choice for the next visit. */
	selectTab(tab: EditorTab): void {
		if (this.tab === tab) return;
		this.tab = tab;
		this.persistTab(tab);
	}

	/** Remembers the active tile or object type; never both at once. */
	private persistSelection(): void {
		const storage = this.localStorage();
		if (!storage) return;
		try {
			if (this.activeObjectTypeId) {
				storage.setItem(
					SELECTION_STORAGE_KEY,
					JSON.stringify({ kind: 'object', id: this.activeObjectTypeId })
				);
			} else if (this.activeTileId) {
				storage.setItem(
					SELECTION_STORAGE_KEY,
					JSON.stringify({ kind: 'tile', id: this.activeTileId })
				);
			} else {
				storage.removeItem(SELECTION_STORAGE_KEY);
			}
		} catch {
			// Ignore quota and privacy errors.
		}
	}

	private restoreSelection(): void {
		const storage = this.localStorage();
		if (!storage) return;

		let saved: SavedSelection | null = null;
		try {
			const raw = storage.getItem(SELECTION_STORAGE_KEY);
			saved = raw ? parseSavedSelection(raw) : null;
		} catch {
			return;
		}
		if (!saved) return;

		if (saved.kind === 'tile' && this.map.tiles[saved.id]) {
			this.selectTile(saved.id);
			return;
		}
		if (saved.kind === 'object' && this.objectTypes.some((type) => type.id === saved.id)) {
			this.selectObjectType(saved.id);
			return;
		}

		// The saved target no longer exists in the restored map or palette.
		try {
			storage.removeItem(SELECTION_STORAGE_KEY);
		} catch {
			// Ignore cleanup failures.
		}
	}

	private restoreMap(): GameMap | null {
		const storage = this.localStorage();
		if (!storage) return null;
		try {
			let raw = storage.getItem(MAP_STORAGE_KEY);
			if (!raw) {
				// Older builds kept the map in session storage. Move it across so
				// upgrading does not lose the map that is currently open.
				raw = this.takeLegacySessionMap();
				if (raw) {
					try {
						storage.setItem(MAP_STORAGE_KEY, raw);
					} catch {
						// The legacy map still restores even if the copy fails.
					}
				}
			}
			if (!raw) return null;
			return parseMapJson(raw);
		} catch {
			try {
				storage.removeItem(MAP_STORAGE_KEY);
			} catch {
				// Ignore cleanup failures.
			}
			return null;
		}
	}

	/** Read and clear a map left behind in session storage by an older build. */
	private takeLegacySessionMap(): string | null {
		try {
			if (typeof window === 'undefined') return null;
			const session = window.sessionStorage;
			const raw = session?.getItem(MAP_STORAGE_KEY) ?? null;
			if (raw) session.removeItem(MAP_STORAGE_KEY);
			return raw;
		} catch {
			return null;
		}
	}

	/**
	 * Queue a batched autosave. The first edit starts the timer; further edits
	 * share it, so storage is written at most once per configured delay.
	 */
	private schedulePersist(): void {
		this.autosaveDirty = true;
		if (!this.localStorage() || this.autosaveTimer !== null) return;
		this.autosaveTimer = setTimeout(() => {
			this.autosaveTimer = null;
			this.flushPendingSave();
		}, this.autosaveDelayMs);
	}

	/** Write any pending changes now, e.g. before the page unloads. */
	flushPendingSave(): void {
		if (this.autosaveTimer !== null) {
			clearTimeout(this.autosaveTimer);
			this.autosaveTimer = null;
		}
		if (!this.autosaveDirty) return;
		this.autosaveDirty = false;
		this.persistMap();
	}

	/** Save the map immediately, e.g. from Ctrl+S. */
	saveNow(): void {
		if (this.autosaveTimer !== null) {
			clearTimeout(this.autosaveTimer);
			this.autosaveTimer = null;
		}
		this.autosaveDirty = false;
		const saved = this.persistMap();
		this.notify(
			saved ? 'Saved map to browser storage' : 'Could not save the map to browser storage',
			saved ? 'success' : 'error'
		);
	}

	/** Best-effort autosave; storage quota or privacy mode should never break editing. */
	private persistMap(): boolean {
		const storage = this.localStorage();
		if (!storage) return false;
		try {
			storage.setItem(MAP_STORAGE_KEY, mapToJson(this.map));
			return true;
		} catch {
			// Ignore quota and privacy errors.
			return false;
		}
	}

	get canUndo(): boolean {
		return this.undoDepth > 0;
	}

	get canRedo(): boolean {
		return this.redoDepth > 0;
	}

	/** Treat every commit until `endUndoGroup` as one undoable step. */
	beginUndoGroup(): void {
		this.undoGroupActive = true;
		this.undoGroupCaptured = false;
	}

	endUndoGroup(): void {
		this.undoGroupActive = false;
		this.undoGroupCaptured = false;
	}

	undo(): boolean {
		if (this.undoStack.length <= 1) return false;
		this.endUndoGroup();
		const current = this.undoStack.pop();
		if (current) this.pushRedo(current);
		this.undoDepth = this.undoStack.length - 1;
		this.applySnapshot(this.undoStack[this.undoStack.length - 1]);
		return true;
	}

	redo(): boolean {
		if (this.redoStack.length === 0) return false;
		this.endUndoGroup();
		const snapshot = this.redoStack.pop();
		if (!snapshot) return false;
		this.undoStack.push(snapshot);
		if (this.undoStack.length > MAX_UNDO_STEPS + 1) this.undoStack.shift();
		this.undoDepth = this.undoStack.length - 1;
		this.redoDepth = this.redoStack.length;
		this.applySnapshot(snapshot);
		return true;
	}

	private pushRedo(snapshot: UndoSnapshot): void {
		this.redoStack.push(snapshot);
		if (this.redoStack.length > MAX_UNDO_STEPS) this.redoStack.shift();
		this.redoDepth = this.redoStack.length;
	}

	private clearRedo(): void {
		if (this.redoStack.length === 0) return;
		this.redoStack = [];
		this.redoDepth = 0;
	}

	private applySnapshot(snapshot: UndoSnapshot): void {
		this.map = structuredClone(snapshot.map);
		this.tileBindings = { ...snapshot.tileBindings };
		this.autoTileColors = { ...snapshot.autoTileColors };
		this.objectTypes = snapshot.objectTypes.map((type) => ({ ...type }));
		this.mapEpoch += 1;
		this.mapRevision += 1;
		this.schedulePersist();

		if (this.activeTileId && !this.map.tiles[this.activeTileId]) this.activeTileId = null;
		if (!findLayer(this.map, this.activeLayerId)) {
			this.activeLayerId = this.map.layers[0]?.id ?? '';
		}
		if (
			this.activeObjectTypeId &&
			!this.objectTypes.some((type) => type.id === this.activeObjectTypeId)
		) {
			this.activeObjectTypeId = null;
		}
	}

	private snapshot(): UndoSnapshot {
		return {
			map: structuredClone(this.map),
			tileBindings: { ...this.tileBindings },
			autoTileColors: { ...this.autoTileColors },
			objectTypes: this.objectTypes.map((type) => ({ ...type }))
		};
	}

	private resetUndoHistory(): void {
		this.undoStack = [this.snapshot()];
		this.redoStack = [];
		this.undoGroupActive = false;
		this.undoGroupCaptured = false;
		this.undoDepth = 0;
		this.redoDepth = 0;
	}

	private recordUndoSnapshot(): void {
		this.clearRedo();
		if (this.undoGroupActive) {
			if (this.undoGroupCaptured) {
				this.undoStack[this.undoStack.length - 1] = this.snapshot();
				return;
			}
			this.undoGroupCaptured = true;
		}
		this.undoStack.push(this.snapshot());
		if (this.undoStack.length > MAX_UNDO_STEPS + 1) this.undoStack.shift();
		this.undoDepth = this.undoStack.length - 1;
	}

	get activeLayer(): LayerData | undefined {
		return findLayer(this.map, this.activeLayerId) ?? this.map.layers[0];
	}

	get activeTile(): TileDef | undefined {
		return this.activeTileId ? this.map.tiles[this.activeTileId] : undefined;
	}

	get activeObjectType(): ObjectTypeDef | undefined {
		return this.activeObjectTypeId
			? this.objectTypes.find((type) => type.id === this.activeObjectTypeId)
			: undefined;
	}

	get objectColors(): Record<string, string> {
		const colors: Record<string, string> = {};
		for (const type of this.objectTypes) colors[type.id] = type.color;
		return colors;
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
		this.recordUndoSnapshot();
		this.schedulePersist();
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
		this.activeObjectTypeId = null;
		this.objectTypes = DEFAULT_OBJECT_TYPES.map((type) => ({ ...type }));
		this.autoTileColors = {};
		this.hiddenLayers = [];
		this.hover = null;
		this.tool = 'paint';
		this.hasFitted = false;
		this.fitView();
		if (this.tileset.image) {
			this.applyTilesetSlicing(false);
		}
		this.resetUndoHistory();
		this.schedulePersist();
		this.notify('Created a new map', 'success');
	}

	loadMap(map: GameMap): void {
		this.map = map;
		this.mapEpoch += 1;
		this.mapRevision += 1;
		this.activeLayerId = map.layers[0]?.id ?? '';
		this.activeTileId = null;
		this.activeObjectTypeId = null;
		this.objectTypes = objectTypesForMap(map);
		this.autoTileColors = {};
		this.tileBindings = {};
		this.hiddenLayers = [];
		this.hover = null;
		this.tool = 'paint';
		this.hasFitted = false;
		this.fitView();
		this.resetUndoHistory();
		this.schedulePersist();
	}

	async importMapFile(file: File): Promise<boolean> {
		try {
			const map = parseMapJson(await file.text());
			this.loadMap(map);
			const errors = this.issues.filter((issue) => issue.severity === 'error').length;
			this.notify(
				`Imported "${file.name}" (${map.width}x${map.height}, ${map.layers.length} layer${
					map.layers.length === 1 ? '' : 's'
				}, ${map.objects.length} object${map.objects.length === 1 ? '' : 's'})${
					errors > 0 ? `; ${errors} validation error${errors === 1 ? '' : 's'}` : ''
				}`,
				errors > 0 ? 'info' : 'success'
			);
			return true;
		} catch (error) {
			this.notify(
				`Could not import "${file.name}": ${error instanceof Error ? error.message : 'unknown error'}`,
				'error'
			);
			return false;
		}
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

	/** `orderedIDs` is bottom-to-top, matching z-index order. */
	reorderLayers(orderedIDs: string[]): boolean {
		if (!reorderLayersOperation(this.map, orderedIDs)) return false;
		this.commit();
		return true;
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

	deleteTile(id: string): boolean {
		if (!this.map.tiles[id]) return false;
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
		if (this.activeTileId === id) this.activeTileId = null;
		this.commit();
		this.notify(`Deleted tile "${id}"`, 'success');
		return true;
	}

	tileReferenceCount(id: string): number {
		return countTileReferences(this.map, id);
	}

	tileIsUsedByObject(id: string): boolean {
		return tileIsReferencedByObject(this.map, id);
	}

	// ----------------------------------------------------------------- objects

	selectObjectType(id: string): void {
		if (!this.objectTypes.some((type) => type.id === id)) return;
		this.activeObjectTypeId = id;
		this.tool = 'object';
	}

	objectCountForType(id: string): number {
		return countObjectsOfType(this.map, id);
	}

	addObjectType(id: string, color: string): boolean {
		const trimmed = id.trim();
		if (trimmed === '') {
			this.notify('Object type ID is required', 'error');
			return false;
		}
		if (this.objectTypes.some((type) => type.id === trimmed)) {
			this.notify(`Object type "${trimmed}" already exists`, 'error');
			return false;
		}
		if (!isValidHexColor(color)) {
			this.notify('Object type colour must be #rgb, #rrggbb, or #rrggbbaa', 'error');
			return false;
		}
		this.objectTypes = [...this.objectTypes, { id: trimmed, color }];
		this.selectObjectType(trimmed);
		this.notify(`Added object type "${trimmed}"`, 'success');
		return true;
	}

	renameObjectType(oldID: string, newID: string): boolean {
		const trimmed = newID.trim();
		if (trimmed === oldID) return true;
		if (trimmed === '') {
			this.notify('Object type ID is required', 'error');
			return false;
		}
		if (this.objectTypes.some((type) => type.id === trimmed)) {
			this.notify(`Object type "${trimmed}" already exists`, 'error');
			return false;
		}
		if (!this.objectTypes.some((type) => type.id === oldID)) return false;

		this.objectTypes = this.objectTypes.map((type) =>
			type.id === oldID ? { ...type, id: trimmed } : type
		);
		renameObjectsOfType(this.map, oldID, trimmed);
		if (this.activeObjectTypeId === oldID) this.activeObjectTypeId = trimmed;
		this.commit();
		return true;
	}

	setObjectTypeColor(id: string, color: string): void {
		if (!isValidHexColor(color)) return;
		this.objectTypes = this.objectTypes.map((type) => (type.id === id ? { ...type, color } : type));
	}

	deleteObjectType(id: string): boolean {
		const count = this.objectCountForType(id);
		if (count > 0) {
			this.notify(
				`Cannot delete "${id}": ${count} placed object${count === 1 ? '' : 's'} use it`,
				'error'
			);
			return false;
		}
		if (!this.objectTypes.some((type) => type.id === id)) return false;
		this.objectTypes = this.objectTypes.filter((type) => type.id !== id);
		if (this.activeObjectTypeId === id) this.activeObjectTypeId = null;
		if (this.activeObjectTypeId === null && this.tool === 'object') this.tool = 'paint';
		this.notify(`Deleted object type "${id}"`, 'success');
		return true;
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
			marginX: this.slicing.marginX,
			marginY: this.slicing.marginY
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
			if (erase && removeObjectsAt(this.map, cell.x, cell.y) > 0) {
				changed = true;
			}
		}
		if (changed) this.commit();
	}

	/** Place the active object type along a drag path. Objects never touch layer grids. */
	placeObjectLine(from: { x: number; y: number }, to: { x: number; y: number }): void {
		const type = this.activeObjectType;
		const layer = this.activeLayer;
		if (!type || !layer) return;

		let changed = false;
		for (const cell of lineBetween(from, to)) {
			if (!inBounds(this.map, cell.x, cell.y)) continue;
			if (addObject(this.map, type.id, layer.id, cell.x, cell.y)) changed = true;
		}
		if (changed) this.commit();
	}

	/** Remove objects along a drag path without touching tile cells. */
	eraseObjectsLine(from: { x: number; y: number }, to: { x: number; y: number }): void {
		let changed = false;
		for (const cell of lineBetween(from, to)) {
			if (!inBounds(this.map, cell.x, cell.y)) continue;
			if (removeObjectsAt(this.map, cell.x, cell.y) > 0) changed = true;
		}
		if (changed) this.commit();
	}

	canPaint(): boolean {
		if (this.tool === 'object') {
			if (!this.activeObjectType) {
				this.notify('Select an object type before placing objects', 'info');
				return false;
			}
			return true;
		}
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
		downloadJson(safeName, mapToJsonWithSchema(this.map, this.mapSchema));
		this.notify(`Exported ${safeName}`, 'success');
		return true;
	}
}
