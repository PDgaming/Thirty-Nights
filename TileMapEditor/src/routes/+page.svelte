<script lang="ts">
	import { onMount } from 'svelte';
	import { EditorState, type EditorTab, type LeftPanelTab } from '#lib/editor-state.svelte';
	import MapCanvas from '#lib/ui/MapCanvas.svelte';
	import MapSettings from '#lib/ui/MapSettings.svelte';
	import ObjectSettings from '#lib/ui/ObjectSettings.svelte';
	import ValidationPanel from '#lib/ui/ValidationPanel.svelte';
	import Toolbar from '#lib/ui/Toolbar.svelte';
	import Toolbox from '#lib/ui/Toolbox.svelte';
	import MenuBar from '#lib/ui/MenuBar.svelte';
	import StatusBar from '#lib/ui/StatusBar.svelte';
	import Toasts from '#lib/ui/Toasts.svelte';
	import TilesetBrowser from '#lib/ui/TilesetBrowser.svelte';
	import TileInspector from '#lib/ui/TileInspector.svelte';
	import ObjectPanel from '#lib/ui/ObjectPanel.svelte';
	import LayerPanel from '#lib/ui/LayerPanel.svelte';

	const editor = new EditorState();

	const tabs: { id: EditorTab; label: string }[] = [
		{ id: 'tileset', label: 'Tileset' },
		{ id: 'inspector', label: 'Tile' },
		{ id: 'objects', label: 'Objects' }
	];

	const leftTabs: { id: LeftPanelTab; label: string }[] = [
		{ id: 'map', label: 'Map' },
		{ id: 'object', label: 'Object' }
	];

	let tabStrip: HTMLDivElement | undefined = $state();
	let leftTabStrip: HTMLDivElement | undefined = $state();

	// Dock widths are user-adjustable within these bounds. The workspace keeps
	// a minimum width, so the allowed maximum shrinks on narrow windows.
	const LEFT_MIN = 200;
	const LEFT_MAX = 420;
	const RIGHT_MIN = 260;
	const RIGHT_MAX = 520;
	const WORKSPACE_MIN = 360;
	const CHROME_WIDTH = 16; // two 5px splitters, four 1px gaps, 2px body padding
	const DEFAULT_LEFT = 248;
	const DEFAULT_RIGHT = 320;
	const LEFT_KEY = 'thirty-nights.tilemap-editor.left-dock-width';
	const RIGHT_KEY = 'thirty-nights.tilemap-editor.right-dock-width';
	const LAYERS_KEY = 'thirty-nights.tilemap-editor.layers-height';
	const KEY_STEP = 8;

	const LAYERS_MIN = 160;
	const LAYERS_TAB_MIN = 180;
	const LAYERS_DEFAULT_RATIO = 0.42;
	const RIGHT_DOCK_CHROME = 7; // 5px splitter plus two 1px dock gaps

	let appBody: HTMLDivElement | undefined = $state();
	let rightDock: HTMLElement | undefined = $state();
	let leftWidth = $state(DEFAULT_LEFT);
	let rightWidth = $state(DEFAULT_RIGHT);
	let resizing = $state<'left' | 'right' | null>(null);
	let dragStartX = 0;
	let dragStartWidth = 0;

	let layersHeight = $state(0);
	let layersCeiling = $state(640);
	let resizingRows = $state(false);
	let rowDragStartY = 0;
	let rowDragStartHeight = 0;

	function availableWidth(): number {
		return appBody?.clientWidth ?? 1280;
	}

	function clampLeft(value: number): number {
		const room = availableWidth() - rightWidth - CHROME_WIDTH - WORKSPACE_MIN;
		const max = Math.max(LEFT_MIN, Math.min(LEFT_MAX, room));
		return Math.round(Math.min(max, Math.max(LEFT_MIN, value)));
	}

	function clampRight(value: number): number {
		const room = availableWidth() - leftWidth - CHROME_WIDTH - WORKSPACE_MIN;
		const max = Math.max(RIGHT_MIN, Math.min(RIGHT_MAX, room));
		return Math.round(Math.min(max, Math.max(RIGHT_MIN, value)));
	}

	function clampWidths() {
		leftWidth = clampLeft(leftWidth);
		rightWidth = clampRight(rightWidth);
	}

	function readStoredNumber(key: string, fallback: number): number {
		try {
			const raw = window.localStorage.getItem(key);
			const value = raw === null ? Number.NaN : Number(raw);
			return Number.isFinite(value) ? value : fallback;
		} catch {
			return fallback;
		}
	}

	function storeNumber(key: string, value: number) {
		try {
			window.localStorage.setItem(key, String(value));
		} catch {
			// Storage can be unavailable in private modes; resizing still works.
		}
	}

	function startResize(side: 'left' | 'right', event: PointerEvent) {
		if (event.button !== 0) return;
		event.preventDefault();
		resizing = side;
		dragStartX = event.clientX;
		dragStartWidth = side === 'left' ? leftWidth : rightWidth;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
		document.body.classList.add('is-resizing');
		document.body.style.userSelect = 'none';
	}

	function moveResize(side: 'left' | 'right', event: PointerEvent) {
		if (resizing !== side) return;
		const delta = event.clientX - dragStartX;
		if (side === 'left') {
			leftWidth = clampLeft(dragStartWidth + delta);
		} else {
			rightWidth = clampRight(dragStartWidth - delta);
		}
	}

	function endResize(side: 'left' | 'right', event: PointerEvent) {
		if (resizing !== side) return;
		const handle = event.currentTarget as HTMLElement;
		if (handle.hasPointerCapture(event.pointerId)) handle.releasePointerCapture(event.pointerId);
		resizing = null;
		document.body.classList.remove('is-resizing');
		document.body.style.userSelect = '';
		storeNumber(side === 'left' ? LEFT_KEY : RIGHT_KEY, side === 'left' ? leftWidth : rightWidth);
	}

	function resetWidth(side: 'left' | 'right') {
		if (side === 'left') {
			leftWidth = clampLeft(DEFAULT_LEFT);
			storeNumber(LEFT_KEY, leftWidth);
		} else {
			rightWidth = clampRight(DEFAULT_RIGHT);
			storeNumber(RIGHT_KEY, rightWidth);
		}
	}

	function resizeWithKeyboard(side: 'left' | 'right', event: KeyboardEvent) {
		if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
		event.preventDefault();
		const step = event.key === 'ArrowLeft' ? -KEY_STEP : KEY_STEP;
		// Arrow keys move the separator, so the right dock grows when it moves left.
		if (side === 'left') {
			leftWidth = clampLeft(leftWidth + step);
			storeNumber(LEFT_KEY, leftWidth);
		} else {
			rightWidth = clampRight(rightWidth - step);
			storeNumber(RIGHT_KEY, rightWidth);
		}
	}

	function rightDockHeight(): number {
		return rightDock?.clientHeight ?? 640;
	}

	function updateLayersCeiling(): number {
		layersCeiling = Math.max(LAYERS_MIN, rightDockHeight() - RIGHT_DOCK_CHROME - LAYERS_TAB_MIN);
		return layersCeiling;
	}

	function clampLayers(value: number): number {
		const ceiling = updateLayersCeiling();
		return Math.round(Math.min(ceiling, Math.max(LAYERS_MIN, value)));
	}

	function defaultLayersHeight(): number {
		return Math.round(rightDockHeight() * LAYERS_DEFAULT_RATIO);
	}

	function startResizeRows(event: PointerEvent) {
		if (event.button !== 0) return;
		event.preventDefault();
		resizingRows = true;
		rowDragStartY = event.clientY;
		rowDragStartHeight = layersHeight > 0 ? layersHeight : defaultLayersHeight();
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
		document.body.classList.add('is-resizing-y');
		document.body.style.userSelect = 'none';
	}

	function moveResizeRows(event: PointerEvent) {
		if (!resizingRows) return;
		layersHeight = clampLayers(rowDragStartHeight - (event.clientY - rowDragStartY));
	}

	function endResizeRows(event: PointerEvent) {
		if (!resizingRows) return;
		const handle = event.currentTarget as HTMLElement;
		if (handle.hasPointerCapture(event.pointerId)) handle.releasePointerCapture(event.pointerId);
		resizingRows = false;
		document.body.classList.remove('is-resizing-y');
		document.body.style.userSelect = '';
		storeNumber(LAYERS_KEY, layersHeight);
	}

	function resetLayersHeight() {
		layersHeight = clampLayers(defaultLayersHeight());
		storeNumber(LAYERS_KEY, layersHeight);
	}

	function resizeLayersWithKeyboard(event: KeyboardEvent) {
		if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
		event.preventDefault();
		const step = event.key === 'ArrowUp' ? KEY_STEP : -KEY_STEP;
		layersHeight = clampLayers((layersHeight > 0 ? layersHeight : defaultLayersHeight()) + step);
		storeNumber(LAYERS_KEY, layersHeight);
	}

	function onTabKeydown(event: KeyboardEvent) {
		if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
		event.preventDefault();
		const index = tabs.findIndex((tab) => tab.id === editor.tab);
		const next = tabs[(index + (event.key === 'ArrowRight' ? 1 : tabs.length - 1)) % tabs.length];
		if (!next) return;
		editor.selectTab(next.id);
		const button = tabStrip?.querySelector<HTMLButtonElement>(`[data-tab="${next.id}"]`);
		button?.focus();
	}

	function onLeftTabKeydown(event: KeyboardEvent) {
		if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
		event.preventDefault();
		const index = leftTabs.findIndex((tab) => tab.id === editor.leftPanelTab);
		const next =
			leftTabs[(index + (event.key === 'ArrowRight' ? 1 : leftTabs.length - 1)) % leftTabs.length];
		if (!next) return;
		editor.leftPanelTab = next.id;
		const button = leftTabStrip?.querySelector<HTMLButtonElement>(`[data-tab="${next.id}"]`);
		button?.focus();
	}

	onMount(() => {
		leftWidth = Math.max(LEFT_MIN, Math.min(LEFT_MAX, readStoredNumber(LEFT_KEY, DEFAULT_LEFT)));
		rightWidth = Math.max(
			RIGHT_MIN,
			Math.min(RIGHT_MAX, readStoredNumber(RIGHT_KEY, DEFAULT_RIGHT))
		);
		clampWidths();
		layersHeight = clampLayers(readStoredNumber(LAYERS_KEY, defaultLayersHeight()));

		const clampOnResize = () => {
			clampWidths();
			if (layersHeight > 0) layersHeight = clampLayers(layersHeight);
		};
		window.addEventListener('resize', clampOnResize);

		const flush = () => editor.flushPendingSave();
		const flushWhenHidden = () => {
			if (document.visibilityState === 'hidden') flush();
		};

		window.addEventListener('beforeunload', flush);
		window.addEventListener('pagehide', flush);
		document.addEventListener('visibilitychange', flushWhenHidden);

		return () => {
			document.body.classList.remove('is-resizing');
			document.body.classList.remove('is-resizing-y');
			document.body.style.userSelect = '';
			window.removeEventListener('resize', clampOnResize);
			window.removeEventListener('beforeunload', flush);
			window.removeEventListener('pagehide', flush);
			document.removeEventListener('visibilitychange', flushWhenHidden);
			flush();
		};
	});
</script>

<svelte:head>
	<title>Tilemap Editor · Thirty Nights</title>
	<meta
		name="description"
		content="Browser-based tilemap editor for Thirty Nights. Import a tileset, paint layers, validate and export game-ready JSON."
	/>
</svelte:head>

<div class="app">
	<MenuBar {editor} />

	<div class="app-body" bind:this={appBody}>
		<aside class="dock-column" style="width: {leftWidth}px; flex-shrink: 0">
			<section class="dock">
				<div class="panel-title">
					<span>Toolbox</span>
				</div>
				<Toolbox {editor} />
			</section>

			<Toolbar {editor} />

			<section class="dock dock-grow">
				<div
					class="tabs"
					role="tablist"
					aria-label="Map and object panels"
					bind:this={leftTabStrip}
				>
					{#each leftTabs as tab (tab.id)}
						<button
							class="tab"
							class:active={editor.leftPanelTab === tab.id}
							role="tab"
							id="left-tab-{tab.id}"
							data-tab={tab.id}
							aria-selected={editor.leftPanelTab === tab.id}
							aria-controls="left-panel-{tab.id}"
							tabindex={editor.leftPanelTab === tab.id ? 0 : -1}
							onclick={() => (editor.leftPanelTab = tab.id)}
							onkeydown={onLeftTabKeydown}
						>
							{tab.label}
						</button>
					{/each}
				</div>
				<div
					class="dock-scroll"
					role="tabpanel"
					tabindex="0"
					id="left-panel-{editor.leftPanelTab}"
					aria-labelledby="left-tab-{editor.leftPanelTab}"
				>
					{#if editor.leftPanelTab === 'map'}
						<MapSettings {editor} />
						<ValidationPanel {editor} />
					{:else if editor.editingObjectId}
						{#key `${editor.editingObjectId}:${editor.mapEpoch}`}
							<ObjectSettings {editor} objectId={editor.editingObjectId} />
						{/key}
					{:else}
						<div class="panel-body">
							<p class="text-[11px] leading-4" style="color: var(--text-faint)">
								Pick the Select tool (<span class="mono">V</span>) and click a cell with an object
								to edit it.
							</p>
						</div>
					{/if}
				</div>
			</section>
		</aside>

		<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="splitter"
			class:active={resizing === 'left'}
			role="separator"
			aria-orientation="vertical"
			aria-label="Resize left sidebar"
			aria-valuemin={LEFT_MIN}
			aria-valuemax={LEFT_MAX}
			aria-valuenow={leftWidth}
			tabindex="0"
			title="Drag to resize · {leftWidth}px · double-click to reset"
			onpointerdown={(event) => startResize('left', event)}
			onpointermove={(event) => moveResize('left', event)}
			onpointerup={(event) => endResize('left', event)}
			onpointercancel={(event) => endResize('left', event)}
			onkeydown={(event) => resizeWithKeyboard('left', event)}
			ondblclick={() => resetWidth('left')}
		></div>

		<main class="workspace">
			<div class="relative min-h-0 flex-1">
				<MapCanvas {editor} />
				<Toasts {editor} />
			</div>
		</main>

		<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="splitter"
			class:active={resizing === 'right'}
			role="separator"
			aria-orientation="vertical"
			aria-label="Resize right sidebar"
			aria-valuemin={RIGHT_MIN}
			aria-valuemax={RIGHT_MAX}
			aria-valuenow={rightWidth}
			tabindex="0"
			title="Drag to resize · {rightWidth}px · double-click to reset"
			onpointerdown={(event) => startResize('right', event)}
			onpointermove={(event) => moveResize('right', event)}
			onpointerup={(event) => endResize('right', event)}
			onpointercancel={(event) => endResize('right', event)}
			onkeydown={(event) => resizeWithKeyboard('right', event)}
			ondblclick={() => resetWidth('right')}
		></div>

		<aside class="dock-column" bind:this={rightDock} style="width: {rightWidth}px; flex-shrink: 0">
			<section class="dock dock-grow">
				<div class="tabs" role="tablist" aria-label="Tile and object panels" bind:this={tabStrip}>
					{#each tabs as tab (tab.id)}
						<button
							class="tab"
							class:active={editor.tab === tab.id}
							role="tab"
							id="tab-{tab.id}"
							data-tab={tab.id}
							aria-selected={editor.tab === tab.id}
							aria-controls="panel-{tab.id}"
							tabindex={editor.tab === tab.id ? 0 : -1}
							onclick={() => editor.selectTab(tab.id)}
							onkeydown={onTabKeydown}
						>
							{tab.label}
						</button>
					{/each}
				</div>
				<div
					class="dock-scroll"
					role="tabpanel"
					tabindex="0"
					id="panel-{editor.tab}"
					aria-labelledby="tab-{editor.tab}"
				>
					{#if editor.tab === 'tileset'}
						<TilesetBrowser {editor} />
					{:else if editor.tab === 'objects'}
						<ObjectPanel {editor} />
					{:else}
						<TileInspector {editor} />
					{/if}
				</div>
			</section>

			<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div
				class="splitter-h"
				class:active={resizingRows}
				role="separator"
				aria-orientation="horizontal"
				aria-label="Resize layers panel"
				aria-valuemin={LAYERS_MIN}
				aria-valuemax={layersCeiling}
				aria-valuenow={layersHeight}
				tabindex="0"
				title="Drag to resize · {layersHeight}px · double-click to reset"
				onpointerdown={startResizeRows}
				onpointermove={moveResizeRows}
				onpointerup={endResizeRows}
				onpointercancel={endResizeRows}
				onkeydown={resizeLayersWithKeyboard}
				ondblclick={resetLayersHeight}
			></div>

			<section class="dock" style="flex: 0 0 {layersHeight > 0 ? `${layersHeight}px` : '42%'}">
				<div class="dock-scroll">
					<LayerPanel {editor} />
				</div>
			</section>
		</aside>
	</div>

	<StatusBar {editor} />
</div>

<style>
	.doc-actions .btn {
		font-weight: 500;
	}
</style>
