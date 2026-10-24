<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import { COLLISION_TYPES, type CollisionName } from '#lib/map-model/types';

	let { editor }: { editor: EditorState } = $props();

	let newLayerId = $state('');
	let newCollision = $state<CollisionName>('none');
	let renameDraft = $state('');

	const layersDescending = $derived([...editor.map.layers].sort((a, b) => b.z_index - a.z_index));
	const activeLayer = $derived(editor.activeLayer);
	const activeCellCount = $derived(activeLayer ? editor.layerCellCount(activeLayer.id) : 0);

	$effect(() => {
		const current = editor.map.layers.find((layer) => layer.id === editor.activeLayerId);
		renameDraft = current?.id ?? '';
	});

	function addLayer() {
		const id = newLayerId.trim() || editor.suggestLayerId();
		if (editor.createLayer(id, newCollision)) {
			newLayerId = '';
		}
	}

	function commitRename() {
		if (!activeLayer) return;
		editor.renameLayer(activeLayer.id, renameDraft);
		renameDraft = editor.activeLayer?.id ?? '';
	}

	function deleteActiveLayer() {
		if (!activeLayer) return;
		const cellCount = editor.layerCellCount(activeLayer.id);
		if (cellCount > 0) {
			const confirmed = window.confirm(
				`Layer "${activeLayer.id}" contains ${cellCount} painted cells. Delete it and its contents?`
			);
			if (!confirmed) return;
		}
		editor.deleteLayer(activeLayer.id);
	}

	function move(direction: number) {
		if (!activeLayer) return;
		editor.moveLayer(activeLayer.id, direction);
	}

	function enterBlur(event: KeyboardEvent) {
		if (event.key === 'Enter') (event.currentTarget as HTMLInputElement).blur();
	}

	function collisionLabel(value: string): string {
		if (value === 'tile') return 'tile collision';
		if (value === 'object') return 'object collision';
		return 'no collision';
	}
</script>

<section class="border-t" style="border-color: var(--border)">
	<div class="panel-title">
		<span>Layers</span>
		<span class="badge">{editor.map.layers.length}</span>
	</div>

	<div class="panel-body">
		<div class="overflow-hidden rounded border" style="border-color: var(--border)">
			{#each layersDescending as layer (layer.id)}
				<div
					class="group flex items-stretch border-b last:border-b-0"
					style="border-color: var(--border); background: {layer.id === editor.activeLayerId
						? 'rgba(94, 234, 212, 0.08)'
						: 'transparent'}; box-shadow: {layer.id === editor.activeLayerId
						? 'inset 2px 0 0 var(--accent)'
						: 'none'}"
				>
					<button
						class="flex min-w-0 flex-1 items-center gap-2 px-2 py-1.5 text-left"
						onclick={() => editor.selectLayer(layer.id)}
						title="Select layer"
					>
						<span class="mono w-5 shrink-0 text-[10px]" style="color: var(--text-faint)"
							>{layer.z_index}</span
						>
						<span class="min-w-0 flex-1">
							<span
								class="mono block truncate text-[11.5px]"
								class:text-[var(--accent)]={layer.id === editor.activeLayerId}
								class:opacity-50={editor.isLayerHidden(layer.id)}>{layer.id}</span
							>
							<span class="text-[9.5px]" style="color: var(--text-faint)">
								{collisionLabel(layer.collision)} · {editor.layerCellCount(layer.id)} cells
							</span>
						</span>
					</button>
					<button
						class="icon-btn mr-1 self-center"
						class:active={!editor.isLayerHidden(layer.id)}
						title={editor.isLayerHidden(layer.id) ? 'Show layer' : 'Hide layer'}
						onclick={() => editor.toggleLayerVisibility(layer.id)}
					>
						{#if editor.isLayerHidden(layer.id)}
							<svg
								width="14"
								height="14"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<path d="M2 2l20 20" />
								<path d="M10.7 5.1A10.7 10.7 0 0 1 12 5c5 0 9 4.5 10 7a17 17 0 0 1-3.2 4.2" />
								<path d="M6.6 6.6A16.6 16.6 0 0 0 2 12c1 2.5 5 7 10 7a10.9 10.9 0 0 0 4.4-.9" />
								<path d="M9.9 9.9a3 3 0 0 0 4.2 4.2" />
							</svg>
						{:else}
							<svg
								width="14"
								height="14"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" />
								<circle cx="12" cy="12" r="3" />
							</svg>
						{/if}
					</button>
				</div>
			{/each}
		</div>

		{#if activeLayer}
			<div class="rounded border p-2.5" style="border-color: var(--border); background: #0d1014">
				<div class="mb-2 flex items-center justify-between">
					<span class="field-label" style="margin: 0">Active layer</span>
					<span class="mono text-[10px]" style="color: var(--text-faint)"
						>{activeCellCount} cells</span
					>
				</div>
				<div class="space-y-2">
					<label class="block">
						<span class="field-label">Layer ID</span>
						<input
							class="field-sm mono"
							value={renameDraft}
							oninput={(event) => (renameDraft = event.currentTarget.value)}
							onchange={commitRename}
							onkeydown={enterBlur}
						/>
					</label>
					<label class="block">
						<span class="field-label">Collision mode</span>
						<select
							class="field-sm"
							value={activeLayer.collision}
							onchange={(event) =>
								editor.setLayerCollision(
									activeLayer.id,
									event.currentTarget.value as CollisionName
								)}
						>
							{#each COLLISION_TYPES as collision (collision)}
								<option value={collision}>{collision}</option>
							{/each}
						</select>
					</label>
				</div>
				<div class="mt-2 flex gap-1.5">
					<button
						class="btn btn-xs flex-1"
						onclick={() => move(1)}
						disabled={activeLayer.z_index >= editor.map.layers.length - 1}
						title="Move layer up"
					>
						↑ Up
					</button>
					<button
						class="btn btn-xs flex-1"
						onclick={() => move(-1)}
						disabled={activeLayer.z_index <= 0}
						title="Move layer down"
					>
						↓ Down
					</button>
					<button
						class="btn btn-xs btn-danger flex-1"
						onclick={deleteActiveLayer}
						disabled={editor.map.layers.length <= 1}
					>
						Delete
					</button>
				</div>
			</div>
		{/if}

		<div>
			<div class="field-label">New layer</div>
			<div class="flex items-center gap-1.5">
				<input
					class="field-sm mono min-w-0 flex-1"
					placeholder={editor.suggestLayerId()}
					value={newLayerId}
					oninput={(event) => (newLayerId = event.currentTarget.value)}
					onkeydown={(event) => {
						if (event.key === 'Enter') addLayer();
					}}
				/>
				<select class="field-sm w-24" bind:value={newCollision}>
					{#each COLLISION_TYPES as collision (collision)}
						<option value={collision}>{collision}</option>
					{/each}
				</select>
				<button class="btn btn-xs" onclick={addLayer}>Add</button>
			</div>
		</div>

		<p class="text-[10px] leading-4" style="color: var(--text-faint)">
			Visibility is a workspace-only setting. Hidden layers still export with all their cells.
		</p>
	</div>
</section>
