<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';

	let { editor }: { editor: EditorState } = $props();

	const hint = $derived.by(() => {
		switch (editor.tool) {
			case 'paint':
				return 'Paint: drag to draw · right-click erases';
			case 'erase':
				return 'Erase: drag to clear cells · right-click paints';
			case 'object':
				return 'Objects: click to place · right-click removes';
			case 'spawn':
				return 'Spawn: click a cell to move the start point';
		}
	});
</script>

<footer class="statusbar">
	<span class="status-cell status-hint grow">{hint}</span>

	<span class="status-cell mono" title="Cursor position in cells">
		{editor.hover ? `${editor.hover.x}, ${editor.hover.y}` : '–, –'}
	</span>

	<span class="status-cell">
		<span class="status-label">layer</span>
		<span class="mono" style="color: var(--text)">{editor.activeLayer?.id ?? '–'}</span>
		{#if editor.tool !== 'object' && !editor.activeLayerIsVisible}
			<span class="badge badge-warn">hidden</span>
		{/if}
	</span>

	{#if editor.tool === 'object' && editor.activeObjectType}
		<span class="status-cell">
			<span class="status-label">object</span>
			<span class="mono" style="color: var(--accent)">{editor.activeObjectType.id}</span>
		</span>
	{:else if editor.activeTileId}
		<span class="status-cell">
			<span class="status-label">tile</span>
			<span class="mono" style="color: var(--accent)">{editor.activeTileId}</span>
		</span>
	{/if}

	<span class="status-cell mono" title="Map size in cells">
		{editor.map.width}×{editor.map.height} · {editor.map.tile_size}px
	</span>

	<span class="status-cell" style="gap: 2px; padding: 0 4px">
		<button
			class="icon-btn"
			style="width: 18px; height: 18px"
			onclick={() => editor.zoomBy(1 / 1.25)}
			title="Zoom out"
			aria-label="Zoom out">−</button
		>
		<button
			class="mono min-w-11 px-1 text-[10.5px]"
			style="color: var(--text)"
			onclick={() => editor.setZoomPercent(100)}
			title="Reset zoom to 100%"
		>
			{Math.round(editor.view.scale * 100)}%
		</button>
		<button
			class="icon-btn"
			style="width: 18px; height: 18px"
			onclick={() => editor.zoomBy(1.25)}
			title="Zoom in"
			aria-label="Zoom in">+</button
		>
		<button class="btn btn-xs" style="height: 18px" onclick={() => editor.fitView()}>Fit</button>
	</span>
</footer>
