<script lang="ts">
	import { onMount } from 'svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import { findSliceAtPoint, type TilesetSlice } from '#lib/tileset/slicing';

	let { editor }: { editor: EditorState } = $props();

	let wrapper: HTMLDivElement | undefined = $state();
	let canvas: HTMLCanvasElement | undefined = $state();
	let width = $state(0);
	let hovered = $state<TilesetSlice | null>(null);

	const image = $derived(editor.tileset.image);
	const slices = $derived(editor.tilesetSlices);
	const selectedIndex = $derived(editor.sliceForActiveTile?.index ?? -1);
	const displayHeight = $derived(
		image && width > 0
			? Math.max(1, Math.round((width * image.naturalHeight) / image.naturalWidth))
			: 0
	);

	onMount(() => {
		const element = wrapper;
		if (!element) return;
		const observer = new ResizeObserver((entries) => {
			const entry = entries[0];
			if (entry) width = Math.floor(entry.contentRect.width);
		});
		observer.observe(element);
		width = Math.floor(element.getBoundingClientRect().width);

		return () => observer.disconnect();
	});

	$effect(() => {
		const element = canvas;
		const currentImage = image;
		const currentSlices = slices;
		const currentSelected = selectedIndex;
		const currentHovered = hovered;
		const drawWidth = width;

		if (!element || !currentImage || drawWidth <= 0) return;

		const drawHeight = Math.max(
			1,
			Math.round((drawWidth * currentImage.naturalHeight) / currentImage.naturalWidth)
		);
		const dpr = window.devicePixelRatio || 1;
		element.width = Math.round(drawWidth * dpr);
		element.height = Math.round(drawHeight * dpr);
		element.style.height = `${drawHeight}px`;

		const ctx = element.getContext('2d');
		if (!ctx) return;

		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		ctx.clearRect(0, 0, drawWidth, drawHeight);
		ctx.imageSmoothingEnabled = false;
		ctx.drawImage(currentImage, 0, 0, drawWidth, drawHeight);

		const scaleX = drawWidth / currentImage.naturalWidth;
		const scaleY = drawHeight / currentImage.naturalHeight;

		ctx.lineWidth = 1;
		for (const slice of currentSlices) {
			const x = slice.sx * scaleX;
			const y = slice.sy * scaleY;
			const w = slice.sw * scaleX;
			const h = slice.sh * scaleY;

			if (slice.index === currentSelected) {
				ctx.fillStyle = 'rgba(94, 234, 212, 0.18)';
				ctx.fillRect(x, y, w, h);
				ctx.strokeStyle = '#5eead4';
				ctx.lineWidth = 2;
				ctx.strokeRect(x + 1, y + 1, w - 2, h - 2);
				ctx.lineWidth = 1;
			} else if (currentHovered && slice.index === currentHovered.index) {
				ctx.fillStyle = 'rgba(255, 255, 255, 0.1)';
				ctx.fillRect(x, y, w, h);
				ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
				ctx.strokeRect(x + 0.5, y + 0.5, w - 1, h - 1);
			} else {
				ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
				ctx.strokeRect(x + 0.5, y + 0.5, w - 1, h - 1);
			}
		}
	});

	function sourcePoint(event: MouseEvent): { x: number; y: number } | null {
		const element = canvas;
		const currentImage = image;
		if (!element || !currentImage) return null;
		const rect = element.getBoundingClientRect();
		return {
			x: ((event.clientX - rect.left) / rect.width) * currentImage.naturalWidth,
			y: ((event.clientY - rect.top) / rect.height) * currentImage.naturalHeight
		};
	}

	function handleMove(event: MouseEvent) {
		const point = sourcePoint(event);
		hovered = point ? findSliceAtPoint(slices, point.x, point.y) : null;
	}

	function handleClick(event: MouseEvent) {
		const point = sourcePoint(event);
		if (!point) return;
		const slice = findSliceAtPoint(slices, point.x, point.y);
		if (slice) editor.selectSlice(slice.index);
	}

	function handleKey(event: KeyboardEvent) {
		if (slices.length === 0) return;
		const current = selectedIndex;
		let next: number;
		if (
			event.key === 'ArrowRight' ||
			event.key === 'ArrowDown' ||
			event.key === 'Enter' ||
			event.key === ' '
		) {
			next = current < 0 ? 0 : Math.min(slices.length - 1, current + 1);
		} else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
			next = Math.max(0, (current < 0 ? slices.length : current) - 1);
		} else {
			return;
		}
		event.preventDefault();
		editor.selectSlice(next);
	}
</script>

<div bind:this={wrapper} class="w-full">
	{#if image && width > 0}
		<canvas
			bind:this={canvas}
			class="checkerboard block w-full cursor-crosshair rounded-sm"
			style="image-rendering: pixelated; height: {displayHeight}px"
			onmousemove={handleMove}
			onmouseleave={() => (hovered = null)}
			onclick={handleClick}
			onkeydown={handleKey}
			tabindex="0"
			role="button"
			aria-label="Tileset image; click a tile or use arrow keys to select one"
		></canvas>
	{:else}
		<div
			class="flex aspect-[4/3] w-full flex-col items-center justify-center gap-2 rounded-sm border border-dashed text-center"
			style="border-color: var(--border-strong); color: var(--text-faint)"
		>
			<span class="text-[11px]">No tileset imported</span>
			<span class="px-6 text-[10px] leading-4">Import a PNG or JPEG, then slice it below.</span>
		</div>
	{/if}
</div>
