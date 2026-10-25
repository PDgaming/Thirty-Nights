<script lang="ts">
	import { onMount } from 'svelte';
	import type { EditorState } from '#lib/editor-state.svelte';

	let { editor, orient, size }: { editor: EditorState; orient: 'x' | 'y'; size: number } = $props();

	const THICK = 17;

	let canvas: HTMLCanvasElement | undefined = $state();
	let fontsReady = $state(false);

	onMount(() => {
		document.fonts?.ready.then(() => {
			fontsReady = true;
		});
	});

	$effect(() => {
		const element = canvas;
		if (!element) return;

		const view = editor.view;
		const map = editor.map;
		const hover = editor.hover;

		const length = Math.max(1, Math.floor(size));
		const width = orient === 'x' ? length : THICK;
		const height = orient === 'x' ? THICK : length;
		const dpr = window.devicePixelRatio || 1;

		const pixelWidth = Math.max(1, Math.round(width * dpr));
		const pixelHeight = Math.max(1, Math.round(height * dpr));
		if (element.width !== pixelWidth || element.height !== pixelHeight) {
			element.width = pixelWidth;
			element.height = pixelHeight;
		}
		element.style.width = `${width}px`;
		element.style.height = `${height}px`;

		const ctx = element.getContext('2d');
		if (!ctx) return;

		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		ctx.clearRect(0, 0, width, height);

		ctx.fillStyle = '#262626';
		ctx.fillRect(0, 0, width, height);

		ctx.fillStyle = '#161616';
		if (orient === 'x') ctx.fillRect(0, height - 1, width, 1);
		else ctx.fillRect(width - 1, 0, 1, height);

		const tilePx = map.tile_size * view.scale;
		if (!Number.isFinite(tilePx) || tilePx <= 0) return;

		const offset = orient === 'x' ? view.offsetX : view.offsetY;
		const count = orient === 'x' ? map.width : map.height;

		const steps = [1, 2, 5, 10, 20, 50, 100, 200, 500, 1000];
		const majorStep = steps.find((step) => step * tilePx >= 46) ?? 1000;
		const minorStep = tilePx >= 7 ? 1 : majorStep;

		const start = Math.max(0, Math.floor(-offset / tilePx) - 1);
		const end = Math.min(count, Math.ceil((length - offset) / tilePx) + 1);

		// Redraw once the mono face has loaded so tick labels keep their metrics.
		ctx.font = fontsReady
			? "9px 'IBM Plex Mono', ui-monospace, monospace"
			: '9px ui-monospace, monospace';
		ctx.textBaseline = 'middle';
		ctx.fillStyle = '#9a9a9a';
		ctx.strokeStyle = '#555555';
		ctx.lineWidth = 1;
		ctx.beginPath();

		for (let index = start; index <= end; index += 1) {
			const position = Math.round(offset + index * tilePx) + 0.5;
			if (position < 0 || position > length) continue;

			const major = index % majorStep === 0;
			const minor = index % minorStep === 0;
			if (!major && !minor) continue;

			const tick = major ? 7 : 3;
			if (orient === 'x') {
				ctx.moveTo(position, height - 1);
				ctx.lineTo(position, height - tick);
			} else {
				ctx.moveTo(width - 1, position);
				ctx.lineTo(width - tick, position);
			}
		}
		ctx.stroke();

		for (let index = start; index <= end; index += 1) {
			if (index % majorStep !== 0) continue;
			const position = Math.round(offset + index * tilePx);
			if (position < -20 || position > length + 20) continue;

			if (orient === 'x') {
				ctx.fillText(String(index), position + 3, 6);
			} else {
				ctx.save();
				ctx.translate(6, position + 3);
				ctx.rotate(-Math.PI / 2);
				ctx.textAlign = 'left';
				ctx.fillText(String(index), 0, 0);
				ctx.restore();
			}
		}

		if (hover) {
			const index = orient === 'x' ? hover.x : hover.y;
			const position = Math.round(offset + index * tilePx) + 0.5;
			if (position >= 0 && position <= length) {
				ctx.strokeStyle = 'rgba(111, 159, 216, 0.9)';
				ctx.beginPath();
				if (orient === 'x') {
					ctx.moveTo(position, 0);
					ctx.lineTo(position, height - 1);
				} else {
					ctx.moveTo(0, position);
					ctx.lineTo(width - 1, position);
				}
				ctx.stroke();
			}
		}
	});
</script>

<canvas bind:this={canvas} class="block" aria-hidden="true" style="image-rendering: pixelated"
></canvas>
