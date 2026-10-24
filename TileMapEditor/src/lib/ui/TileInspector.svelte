<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import { isValidHexColor } from '#lib/map-validation';
	import TilePreview from './TilePreview.svelte';

	let { editor }: { editor: EditorState } = $props();

	let idDraft = $state('');
	let colorDraft = $state('#7a828c');
	let newTileId = $state('');
	let newTileColor = $state('#7a828c');

	const tile = $derived(editor.activeTile);
	const metadata = $derived(editor.activeTileId);
	const slice = $derived(editor.sliceForActiveTile);
	const references = $derived(
		editor.activeTileId ? editor.tileReferenceCount(editor.activeTileId) : 0
	);

	$effect(() => {
		const currentID = editor.activeTileId;
		const currentTile = editor.activeTile;
		idDraft = currentID ?? '';
		colorDraft = currentTile?.visual.color ?? '#7a828c';
	});

	function pickerColor(value: string): string {
		if (/^#[0-9a-fA-F]{6}$/.test(value)) return value;
		if (/^#[0-9a-fA-F]{3}$/.test(value)) {
			return `#${value[1]}${value[1]}${value[2]}${value[2]}${value[3]}${value[3]}`;
		}
		if (/^#[0-9a-fA-F]{8}$/.test(value)) return value.slice(0, 7);
		return '#7a828c';
	}

	function commitRename() {
		if (!editor.activeTileId) return;
		editor.renameActiveTile(idDraft);
		idDraft = editor.activeTileId ?? '';
	}

	function commitColorText() {
		if (!editor.activeTileId) return;
		if (!isValidHexColor(colorDraft)) {
			editor.notify('Colour must be #rgb, #rrggbb, or #rrggbbaa', 'error');
			colorDraft = tile?.visual.color ?? '#7a828c';
			return;
		}
		editor.setTileColor(editor.activeTileId, colorDraft);
	}

	function addTile() {
		if (editor.addTileDefinition(newTileId, newTileColor)) {
			newTileId = '';
		}
	}

	function enterBlur(event: KeyboardEvent) {
		if (event.key === 'Enter') (event.currentTarget as HTMLInputElement).blur();
	}
</script>

<div class="flex flex-col">
	<div class="panel-title">
		<span>Tile inspector</span>
		{#if metadata}
			<span class="badge badge-accent">{metadata}</span>
		{/if}
	</div>

	<div class="panel-body">
		{#if tile && metadata}
			<div class="flex items-start gap-3">
				<div class="checkerboard shrink-0 rounded-sm p-1">
					<TilePreview {editor} tileId={metadata} size={104} />
				</div>
				<div class="min-w-0 flex-1 space-y-2">
					<label>
						<span class="field-label">Tile ID</span>
						<input
							class="field-sm mono"
							value={idDraft}
							oninput={(event) => (idDraft = event.currentTarget.value)}
							onchange={commitRename}
							onkeydown={enterBlur}
						/>
					</label>
					<label>
						<span class="field-label">Colour (#rrggbb)</span>
						<div class="flex items-center gap-1.5">
							<input
								type="color"
								value={pickerColor(tile.visual.color)}
								oninput={(event) => editor.setTileColor(metadata, event.currentTarget.value)}
								title="Tile colour used by the game"
							/>
							<input
								class="field-sm mono min-w-0 flex-1"
								value={colorDraft}
								oninput={(event) => (colorDraft = event.currentTarget.value)}
								onchange={commitColorText}
								onkeydown={enterBlur}
							/>
						</div>
					</label>
				</div>
			</div>

			<label
				class="flex cursor-pointer items-center justify-between rounded border px-2.5 py-2"
				style="border-color: var(--border); background: #0d1014"
			>
				<span>
					<span class="block text-[11.5px] font-semibold">Solid</span>
					<span class="block text-[10px]" style="color: var(--text-faint)"
						>Blocks movement on collision layers</span
					>
				</span>
				<input
					type="checkbox"
					class="toggle"
					checked={tile.solid}
					onchange={(event) => editor.setTileSolid(metadata, event.currentTarget.checked)}
				/>
			</label>

			<div
				class="rounded border px-2.5 py-2 text-[10px] leading-4"
				style="border-color: var(--border)"
			>
				<div class="flex justify-between gap-3">
					<span style="color: var(--text-faint)">Source tile</span>
					<span class="mono" style="color: var(--text-dim)">
						{#if slice}
							{slice.sx},{slice.sy} · {slice.sw}×{slice.sh}
						{:else}
							none
						{/if}
					</span>
				</div>
				<div class="flex justify-between gap-3">
					<span style="color: var(--text-faint)">Map cells</span>
					<span class="mono" style="color: var(--text-dim)">{references}</span>
				</div>
			</div>

			<button
				class="btn btn-danger w-full"
				onclick={() => {
					if (window.confirm(`Delete tile "${metadata}"?`)) editor.deleteActiveTile();
				}}
			>
				Delete tile
			</button>
			<p class="text-[10px] leading-4" style="color: var(--text-faint)">
				The game renders tiles from <span class="mono">visual.color</span>; the tileset image is an
				editor reference only.
			</p>
		{:else}
			<div
				class="rounded border border-dashed px-3 py-5 text-center text-[11px]"
				style="border-color: var(--border-strong); color: var(--text-faint)"
			>
				Select a tile from the tileset or the grid to inspect it.
			</div>
		{/if}

		<div class="divider"></div>

		<div>
			<div class="field-label">Add tile definition</div>
			<div class="flex items-center gap-1.5">
				<input
					class="field-sm mono min-w-0 flex-1"
					placeholder="tile_id"
					value={newTileId}
					oninput={(event) => (newTileId = event.currentTarget.value)}
					onkeydown={(event) => {
						if (event.key === 'Enter') addTile();
					}}
				/>
				<input type="color" bind:value={newTileColor} title="New tile colour" />
				<button class="btn btn-xs" onclick={addTile} disabled={newTileId.trim() === ''}>Add</button>
			</div>
			<p class="mt-1.5 text-[10px] leading-4" style="color: var(--text-faint)">
				Use this for colour-only tiles that are not tied to a tileset slice.
			</p>
		</div>
	</div>
</div>
