<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import { isValidHexColor } from '#lib/map-validation';
	import Icon from '@iconify/svelte';
	import { tick } from 'svelte';
	import { icons } from './icons';
	import TilePreview from './TilePreview.svelte';

	let { editor }: { editor: EditorState } = $props();

	let idDraft = $state('');
	let colorDraft = $state('#7a828c');
	let newTileId = $state('');
	let newTileColor = $state('#7a828c');
	let addTileOpen = $state(false);
	let newTileInput: HTMLInputElement | undefined = $state();

	const tile = $derived(editor.activeTile);
	const metadata = $derived(editor.activeTileId);
	const tiles = $derived(Object.entries(editor.map.tiles));

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
			addTileOpen = false;
		}
	}

	async function toggleAddTile() {
		addTileOpen = !addTileOpen;
		if (addTileOpen) {
			await tick();
			newTileInput?.focus();
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

	<div class="panel-body flex flex-col">
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
			<div class="flex items-center gap-2">
				<label
					class="flex flex-1 cursor-pointer items-center justify-between rounded border px-2.5 py-2"
					style="border-color: var(--border); background: #222222"
				>
					<span>
						<span class="block text-[11.5px] font-semibold">Solid</span>
					</span>
					<input
						type="checkbox"
						class="toggle"
						checked={tile.solid}
						onchange={(event) => editor.setTileSolid(metadata, event.currentTarget.checked)}
					/>
				</label>
			</div>
		{:else}
			<div
				class="rounded border border-dashed px-3 py-5 text-center text-[11px]"
				style="border-color: var(--border-strong); color: var(--text-faint)"
			>
				Select a tile from the palette below or the tileset to inspect it.
			</div>
		{/if}

		<div class="divider"></div>

		<div>
			<div class="mb-2 flex items-center justify-between">
				<span class="field-label" style="margin: 0">Tile palette</span>
				<div class="flex items-center gap-1">
					<span class="badge">{tiles.length}</span>
					<button
						class="icon-btn"
						class:active={addTileOpen}
						onclick={toggleAddTile}
						aria-expanded={addTileOpen}
						aria-label="Toggle new tile form"
						title={addTileOpen ? 'Hide new tile form' : 'New tile'}
					>
						<Icon icon={icons.plus} width="13" height="13" />
					</button>
				</div>
			</div>
			{#if tiles.length === 0}
				<p class="text-[10.5px] leading-4" style="color: var(--text-faint)">
					No tiles yet. Add one with the + button or slice a tileset.
				</p>
			{:else}
				<div class="max-h-56 space-y-1 overflow-y-auto pr-0.5">
					{#each tiles as [id, def] (id)}
						{@const cellCount = editor.tileReferenceCount(id)}
						{@const usedByObject = editor.tileIsUsedByObject(id)}
						{@const blocked = cellCount > 0 || usedByObject}
						<div class="tile-row" class:active={id === editor.activeTileId}>
							<button
								class="flex min-w-0 flex-1 items-center gap-1.5 text-left"
								onclick={() => editor.selectTile(id)}
								title="Paint with {id}"
							>
								<span
									class="h-4 w-4 shrink-0 rounded-sm border"
									style="background: {def.visual.color}; border-color: var(--border-strong)"
								></span>
								<span class="mono min-w-0 flex-1 truncate">{id}</span>
								{#if def.solid}
									<span class="badge badge-warn shrink-0">solid</span>
								{/if}
							</button>
							<button
								class="icon-btn tile-delete shrink-0 disabled:cursor-not-allowed disabled:opacity-30"
								disabled={blocked}
								title={cellCount > 0
									? `Used by ${cellCount} map cell${cellCount === 1 ? '' : 's'}`
									: usedByObject
										? 'Used by an object type'
										: `Delete tile "${id}"`}
								onclick={() => {
									if (window.confirm(`Delete tile "${id}"?`)) editor.deleteTile(id);
								}}
							>
								<Icon icon={icons.trash} width="13" height="13" />
							</button>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		{#if addTileOpen}
			<div class="divider"></div>

			<div>
				<div class="field-label">Add tile definition</div>
				<div class="flex items-center gap-1.5">
					<input
						bind:this={newTileInput}
						class="field-sm mono min-w-0 flex-1"
						placeholder="tile_id"
						value={newTileId}
						oninput={(event) => (newTileId = event.currentTarget.value)}
						onkeydown={(event) => {
							if (event.key === 'Enter') addTile();
							if (event.key === 'Escape') addTileOpen = false;
						}}
					/>
					<input type="color" bind:value={newTileColor} title="New tile colour" />
					<button class="btn btn-xs" onclick={addTile} disabled={newTileId.trim() === ''}>
						Add
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.tile-delete:hover:not(:disabled) {
		color: var(--danger);
		background: rgba(248, 113, 113, 0.12);
	}
</style>
