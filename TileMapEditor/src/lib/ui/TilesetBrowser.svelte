<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import { MAX_TILESET_SLICES } from '#lib/tileset/slicing';
	import { wheelNumber } from './wheel-number';
	import TilesetPreview from './TilesetPreview.svelte';

	let { editor }: { editor: EditorState } = $props();

	let fileInput: HTMLInputElement | undefined = $state();
	let dragActive = $state(false);

	const previewCount = $derived(editor.getPreviewSlices().length);
	const selectedSlice = $derived(editor.sliceForActiveTile);
	const selectedTile = $derived(
		editor.activeTileId ? editor.map.tiles[editor.activeTileId] : undefined
	);

	async function handleFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		await editor.importTileset(file);
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy';
		dragActive = true;
	}

	function handleDragLeave(event: DragEvent) {
		const related = event.relatedTarget;
		if (related instanceof Node && (event.currentTarget as HTMLElement).contains(related)) return;
		dragActive = false;
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragActive = false;
		const file = event.dataTransfer?.files?.[0];
		if (file) void editor.importTileset(file);
	}
</script>

<div class="flex flex-col">
	<div class="panel-title">
		<span>Tileset browser</span>
		{#if editor.tileset.image}
			<span class="badge">{editor.tileset.width}×{editor.tileset.height}</span>
		{/if}
	</div>

	<div class="panel-body flex-col flex">
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*"
			class="hidden"
			onchange={handleFile}
		/>
		<div
			class="dropzone"
			class:loaded={editor.tileset.image}
			class:drag-over={dragActive}
			role="group"
			aria-label="Tileset image drop zone"
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
			ondrop={handleDrop}
		>
			{#if editor.tileset.image}
				<button
					type="button"
					class="dropzone-bar"
					onclick={() => fileInput?.click()}
					title="Drop a new image or click to replace"
				>
					<span class="mono min-w-0 flex-1 truncate">{editor.tileset.fileName}</span>
					<span class="shrink-0 text-[10.5px]" style="color: var(--text-faint)">
						click to replace
					</span>
				</button>
				<TilesetPreview {editor} />
			{:else}
				<button
					type="button"
					class="dropzone-empty"
					onclick={() => fileInput?.click()}
					title="Drop a tileset image or click to browse"
				>
					<span>Drop a tileset image here</span>
					<span style="color: var(--text-faint)">or click to browse</span>
				</button>
			{/if}
		</div>

		{#if editor.tileset.image}
			<div class="divider"></div>
			<div>
				<div class="mb-2 flex items-center justify-between">
					<span class="field-label" style="margin: 0">Slicing</span>
					<span class="badge" class:badge-danger={previewCount === 0}>
						{previewCount} tiles
					</span>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<label>
						<span class="field-label">Source tile W</span>
						<input
							class="field-sm"
							type="number"
							use:wheelNumber
							min="1"
							bind:value={editor.slicing.tileWidth}
						/>
					</label>
					<label>
						<span class="field-label">Source tile H</span>
						<input
							class="field-sm"
							type="number"
							use:wheelNumber
							min="1"
							bind:value={editor.slicing.tileHeight}
						/>
					</label>
					<label>
						<span class="field-label">Spacing</span>
						<input
							class="field-sm"
							type="number"
							use:wheelNumber
							min="0"
							bind:value={editor.slicing.spacing}
						/>
					</label>
					<label>
						<span class="field-label">Margin X</span>
						<input
							class="field-sm"
							type="number"
							use:wheelNumber
							min="0"
							bind:value={editor.slicing.marginX}
						/>
					</label>
					<label>
						<span class="field-label">Margin Y</span>
						<input
							class="field-sm"
							type="number"
							use:wheelNumber
							min="0"
							bind:value={editor.slicing.marginY}
						/>
					</label>
				</div>
				<div class="mt-2 flex items-center justify-between gap-2">
					<p class="text-[10.5px] leading-4" style="color: var(--text-faint)">
						Map tile size is <span class="mono">{editor.map.tile_size}px</span> (game units). Source tiles
						are scaled into it.
					</p>
					<button
						class="btn btn-xs"
						onclick={() => editor.applyTilesetSlicing()}
						disabled={previewCount === 0 || previewCount > MAX_TILESET_SLICES}
					>
						Apply
					</button>
				</div>
				{#if previewCount > MAX_TILESET_SLICES}
					<p class="mt-1 text-[10.5px]" style="color: var(--danger)">
						Too many tiles ({previewCount}); adjust the slicing settings.
					</p>
				{/if}
			</div>
		{/if}

		{#if selectedTile && editor.activeTileId}
			<div class="divider"></div>
			<div class="flex items-center gap-3">
				<div class="min-w-0 flex-1">
					<div class="field-label" style="margin: 0">Selected tile</div>
					<div class="mono truncate text-[12px]" style="color: var(--text)">
						{editor.activeTileId}
					</div>
					{#if selectedSlice}
						<div class="mono mt-0.5 text-[10.5px]" style="color: var(--text-faint)">
							source {selectedSlice.sx},{selectedSlice.sy} · {selectedSlice.sw}×{selectedSlice.sh}
						</div>
					{:else}
						<div class="mt-0.5 text-[10.5px]" style="color: var(--text-faint)">
							No image binding
						</div>
					{/if}
				</div>
				<span
					class="h-8 w-8 shrink-0 rounded-sm border"
					style="background: {selectedTile.visual.color}; border-color: var(--border-strong)"
					title={selectedTile.visual.color}
				></span>
				<span class="badge" class:badge-warn={selectedTile.solid}>
					{selectedTile.solid ? 'solid' : 'passable'}
				</span>
			</div>
		{:else}
			<p class="text-[10.5px] leading-4" style="color: var(--text-faint)">
				Click a tile in the image grid to make it the active paint tile.
			</p>
		{/if}
	</div>
</div>
