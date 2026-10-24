<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';

	let {
		editor,
		tileId,
		size = 96
	}: { editor: EditorState; tileId: string | null; size?: number } = $props();

	let canvas: HTMLCanvasElement | undefined = $state();

	const tile = $derived(tileId ? editor.map.tiles[tileId] : undefined);
	const slice = $derived(editor.sliceForTile(tileId));

	$effect(() => {
		const element = canvas;
		if (!element) return;
		// Establish reactive dependencies.
		const currentTile = tile;
		const currentSlice = slice;
		const image = editor.tileset.image;

		const dpr = window.devicePixelRatio || 1;
		element.width = Math.round(size * dpr);
		element.height = Math.round(size * dpr);
		const ctx = element.getContext('2d');
		if (!ctx) return;

		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		ctx.clearRect(0, 0, size, size);
		ctx.fillStyle = '#0d1014';
		ctx.fillRect(0, 0, size, size);

		if (!currentTile) {
			ctx.strokeStyle = 'rgba(255,255,255,0.08)';
			ctx.strokeRect(0.5, 0.5, size - 1, size - 1);
			return;
		}

		ctx.fillStyle = currentTile.visual.color;
		ctx.fillRect(0, 0, size, size);

		if (image && currentSlice) {
			ctx.imageSmoothingEnabled = false;
			ctx.drawImage(
				image,
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
</script>

<canvas
	bind:this={canvas}
	style="width: {size}px; height: {size}px; image-rendering: pixelated"
	class="rounded-sm"
></canvas>
