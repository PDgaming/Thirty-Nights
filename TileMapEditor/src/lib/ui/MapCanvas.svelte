<script lang="ts">
	import { onMount } from 'svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import { drawMap } from '#lib/renderer/draw-map';
	import { screenToCell } from '#lib/renderer/viewport';
	import Ruler from './Ruler.svelte';

	let { editor }: { editor: EditorState } = $props();

	let container: HTMLDivElement | undefined = $state();
	let canvas: HTMLCanvasElement | undefined = $state();
	let width = $state(0);
	let height = $state(0);
	let spacePressed = $state(false);
	let isPanning = $state(false);
	let isPainting = $state(false);

	let frame = 0;
	let lastCell: { x: number; y: number } | null = null;
	let eraseStroke = false;
	let panX = 0;
	let panY = 0;

	function scheduleRender() {
		if (frame !== 0) return;
		frame = requestAnimationFrame(() => {
			frame = 0;
			render();
		});
	}

	function render() {
		const element = canvas;
		if (!element || width <= 0 || height <= 0) return;
		const dpr = window.devicePixelRatio || 1;
		const pixelWidth = Math.max(1, Math.round(width * dpr));
		const pixelHeight = Math.max(1, Math.round(height * dpr));
		if (element.width !== pixelWidth || element.height !== pixelHeight) {
			element.width = pixelWidth;
			element.height = pixelHeight;
		}
		const ctx = element.getContext('2d');
		if (!ctx) return;
		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

		drawMap({
			ctx,
			width,
			height,
			map: editor.map,
			view: editor.view,
			hiddenLayers: editor.hiddenLayers,
			showGrid: editor.showGrid,
			showTilesetArt: editor.showTilesetArt,
			tileset: editor.tilesetRenderData,
			hover: editor.hover,
			activeTileId: editor.activeTileId,
			activeLayerId: editor.activeLayerId,
			activeObjectTypeId: editor.activeObjectTypeId,
			objectColors: editor.objectColors,
			tool: editor.tool
		});
	}

	onMount(() => {
		const element = container;
		const elementCanvas = canvas;
		if (!element || !elementCanvas) return;

		const observer = new ResizeObserver((entries) => {
			const entry = entries[0];
			if (!entry) return;
			width = Math.max(1, Math.floor(entry.contentRect.width));
			height = Math.max(1, Math.floor(entry.contentRect.height));
			editor.setViewportSize(width, height);
			scheduleRender();
		});
		observer.observe(element);
		width = Math.max(1, Math.floor(element.getBoundingClientRect().width));
		height = Math.max(1, Math.floor(element.getBoundingClientRect().height));
		editor.setViewportSize(width, height);

		function editableTarget(target: EventTarget | null): boolean {
			if (!(target instanceof HTMLElement)) return false;
			const tag = target.tagName;
			return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable;
		}

		function keyDown(event: KeyboardEvent) {
			if (editableTarget(event.target)) return;
			if ((event.ctrlKey || event.metaKey) && !event.shiftKey && event.code === 'KeyZ') {
				event.preventDefault();
				editor.undo();
				return;
			}
			if ((event.ctrlKey || event.metaKey) && !event.shiftKey && event.code === 'KeyY') {
				event.preventDefault();
				editor.redo();
				return;
			}
			if ((event.ctrlKey || event.metaKey) && event.shiftKey && event.code === 'KeyJ') {
				event.preventDefault();
				editor.fitView();
				return;
			}
			if (!event.ctrlKey && !event.metaKey && !event.altKey) {
				if (event.code === 'Equal' || event.code === 'NumpadAdd') {
					editor.zoomBy(1.25);
					return;
				}
				if (event.code === 'Minus' || event.code === 'NumpadSubtract') {
					editor.zoomBy(1 / 1.25);
					return;
				}
				if (event.code === 'KeyB') {
					editor.tool = 'paint';
					return;
				}
				if (event.code === 'KeyE') {
					editor.tool = 'erase';
					return;
				}
				if (event.code === 'KeyO') {
					editor.tool = 'object';
					return;
				}
				if (event.code === 'KeyS') {
					editor.tool = 'spawn';
					return;
				}
				if (event.code === 'KeyG') {
					editor.showGrid = !editor.showGrid;
					return;
				}
				if (event.code === 'Digit1') {
					editor.setZoomPercent(100);
					return;
				}
			}
			if (event.code !== 'Space') return;
			event.preventDefault();
			spacePressed = true;
		}

		function keyUp(event: KeyboardEvent) {
			if (event.code !== 'Space') return;
			spacePressed = false;
		}

		function wheel(event: WheelEvent) {
			event.preventDefault();
			const rect = elementCanvas!.getBoundingClientRect();
			const x = event.clientX - rect.left;
			const y = event.clientY - rect.top;
			const factor = Math.exp(-event.deltaY * 0.0015);
			editor.zoomAt(factor, x, y);
		}

		elementCanvas.addEventListener('wheel', wheel, { passive: false });
		window.addEventListener('keydown', keyDown);
		window.addEventListener('keyup', keyUp);
		scheduleRender();

		return () => {
			observer.disconnect();
			elementCanvas.removeEventListener('wheel', wheel);
			window.removeEventListener('keydown', keyDown);
			window.removeEventListener('keyup', keyUp);
			if (frame !== 0) cancelAnimationFrame(frame);
		};
	});

	function trackRenderInputs(...values: unknown[]): number {
		return values.length;
	}

	$effect(() => {
		// Reactive render inputs. The map itself is a raw state object, so its
		// revision is tracked separately from nested grid mutations.
		trackRenderInputs(
			editor.map,
			editor.mapRevision,
			editor.view,
			editor.hiddenLayers,
			editor.showGrid,
			editor.showTilesetArt,
			editor.tileset,
			editor.tilesetSlices,
			editor.tileBindings,
			editor.hover,
			editor.activeTileId,
			editor.activeLayerId,
			editor.activeObjectTypeId,
			editor.objectTypes,
			editor.tool,
			width,
			height
		);
		scheduleRender();
	});

	function localPoint(event: PointerEvent): { x: number; y: number } {
		const rect = canvas!.getBoundingClientRect();
		return { x: event.clientX - rect.left, y: event.clientY - rect.top };
	}

	function cellFromEvent(event: PointerEvent) {
		const point = localPoint(event);
		return screenToCell(point.x, point.y, editor.view, editor.map.tile_size);
	}

	function handlePointerDown(event: PointerEvent) {
		if (!canvas) return;
		event.preventDefault();
		canvas.focus();

		if (event.button === 1 || (event.button === 0 && spacePressed)) {
			isPanning = true;
			panX = event.clientX;
			panY = event.clientY;
			canvas.setPointerCapture(event.pointerId);
			return;
		}
		if (event.button !== 0 && event.button !== 2) return;

		// Right-click uses the alternate paint/erase tool.
		const alternateTool = event.button === 2;

		const cell = cellFromEvent(event);
		if (!alternateTool && editor.tool === 'spawn') {
			editor.setSpawnFromCell(cell.x, cell.y);
			editor.tool = 'paint';
			return;
		}
		if (alternateTool && editor.tool === 'spawn') return;

		const erase = alternateTool ? editor.tool === 'paint' : editor.tool === 'erase';
		const objectErase = editor.tool === 'object' && alternateTool;

		if (!editor.canPaint()) return;
		if (!erase && editor.tool === 'paint' && (!editor.activeTileId || !editor.activeTile)) {
			editor.notify('Select a tile before painting', 'info');
			return;
		}

		isPainting = true;
		editor.beginUndoGroup();
		eraseStroke = erase || objectErase;
		lastCell = cell;
		if (editor.tool === 'object') {
			if (objectErase) editor.eraseObjectsLine(cell, cell);
			else editor.placeObjectLine(cell, cell);
		} else {
			editor.paintLine(cell, cell, eraseStroke);
		}
		canvas.setPointerCapture(event.pointerId);
	}

	function handlePointerMove(event: PointerEvent) {
		const cell = cellFromEvent(event);
		const inside =
			cell.x >= 0 && cell.y >= 0 && cell.x < editor.map.width && cell.y < editor.map.height;
		editor.hover = inside ? cell : null;

		if (isPanning) {
			editor.panBy(event.clientX - panX, event.clientY - panY);
			panX = event.clientX;
			panY = event.clientY;
			return;
		}

		if (isPainting && lastCell && (cell.x !== lastCell.x || cell.y !== lastCell.y)) {
			if (editor.tool === 'object') {
				if (eraseStroke) editor.eraseObjectsLine(lastCell, cell);
				else editor.placeObjectLine(lastCell, cell);
			} else {
				editor.paintLine(lastCell, cell, eraseStroke);
			}
			lastCell = cell;
		}
	}

	function endPointer(event: PointerEvent) {
		if (canvas?.hasPointerCapture(event.pointerId)) {
			canvas.releasePointerCapture(event.pointerId);
		}
		if (isPainting) editor.endUndoGroup();
		isPanning = false;
		isPainting = false;
		lastCell = null;
	}

	function handlePointerLeave() {
		if (!isPanning && !isPainting) editor.hover = null;
	}

	const cursor = $derived(
		isPanning
			? 'grabbing'
			: spacePressed
				? 'grab'
				: editor.tool === 'spawn'
					? 'cell'
					: editor.tool === 'object'
						? 'copy'
						: 'crosshair'
	);
</script>

<div class="relative h-full w-full overflow-hidden" style="background: var(--canvas)">
	<div bind:this={container} class="absolute" style="top: 17px; left: 17px; right: 0; bottom: 0">
		<canvas
			bind:this={canvas}
			class="absolute inset-0 block h-full w-full outline-none"
			style="cursor: {cursor}; touch-action: none"
			tabindex="0"
			aria-label="Map canvas. Scroll to zoom, space-drag to pan."
			onpointerdown={handlePointerDown}
			onpointermove={handlePointerMove}
			onpointerup={endPointer}
			onpointercancel={endPointer}
			onpointerleave={handlePointerLeave}
			oncontextmenu={(event) => event.preventDefault()}
			ondblclick={() => editor.fitView()}
		></canvas>
	</div>

	<div class="ruler-x">
		<Ruler {editor} orient="x" size={width} />
	</div>
	<div class="ruler-y">
		<Ruler {editor} orient="y" size={height} />
	</div>
	<div class="ruler-corner"></div>
</div>

<style>
	canvas:focus-visible {
		outline: 1px solid var(--focus);
		outline-offset: -1px;
	}
</style>
