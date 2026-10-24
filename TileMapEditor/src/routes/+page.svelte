<script lang="ts">
	import { EditorState } from '#lib/editor-state.svelte';
	import MapCanvas from '#lib/ui/MapCanvas.svelte';
	import MapSettings from '#lib/ui/MapSettings.svelte';
	import ValidationPanel from '#lib/ui/ValidationPanel.svelte';
	import Toolbar from '#lib/ui/Toolbar.svelte';
	import TilesetBrowser from '#lib/ui/TilesetBrowser.svelte';
	import TileInspector from '#lib/ui/TileInspector.svelte';
	import LayerPanel from '#lib/ui/LayerPanel.svelte';

	const editor = new EditorState();
</script>

<svelte:head>
	<title>Tilemap Editor · Thirty Nights</title>
	<meta
		name="description"
		content="Browser-based tilemap editor for Thirty Nights. Import a tileset, paint layers, validate and export game-ready JSON."
	/>
</svelte:head>

<div class="flex h-screen w-screen flex-col overflow-hidden" style="background: var(--bg)">
	<header
		class="flex h-10 shrink-0 items-center gap-3 border-b px-3"
		style="border-color: var(--border); background: #0e1116"
	>
		<div class="flex items-center gap-2">
			<span class="h-2 w-2 rounded-full" style="background: var(--accent)"></span>
			<span class="text-[10px] font-bold tracking-[0.18em]" style="color: var(--text-faint)"
				>THIRTY NIGHTS</span
			>
		</div>
		<span class="h-4 w-px" style="background: var(--border-strong)"></span>
		<h1 class="text-[12.5px] font-semibold" style="color: var(--text)">Tilemap editor</h1>
		<span class="mono truncate text-[11px]" style="color: var(--text-faint)">
			{editor.map.name} · {editor.map.id}
		</span>
		<div class="ml-auto flex items-center gap-2">
			<span class="badge">format v{editor.map.format_version}</span>
			<span class="hidden text-[10px] md:inline" style="color: var(--text-faint)"
				>Runs locally in your browser</span
			>
		</div>
	</header>

	<main
		class="grid min-h-0 flex-1"
		style="grid-template-columns: minmax(240px, 280px) minmax(0, 1fr) minmax(300px, 340px)"
	>
		<aside
			class="flex min-h-0 flex-col overflow-y-auto border-r"
			style="border-color: var(--border); background: var(--panel)"
		>
			<MapSettings {editor} />
			<ValidationPanel {editor} />
		</aside>

		<section class="flex min-h-0 min-w-0 flex-col">
			<Toolbar {editor} />
			<div class="min-h-0 flex-1">
				<MapCanvas {editor} />
			</div>
		</section>

		<aside
			class="flex min-h-0 flex-col border-l"
			style="border-color: var(--border); background: var(--panel)"
		>
			<div class="flex shrink-0 border-b" style="border-color: var(--border)">
				<button
					class="tab"
					class:active={editor.tab === 'tileset'}
					onclick={() => (editor.tab = 'tileset')}
				>
					Tileset
				</button>
				<button
					class="tab"
					class:active={editor.tab === 'inspector'}
					onclick={() => (editor.tab = 'inspector')}
				>
					Tile inspector
				</button>
			</div>
			<div class="min-h-0 flex-1 overflow-y-auto">
				{#if editor.tab === 'tileset'}
					<TilesetBrowser {editor} />
				{:else}
					<TileInspector {editor} />
				{/if}
				<LayerPanel {editor} />
			</div>
		</aside>
	</main>
</div>
