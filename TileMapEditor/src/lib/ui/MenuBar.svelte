<script lang="ts">
	import { onMount } from 'svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import type { EditorTool } from '#lib/renderer/draw-map';
	import SettingsModal from './SettingsModal.svelte';

	let { editor }: { editor: EditorState } = $props();

	type MenuName = 'file' | 'edit' | 'view' | 'tools' | 'help';

	let open = $state<MenuName | null>(null);
	let settingsOpen = $state(false);
	let mapInput: HTMLInputElement | undefined = $state();
	let root: HTMLDivElement | undefined = $state();

	function toggle(name: MenuName) {
		open = open === name ? null : name;
	}

	function hover(name: MenuName) {
		if (open) open = name;
	}

	function close() {
		open = null;
	}

	function run(action: () => void) {
		close();
		action();
	}

	function newMap() {
		if (!window.confirm('Create a new blank map? Unsaved changes will be lost.')) return;
		editor.newMap();
	}

	async function importMap(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		if (!window.confirm(`Import "${file.name}"? The current map will be replaced.`)) return;
		await editor.importMapFile(file);
	}

	function exportMap() {
		editor.exportMap(editor.suggestedFilename);
	}

	function pickTool(tool: EditorTool) {
		editor.tool = tool;
	}

	onMount(() => {
		function pointerDown(event: PointerEvent) {
			if (!root) return;
			const target = event.target;
			if (target instanceof Node && root.contains(target)) return;
			close();
		}
		function editableTarget(target: EventTarget | null): boolean {
			if (!(target instanceof HTMLElement)) return false;
			const tag = target.tagName;
			return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
		}

		function keyDown(event: KeyboardEvent) {
			if (event.key === 'Escape') {
				close();
				return;
			}
			if (!(event.ctrlKey || event.metaKey) || editableTarget(event.target)) return;
			if (event.code === 'KeyN' && !event.shiftKey) {
				event.preventDefault();
				newMap();
			} else if (event.code === 'KeyO' && !event.shiftKey) {
				event.preventDefault();
				mapInput?.click();
			} else if (event.code === 'KeyS' && !event.shiftKey) {
				event.preventDefault();
				editor.saveNow();
			} else if (event.code === 'KeyE' && !event.shiftKey) {
				event.preventDefault();
				if (editor.canExport) exportMap();
			}
		}
		window.addEventListener('pointerdown', pointerDown, true);
		window.addEventListener('keydown', keyDown);
		return () => {
			window.removeEventListener('pointerdown', pointerDown, true);
			window.removeEventListener('keydown', keyDown);
		};
	});
</script>

<div class="menubar" bind:this={root}>
	<span class="menubar-brand">Tile Map Editor</span>

	<input
		bind:this={mapInput}
		type="file"
		accept=".json,application/json"
		class="hidden"
		onchange={importMap}
	/>

	<div class="menu">
		<button
			class="menu-trigger"
			aria-haspopup="menu"
			aria-expanded={open === 'file'}
			onclick={() => toggle('file')}
			onmouseenter={() => hover('file')}
		>
			File
		</button>
		{#if open === 'file'}
			<div class="menu-popup" role="menu">
				<button class="menu-item" role="menuitem" onclick={() => run(newMap)}>
					<span>New map</span>
					<kbd>Ctrl N</kbd>
				</button>
				<button class="menu-item" role="menuitem" onclick={() => run(() => mapInput?.click())}>
					<span>Import map…</span>
					<kbd>Ctrl O</kbd>
				</button>
				<div class="menu-sep"></div>
				<button class="menu-item" role="menuitem" onclick={() => run(() => editor.saveNow())}>
					<span>Save map now</span>
					<kbd>Ctrl S</kbd>
				</button>
				<div class="menu-sep"></div>
				<button
					class="menu-item"
					role="menuitem"
					disabled={!editor.canExport}
					onclick={() => run(exportMap)}
				>
					<span>Export JSON</span>
					<kbd>Ctrl E</kbd>
				</button>
			</div>
		{/if}
	</div>

	<div class="menu">
		<button
			class="menu-trigger"
			aria-haspopup="menu"
			aria-expanded={open === 'edit'}
			onclick={() => toggle('edit')}
			onmouseenter={() => hover('edit')}
		>
			Edit
		</button>
		{#if open === 'edit'}
			<div class="menu-popup" role="menu">
				<button
					class="menu-item"
					role="menuitem"
					disabled={!editor.canUndo}
					onclick={() => run(() => editor.undo())}
				>
					<span>Undo</span>
					<kbd>Ctrl Z</kbd>
				</button>
				<button
					class="menu-item"
					role="menuitem"
					disabled={!editor.canRedo}
					onclick={() => run(() => editor.redo())}
				>
					<span>Redo</span>
					<kbd>Ctrl Y</kbd>
				</button>
				<div class="menu-sep"></div>
				<button class="menu-item" role="menuitem" onclick={() => run(() => (settingsOpen = true))}>
					<span>Settings…</span>
				</button>
			</div>
		{/if}
	</div>

	<div class="menu">
		<button
			class="menu-trigger"
			aria-haspopup="menu"
			aria-expanded={open === 'view'}
			onclick={() => toggle('view')}
			onmouseenter={() => hover('view')}
		>
			View
		</button>
		{#if open === 'view'}
			<div class="menu-popup" role="menu">
				<button
					class="menu-item"
					role="menuitemcheckbox"
					aria-checked={editor.showGrid}
					onclick={() => run(() => (editor.showGrid = !editor.showGrid))}
				>
					<span>Show grid</span>
					<span class="menu-check">{editor.showGrid ? '✓' : ''}</span>
				</button>
				<button
					class="menu-item"
					role="menuitemcheckbox"
					aria-checked={editor.showTilesetArt}
					onclick={() => run(() => (editor.showTilesetArt = !editor.showTilesetArt))}
				>
					<span>Show tileset artwork</span>
					<span class="menu-check">{editor.showTilesetArt ? '✓' : ''}</span>
				</button>
				<div class="menu-sep"></div>
				<button class="menu-item" role="menuitem" onclick={() => run(() => editor.fitView())}>
					<span>Fit map in window</span>
					<kbd>Shift Ctrl J</kbd>
				</button>
				<button class="menu-item" role="menuitem" onclick={() => run(() => editor.zoomBy(1.25))}>
					<span>Zoom in</span>
					<kbd>+</kbd>
				</button>
				<button
					class="menu-item"
					role="menuitem"
					onclick={() => run(() => editor.zoomBy(1 / 1.25))}
				>
					<span>Zoom out</span>
					<kbd>−</kbd>
				</button>
				<button
					class="menu-item"
					role="menuitem"
					onclick={() => run(() => editor.setZoomPercent(100))}
				>
					<span>Actual size</span>
					<kbd>1</kbd>
				</button>
			</div>
		{/if}
	</div>

	<div class="menu">
		<button
			class="menu-trigger"
			aria-haspopup="menu"
			aria-expanded={open === 'tools'}
			onclick={() => toggle('tools')}
			onmouseenter={() => hover('tools')}
		>
			Tools
		</button>
		{#if open === 'tools'}
			<div class="menu-popup" role="menu">
				<button class="menu-item" role="menuitem" onclick={() => run(() => pickTool('paint'))}>
					<span>Paint</span>
					<kbd>B</kbd>
				</button>
				<button class="menu-item" role="menuitem" onclick={() => run(() => pickTool('erase'))}>
					<span>Erase</span>
					<kbd>E</kbd>
				</button>
				<button class="menu-item" role="menuitem" onclick={() => run(() => pickTool('object'))}>
					<span>Place object</span>
					<kbd>O</kbd>
				</button>
				<button class="menu-item" role="menuitem" onclick={() => run(() => pickTool('spawn'))}>
					<span>Set spawn point</span>
					<kbd>S</kbd>
				</button>
			</div>
		{/if}
	</div>

	<div class="menu">
		<button
			class="menu-trigger"
			aria-haspopup="menu"
			aria-expanded={open === 'help'}
			onclick={() => toggle('help')}
			onmouseenter={() => hover('help')}
		>
			Help
		</button>
		{#if open === 'help'}
			<div class="menu-popup" role="menu">
				<button
					class="menu-item"
					role="menuitem"
					onclick={() =>
						run(() =>
							editor.notify(
								'Scroll zooms · Space or middle-drag pans · right-click erases · double-click fits',
								'info'
							)
						)}
				>
					<span>Canvas controls</span>
				</button>
				<button
					class="menu-item"
					role="menuitem"
					onclick={() =>
						run(() =>
							editor.notify(
								'B paint · E erase · O object · S spawn · G grid · +/- zoom · 1 actual size · Ctrl+Z/Y undo/redo',
								'info'
							)
						)}
				>
					<span>Keyboard shortcuts</span>
				</button>
				<div class="menu-sep"></div>
				<button
					class="menu-item"
					role="menuitem"
					onclick={() => run(() => editor.notify('Tilemap editor · Thirty Nights', 'info'))}
				>
					<span>About this editor</span>
				</button>
			</div>
		{/if}
	</div>
</div>

{#if settingsOpen}
	<SettingsModal {editor} onclose={() => (settingsOpen = false)} />
{/if}
