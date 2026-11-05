<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';

	let {
		editor,
		tileId,
		size = 96
	}: { editor: EditorState; tileId: string | null; size?: number } = $props();

	let canvas: HTMLCanvasElement | undefined = $state();
	let fileInput: HTMLInputElement | undefined = $state();

	const tile = $derived(tileId ? editor.map.tiles[tileId] : undefined);
	const slice = $derived(editor.sliceForTile(tileId));

	// Track if the tile has a texture image
	const hasTexture = $derived(!!tile?.visual.texture);

	$effect(() => {
		const element = canvas;
		if (!element) return;
		// Establish reactive dependencies.
		const currentTile = tile;
		const currentSlice = slice;
		const tilesetImage = editor.tileset.image;

		const dpr = window.devicePixelRatio || 1;
		element.width = Math.round(size * dpr);
		element.height = Math.round(size * dpr);
		const ctx = element.getContext('2d');
		if (!ctx) return;

		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		ctx.clearRect(0, 0, size, size);
		ctx.fillStyle = '#222222';
		ctx.fillRect(0, 0, size, size);

		if (!currentTile) {
			ctx.strokeStyle = 'rgba(255,255,255,0.08)';
			ctx.strokeRect(0.5, 0.5, size - 1, size - 1);
			return;
		}

		// If the tile has a texture image, draw it
		if (currentTile.visual.texture) {
			const img = new Image();
			img.onload = () => {
				ctx.imageSmoothingEnabled = false;
				ctx.drawImage(img, 0, 0, size, size);
			};
			img.src = currentTile.visual.texture!;
			if (img.complete) {
				ctx.imageSmoothingEnabled = false;
				ctx.drawImage(img, 0, 0, size, size);
			}
		} else {
			ctx.fillStyle = currentTile.visual.color;
			ctx.fillRect(0, 0, size, size);
		}

		if (tilesetImage && currentSlice && !currentTile.visual.texture) {
			ctx.imageSmoothingEnabled = false;
			ctx.drawImage(
				tilesetImage,
				currentSlice.sx,
				currentSlice.sy,
				currentSlice.sw,
				currentSlice.sh,
				0,
				0,
				size,
				size
			);
		}

		ctx.strokeStyle = 'rgba(255,255,255,0.14)';
		ctx.strokeRect(0.5, 0.5, size - 1, size - 1);
	});

	// Handle file selection for tile texture
	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (!input.files || input.files.length === 0) return;

		const file = input.files[0];
		if (!file.type.startsWith('image/')) {
			editor.notify('Please select an image file', 'error');
			return;
		}

		const reader = new FileReader();
		reader.onload = (e) => {
			if (!e.target?.result) return;
			editor.setTileTexture(tileId as string, e.target.result as string);
			// Reset the input so the same file can be selected again
			input.value = '';
		};
		reader.readAsDataURL(file);
	}

	// Allow drag-and-drop on the canvas preview
	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		const dt = event.dataTransfer;
		if (dt) dt.dropEffect = 'copy';
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		const file = event.dataTransfer?.files?.[0];
		if (!file || !file.type.startsWith('image/')) return;

		const reader = new FileReader();
		reader.onload = (e) => {
			if (!e.target?.result) return;
			editor.setTileTexture(tileId as string, e.target.result as string);
		};
		reader.readAsDataURL(file);
	}
</script>

<canvas
	bind:this={canvas}
	style="width: {size}px; height: {size}px; image-rendering: pixelated"
	class="rounded-sm"
	ondragover={handleDragOver}
	ondrop={handleDrop}
></canvas>

<!-- File input for selecting tile texture (hidden, triggered by drag-and-drop or button) -->
<input
	type="file"
	accept="image/*"
	bind:this={fileInput}
	onchange={handleFileSelect}
	style="position: absolute; left: -9999px; width: 1px; height: 1px"
/>

<!-- Button to select texture (shown when tile has no texture yet) -->
{#if !hasTexture}
	<div class="absolute inset-0 flex items-center justify-center pointer-events-none">
		<div class="rounded-sm p-2 text-sm text-[var(--text-faint)] opacity-50">
			Click or drag an image here to use as tile texture
		</div>
	</div>
{/if}
