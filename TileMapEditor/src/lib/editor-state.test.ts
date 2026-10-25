import { afterEach, describe, expect, it, vi } from 'vitest';
import {
	AUTOSAVE_DELAY_MS,
	EditorState,
	MAP_STORAGE_KEY,
	SELECTION_STORAGE_KEY
} from './editor-state.svelte';
import { DEFAULT_MAP_SCHEMA } from './map-schema';
import { mapToJson } from './map-serializer';

function memoryStorage(): Storage {
	const data = new Map<string, string>();
	return {
		get length() {
			return data.size;
		},
		clear: () => data.clear(),
		getItem: (key) => data.get(key) ?? null,
		key: (index) => [...data.keys()][index] ?? null,
		removeItem: (key) => {
			data.delete(key);
		},
		setItem: (key, value) => {
			data.set(key, value);
		}
	};
}

describe('undo history', () => {
	it('starts with nothing to undo', () => {
		const editor = new EditorState();
		expect(editor.canUndo).toBe(false);
		expect(editor.canRedo).toBe(false);
		expect(editor.undo()).toBe(false);
		expect(editor.redo()).toBe(false);
	});

	it('restores map fields from before the change', () => {
		const editor = new EditorState();

		editor.setMapName('Changed');
		expect(editor.map.name).toBe('Changed');
		expect(editor.canUndo).toBe(true);

		expect(editor.undo()).toBe(true);
		expect(editor.map.name).toBe('Untitled Map');
		expect(editor.canUndo).toBe(false);
	});

	it('undoes a group of commits as one step', () => {
		const editor = new EditorState();
		editor.selectTile('wall');

		editor.beginUndoGroup();
		editor.paintLine({ x: 0, y: 0 }, { x: 2, y: 0 }, false);
		editor.paintLine({ x: 2, y: 0 }, { x: 4, y: 0 }, false);
		editor.endUndoGroup();

		const layer = editor.map.layers[0];
		expect(layer.grid[0][4]).toBe('wall');

		editor.undo();
		expect(editor.map.layers[0].grid[0][4]).toBe('');
		expect(editor.canUndo).toBe(false);
	});

	it('redoes an undone change', () => {
		const editor = new EditorState();
		editor.setMapName('Changed');

		editor.undo();
		expect(editor.map.name).toBe('Untitled Map');
		expect(editor.canRedo).toBe(true);

		expect(editor.redo()).toBe(true);
		expect(editor.map.name).toBe('Changed');
		expect(editor.canRedo).toBe(false);
		expect(editor.canUndo).toBe(true);
	});

	it('redoes a grouped stroke as one step', () => {
		const editor = new EditorState();
		editor.selectTile('wall');

		editor.beginUndoGroup();
		editor.paintLine({ x: 0, y: 0 }, { x: 2, y: 0 }, false);
		editor.paintLine({ x: 2, y: 0 }, { x: 4, y: 0 }, false);
		editor.endUndoGroup();

		editor.undo();
		expect(editor.map.layers[0].grid[0][4]).toBe('');

		editor.redo();
		expect(editor.map.layers[0].grid[0][4]).toBe('wall');
		expect(editor.canRedo).toBe(false);
	});

	it('drops redo history when a new edit is made', () => {
		const editor = new EditorState();
		editor.setMapName('One');
		editor.undo();
		expect(editor.canRedo).toBe(true);

		editor.setMapName('Two');

		expect(editor.canRedo).toBe(false);
		expect(editor.map.name).toBe('Two');
	});

	it('clears history when a new map replaces the current one', () => {
		const editor = new EditorState();
		editor.setMapName('Changed');

		editor.newMap({ name: 'Fresh' });

		expect(editor.map.name).toBe('Fresh');
		expect(editor.canUndo).toBe(false);
		expect(editor.canRedo).toBe(false);
	});
});

describe('local autosave', () => {
	afterEach(() => {
		vi.useRealTimers();
		vi.unstubAllGlobals();
	});

	it('batches edits into one delayed write', () => {
		vi.useFakeTimers();
		const storage = memoryStorage();
		vi.stubGlobal('window', { localStorage: storage });
		const editor = new EditorState();

		editor.setMapName('One');
		editor.setMapName('Two');
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();

		vi.advanceTimersByTime(AUTOSAVE_DELAY_MS - 1);
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();

		vi.advanceTimersByTime(1);
		expect(storage.getItem(MAP_STORAGE_KEY)).toContain('Two');
	});

	it('uses and remembers a configured autosave delay', () => {
		vi.useFakeTimers();
		const storage = memoryStorage();
		vi.stubGlobal('window', { localStorage: storage });
		const editor = new EditorState();

		editor.setAutosaveDelay(2000);
		editor.setMapName('Fast');
		vi.advanceTimersByTime(1999);
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();

		vi.advanceTimersByTime(1);
		expect(storage.getItem(MAP_STORAGE_KEY)).toContain('Fast');
		expect(new EditorState().autosaveDelayMs).toBe(2000);
	});

	it('saves immediately and cancels the pending write', () => {
		vi.useFakeTimers();
		const storage = memoryStorage();
		vi.stubGlobal('window', { localStorage: storage });
		const editor = new EditorState();

		editor.setMapName('Manual save');
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();

		editor.saveNow();
		expect(storage.getItem(MAP_STORAGE_KEY)).toContain('Manual save');

		// The timer from the edit must not fire a second write.
		storage.removeItem(MAP_STORAGE_KEY);
		vi.advanceTimersByTime(AUTOSAVE_DELAY_MS);
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();
	});

	it('flushes pending edits on demand', () => {
		vi.useFakeTimers();
		const storage = memoryStorage();
		vi.stubGlobal('window', { localStorage: storage });
		const editor = new EditorState();

		editor.setMapName('Unloading');
		editor.flushPendingSave();

		expect(storage.getItem(MAP_STORAGE_KEY)).toContain('Unloading');
	});

	it('restores the map after a reload', () => {
		vi.useFakeTimers();
		const storage = memoryStorage();
		vi.stubGlobal('window', { localStorage: storage });
		const editor = new EditorState();
		editor.setMapName('Restored map');
		editor.selectTile('wall');
		editor.paintLine({ x: 1, y: 1 }, { x: 1, y: 1 }, false);
		vi.advanceTimersByTime(AUTOSAVE_DELAY_MS);

		const reloaded = new EditorState();
		expect(reloaded.map.name).toBe('Restored map');
		expect(reloaded.map.layers[0].grid[1][1]).toBe('wall');
		expect(reloaded.canUndo).toBe(false);
	});

	it('moves a map saved in session storage by an older build', () => {
		const storage = memoryStorage();
		const legacy = memoryStorage();
		const source = new EditorState();
		source.setMapName('Legacy map');
		legacy.setItem(MAP_STORAGE_KEY, mapToJson(source.map));
		vi.stubGlobal('window', { localStorage: storage, sessionStorage: legacy });

		const reloaded = new EditorState();

		expect(reloaded.map.name).toBe('Legacy map');
		expect(storage.getItem(MAP_STORAGE_KEY)).toContain('Legacy map');
		expect(legacy.getItem(MAP_STORAGE_KEY)).toBeNull();
	});

	it('clears corrupt saved data instead of failing to start', () => {
		const storage = memoryStorage();
		storage.setItem(MAP_STORAGE_KEY, '{"not":"a map"}');
		vi.stubGlobal('window', { localStorage: storage });

		const editor = new EditorState();

		expect(editor.map.name).toBe('Untitled Map');
		expect(storage.getItem(MAP_STORAGE_KEY)).toBeNull();
	});
});

describe('selection persistence', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	function stubBrowser() {
		const localStorage = memoryStorage();
		vi.stubGlobal('window', { localStorage, sessionStorage: memoryStorage() });
		return localStorage;
	}

	it('keeps at most one of a tile or an object type selected', () => {
		const editor = new EditorState();

		editor.selectTile('wall');
		expect(editor.activeTileId).toBe('wall');

		editor.selectObjectType('door');
		expect(editor.activeObjectTypeId).toBe('door');
		expect(editor.activeTileId).toBeNull();

		editor.selectTile('wall');
		expect(editor.activeTileId).toBe('wall');
		expect(editor.activeObjectTypeId).toBeNull();
	});

	it('restores the selected tile after a reload', () => {
		stubBrowser();
		const editor = new EditorState();
		editor.selectTile('wall');

		const reloaded = new EditorState();

		expect(reloaded.activeTileId).toBe('wall');
		expect(reloaded.activeObjectTypeId).toBeNull();
		expect(reloaded.tool).toBe('paint');
	});

	it('restores the selected object type after a reload', () => {
		stubBrowser();
		const editor = new EditorState();
		editor.selectObjectType('door');

		const reloaded = new EditorState();

		expect(reloaded.activeObjectTypeId).toBe('door');
		expect(reloaded.activeTileId).toBeNull();
		expect(reloaded.tool).toBe('object');
	});

	it('drops a saved selection whose target no longer exists', () => {
		const storage = stubBrowser();
		storage.setItem(SELECTION_STORAGE_KEY, JSON.stringify({ kind: 'tile', id: 'ghost' }));

		const editor = new EditorState();

		expect(editor.activeTileId).toBeNull();
		expect(storage.getItem(SELECTION_STORAGE_KEY)).toBeNull();
	});

	it('ignores malformed saved selection data', () => {
		const storage = stubBrowser();
		storage.setItem(SELECTION_STORAGE_KEY, '{ nope');

		const editor = new EditorState();

		expect(editor.activeTileId).toBeNull();
		expect(editor.activeObjectTypeId).toBeNull();
	});
});

describe('object editing', () => {
	function placeDoor(editor: EditorState, x: number, y: number) {
		editor.selectObjectType('door');
		editor.placeObjectLine({ x, y }, { x, y });
	}

	it('finds the topmost object on a cell', () => {
		const editor = new EditorState();
		placeDoor(editor, 2, 3);

		expect(editor.objectAtCell(2, 3)?.id).toBe(editor.map.objects[0].id);
		expect(editor.objectAtCell(0, 0)).toBeUndefined();
	});

	it('opens the left object tab for an existing object', () => {
		const editor = new EditorState();
		placeDoor(editor, 2, 3);
		const object = editor.map.objects[0];

		editor.openObjectEditor(object.id);

		expect(editor.leftPanelTab).toBe('object');
		expect(editor.editingObjectId).toBe(object.id);
	});

	it('selects a cell and shows the object on it, or clears the object on an empty cell', () => {
		const editor = new EditorState();
		placeDoor(editor, 2, 3);

		editor.selectCell(2, 3);
		expect(editor.selectedCell).toEqual({ x: 2, y: 3 });
		expect(editor.editingObjectId).toBe(editor.map.objects[0].id);
		expect(editor.leftPanelTab).toBe('object');

		editor.selectCell(0, 0);
		expect(editor.selectedCell).toEqual({ x: 0, y: 0 });
		expect(editor.editingObjectId).toBeNull();
	});

	it('moves an object and saves edited property values for export', () => {
		const editor = new EditorState();
		editor.addObjectAttribute('door', 'locked', 'boolean');
		placeDoor(editor, 1, 1);
		const object = editor.map.objects[0];

		expect(
			editor.updateObject(object.id, {
				x: 4,
				y: 5,
				properties: { ...object.properties, locked: true }
			})
		).toBe(true);
		expect(object.x).toBe(4);
		expect(object.y).toBe(5);

		const exported = JSON.parse(mapToJson(editor.map)) as {
			objects: { x: number; y: number; properties: Record<string, unknown> }[];
		};
		expect(exported.objects[0]).toMatchObject({ x: 4, y: 5, properties: { locked: true } });
	});

	it('rejects positions outside the map or already taken by the same type', () => {
		const editor = new EditorState();
		placeDoor(editor, 1, 1);
		placeDoor(editor, 2, 2);
		const [first, second] = editor.map.objects;

		expect(editor.updateObject(first.id, { x: -1, y: 0 })).toBe(false);
		expect(editor.updateObject(first.id, { x: second.x, y: second.y })).toBe(false);
		expect(first).toMatchObject({ x: 1, y: 1 });
	});
});

describe('object attributes', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	function placeDoor(editor: EditorState) {
		editor.selectObjectType('door');
		editor.placeObjectLine({ x: 1, y: 1 }, { x: 1, y: 1 });
		return editor.map.objects[0];
	}

	it('gives placed objects the declared attributes and exports them as properties', () => {
		const editor = new EditorState();
		expect(editor.addObjectAttribute('door', 'locked', 'boolean')).toBe(true);

		const object = placeDoor(editor);

		expect(object.properties).toEqual({ locked: false });
		const exported = JSON.parse(mapToJson(editor.map)) as {
			objects: { properties: Record<string, unknown> }[];
		};
		expect(exported.objects[0].properties).toEqual({ locked: false });
	});

	it('updates properties on already placed objects when attributes change', () => {
		const editor = new EditorState();
		const object = placeDoor(editor);

		editor.addObjectAttribute('door', 'locked', 'boolean');
		editor.renameObjectAttribute('door', 'locked', 'sealed');
		editor.setObjectAttributeType('door', 'sealed', 'string');
		expect(object.properties).toEqual({ sealed: '' });

		editor.removeObjectAttribute('door', 'sealed');
		expect(object.properties).toEqual({});
	});

	it('rejects duplicate attribute names', () => {
		const editor = new EditorState();
		expect(editor.addObjectAttribute('door', 'locked', 'boolean')).toBe(true);
		expect(editor.addObjectAttribute('door', 'locked', 'string')).toBe(false);
		expect(editor.objectTypes.find((type) => type.id === 'door')?.attributes).toEqual([
			{ name: 'locked', type: 'boolean' }
		]);
	});

	it('infers attribute declarations from an imported map', () => {
		vi.stubGlobal('window', { localStorage: memoryStorage() });
		const editor = new EditorState();
		editor.addObjectAttribute('door', 'locked', 'boolean');
		placeDoor(editor);
		editor.flushPendingSave();

		const reloaded = new EditorState();

		expect(reloaded.objectTypes.find((type) => type.id === 'door')?.attributes).toEqual([
			{ name: 'locked', type: 'boolean' }
		]);
	});
});

describe('map schema', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('restores a saved schema on the next visit', () => {
		vi.stubGlobal('window', { localStorage: memoryStorage() });
		const editor = new EditorState();
		const schema = {
			...DEFAULT_MAP_SCHEMA,
			properties: {
				...(DEFAULT_MAP_SCHEMA.properties as Record<string, unknown>),
				weather: { type: 'string', default: 'rain' }
			}
		};

		editor.setMapSchema(schema);

		const reloaded = new EditorState();
		const properties = reloaded.mapSchema.properties as Record<string, unknown>;
		expect(properties.weather).toEqual({ type: 'string', default: 'rain' });
	});

	it('rejects a schema that removes a compulsory field', () => {
		vi.stubGlobal('window', { localStorage: memoryStorage() });
		const editor = new EditorState();
		const properties = { ...(DEFAULT_MAP_SCHEMA.properties as Record<string, unknown>) };
		delete properties.width;

		expect(() => editor.setMapSchema({ ...DEFAULT_MAP_SCHEMA, properties })).toThrow(/width/);
	});
});
