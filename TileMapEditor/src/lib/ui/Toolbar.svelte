<script lang="ts">
	import Icon from '@iconify/svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import { icons } from './icons';

	let { editor }: { editor: EditorState } = $props();

	const toolInfo = $derived.by(() => {
		switch (editor.tool) {
			case 'select':
				return { name: 'Select', icon: icons.mousePointer, key: 'V' };
			case 'paint':
				return { name: 'Paint', icon: icons.paintbrush, key: 'B' };
			case 'erase':
				return { name: 'Erase', icon: icons.eraser, key: 'E' };
			case 'object':
				return { name: 'Place object', icon: icons.diamond, key: 'O' };
			case 'spawn':
				return { name: 'Set spawn point', icon: icons.crosshair, key: 'S' };
		}
	});
</script>

<section class="dock">
	<div class="panel-title">
		<span>Tool options</span>
		<kbd class="mono text-[10.5px]" style="color: var(--text-faint)">{toolInfo.key}</kbd>
	</div>
	<div class="panel-body flex flex-col" style="gap: 7px">
		<div class="flex flex-row" style="gap:9px">
			<div class="flex items-center gap-2">
				<Icon icon={toolInfo.icon} width="14" height="14" />
				<span class="text-[11.5px]" style="color: var(--text)">{toolInfo.name}</span>
			</div>

			{#if editor.tool === 'paint'}
				{#if editor.activeTileId && editor.activeTile}
					<div class="flex items-center gap-2">
						<span
							class="h-6 w-6 shrink-0 border"
							style="background: {editor.activeTile.visual.color}; border-color: #131313"
						></span>
						<div class="min-w-0 flex-1">
							<div class="mono truncate text-[10.5px]" style="color: var(--text)">
								{editor.activeTileId}
							</div>
							<div class="t-meta">
								{editor.activeTile.solid ? 'solid' : 'passable'}
							</div>
						</div>
					</div>
				{:else}
					<p class="t-meta leading-4">
						No paint tile selected. Pick one in the Tileset or Tile tab.
					</p>
				{/if}
			{:else if editor.tool === 'erase'}
				<p class="t-meta leading-4">
					Clears cells on <span class="mono" style="color: var(--text)"
						>{editor.activeLayer?.id ?? 'the active layer'}</span
					>.
				</p>
			{:else if editor.tool === 'select'}
				{#if editor.selectedCell}
					<p class="t-meta leading-4">
						Cell <span class="mono" style="color: var(--text)"
							>{editor.selectedCell.x}, {editor.selectedCell.y}</span
						>
						{#if editor.editingObjectId}
							· <span class="mono" style="color: var(--text)">{editor.editingObjectId}</span>
						{/if}
					</p>
				{:else}
					<p class="t-meta leading-4">Click a cell to select it. Objects open in the Object tab.</p>
				{/if}
			{:else if editor.tool === 'object'}
				{#if editor.activeObjectType}
					<div class="flex items-center gap-2">
						<span
							class="h-3 w-3 shrink-0 rotate-45 border"
							style="background: {editor.activeObjectType.color}; border-color: #131313"
						></span>
						<div class="mono min-w-0 flex-1 truncate text-[10.5px]" style="color: var(--text)">
							{editor.activeObjectType.id}
						</div>
					</div>
				{:else}
					<p class="t-meta leading-4">No object type selected. Create one in the Objects tab.</p>
				{/if}
			{:else}
				<p class="mono text-[10.5px]" style="color: var(--text)">
					spawn {editor.map.spawn.x}, {editor.map.spawn.y}
				</p>
			{/if}
		</div>

		<div class="divider"></div>

		<div class="flex items-center gap-1">
			<button
				class="tool-btn"
				class:active={editor.showGrid}
				aria-pressed={editor.showGrid}
				onclick={() => (editor.showGrid = !editor.showGrid)}
				title="Toggle grid (G)"
			>
				<Icon icon={icons.grid} width="13" height="13" />
				Grid
			</button>
			<button
				class="tool-btn"
				class:active={editor.showTilesetArt}
				aria-pressed={editor.showTilesetArt}
				onclick={() => (editor.showTilesetArt = !editor.showTilesetArt)}
				title="Overlay the imported tileset image"
			>
				Artwork
			</button>
		</div>
	</div>
</section>
