<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';

	let { editor }: { editor: EditorState } = $props();
</script>

<div
	class="flex flex-wrap items-center gap-1.5 border-b px-2.5 py-1.5"
	style="border-color: var(--border); background: var(--panel)"
>
	<div class="flex items-center gap-0.5 border-r pr-1.5" style="border-color: var(--border)">
		<button
			class="tool-btn"
			class:active={editor.tool === 'paint'}
			onclick={() => (editor.tool = 'paint')}
			title="Paint with the selected tile"
		>
			<svg
				width="13"
				height="13"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M3 21l4-1 11-11-3-3L4 17l-1 4z" />
				<path d="M14 6l4 4" />
			</svg>
			Paint
		</button>
		<button
			class="tool-btn"
			class:active={editor.tool === 'erase'}
			onclick={() => (editor.tool = 'erase')}
			title="Erase cells on the active layer"
		>
			<svg
				width="13"
				height="13"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M4 16l8-8 6 6-4 4H6l-2-2z" />
				<path d="M12 8l4-4 6 6-4 4" />
				<path d="M9 20h11" />
			</svg>
			Erase
		</button>
		<button
			class="tool-btn"
			class:active={editor.tool === 'spawn'}
			onclick={() => (editor.tool = 'spawn')}
			title="Place the spawn point"
		>
			<svg
				width="13"
				height="13"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<circle cx="12" cy="12" r="6" />
				<path d="M12 2v4M12 18v4M2 12h4M18 12h4" />
			</svg>
			Spawn
		</button>
	</div>

	<div class="flex items-center gap-1">
		<button
			class="icon-btn"
			class:active={editor.showGrid}
			onclick={() => (editor.showGrid = !editor.showGrid)}
			title="Toggle grid"
		>
			<svg
				width="14"
				height="14"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<rect x="3" y="3" width="18" height="18" rx="1" />
				<path d="M9 3v18M15 3v18M3 9h18M3 15h18" />
			</svg>
		</button>
		<button
			class="icon-btn text-[10px] font-semibold"
			class:active={editor.showTilesetArt}
			onclick={() => (editor.showTilesetArt = !editor.showTilesetArt)}
			title="Toggle tileset artwork overlay"
		>
			ART
		</button>
	</div>

	<div class="ml-auto flex items-center gap-1">
		<button class="icon-btn" onclick={() => editor.zoomBy(1 / 1.25)} title="Zoom out">−</button>
		<button
			class="mono min-w-14 rounded px-1.5 py-0.5 text-[10.5px]"
			style="color: var(--text-dim); background: var(--panel-raised)"
			onclick={() => editor.setZoomPercent(100)}
			title="Reset zoom to 100%"
		>
			{Math.round(editor.view.scale * 100)}%
		</button>
		<button class="icon-btn" onclick={() => editor.zoomBy(1.25)} title="Zoom in">+</button>
		<button class="btn btn-xs" onclick={() => editor.fitView()}>Fit</button>
	</div>

	{#if editor.notice}
		<div
			class="flex max-w-[46%] items-center gap-2 rounded border px-2 py-1 text-[10.5px]"
			style="border-color: {editor.notice.kind === 'error'
				? 'rgba(248,113,113,0.45)'
				: editor.notice.kind === 'success'
					? 'rgba(74,222,128,0.35)'
					: 'var(--border-strong)'}; background: {editor.notice.kind === 'error'
				? 'rgba(248,113,113,0.1)'
				: editor.notice.kind === 'success'
					? 'rgba(74,222,128,0.08)'
					: 'var(--panel-raised)'}; color: {editor.notice.kind === 'error'
				? 'var(--danger)'
				: editor.notice.kind === 'success'
					? 'var(--success)'
					: 'var(--text-dim)'}"
			title={editor.notice.text}
		>
			<span class="truncate">{editor.notice.text}</span>
			<button
				class="icon-btn h-4 w-4 shrink-0"
				onclick={() => editor.dismissNotice()}
				title="Dismiss"
			>
				×
			</button>
		</div>
	{/if}
</div>
