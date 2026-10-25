<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import { COLLISION_TYPES, type CollisionName, type LayerData } from '#lib/map-model/types';
	import Icon from '@iconify/svelte';
	import { tick } from 'svelte';
	import { icons } from './icons';

	let { editor }: { editor: EditorState } = $props();

	let newLayerId = $state('');
	let newCollision = $state<CollisionName>('none');
	let newLayerOpen = $state(false);
	let newLayerInput: HTMLInputElement | undefined = $state();
	let renameDraft = $state('');
	let dragLayerId = $state<string | null>(null);
	let dropTargetId = $state<string | null>(null);
	let dropBefore = $state(true);

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
			newLayerOpen = false;
		}
	}

	async function toggleNewLayer() {
		newLayerOpen = !newLayerOpen;
		if (newLayerOpen) {
			await tick();
			newLayerInput?.focus();
		}
	}

	function commitRename() {
		if (!activeLayer) return;
		editor.renameLayer(activeLayer.id, renameDraft);
		renameDraft = editor.activeLayer?.id ?? '';
	}

	function deleteLayer(layer: LayerData) {
		if (editor.map.layers.length <= 1) return;
		const cellCount = editor.layerCellCount(layer.id);
		if (cellCount > 0) {
			const confirmed = window.confirm(
				`Layer "${layer.id}" contains ${cellCount} painted cells. Delete it and its contents?`
			);
			if (!confirmed) return;
		}
		editor.deleteLayer(layer.id);
	}

	function layerRowStyle(layer: LayerData): string {
		const shadows: string[] = [];
		if (dragLayerId && dropTargetId === layer.id && dragLayerId !== layer.id) {
			shadows.push(dropBefore ? 'inset 0 2px 0 var(--accent)' : 'inset 0 -2px 0 var(--accent)');
		}
		if (layer.id === editor.activeLayerId) shadows.push('inset 2px 0 0 var(--accent)');

		const background =
			layer.id === editor.activeLayerId ? 'rgba(111, 159, 216, 0.13)' : 'transparent';
		const opacity = dragLayerId === layer.id ? '0.4' : '1';
		return `background: ${background}; box-shadow: ${
			shadows.length > 0 ? shadows.join(', ') : 'none'
		}; opacity: ${opacity}`;
	}

	function startDrag(event: DragEvent, layerID: string) {
		dragLayerId = layerID;
		dropTargetId = null;
		dropBefore = true;
		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			event.dataTransfer.setData('text/plain', layerID);
		}
	}

	function dragOverLayer(event: DragEvent, layerID: string) {
		if (!dragLayerId || dragLayerId === layerID) return;
		event.preventDefault();
		if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';

		const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
		dropBefore = event.clientY < rect.top + rect.height / 2;
		dropTargetId = layerID;
	}

	function dragLeaveLayer(event: DragEvent, layerID: string) {
		if (dropTargetId !== layerID) return;
		const related = event.relatedTarget as Node | null;
		if (!related || !(event.currentTarget as HTMLElement).contains(related)) {
			dropTargetId = null;
		}
	}

	function dropLayer(event: DragEvent, layerID: string) {
		event.preventDefault();
		const sourceID = dragLayerId;
		if (!sourceID || sourceID === layerID) {
			resetDrag();
			return;
		}

		const displayed = layersDescending.map((layer) => layer.id);
		const sourceIndex = displayed.indexOf(sourceID);
		if (sourceIndex === -1) {
			resetDrag();
			return;
		}

		displayed.splice(sourceIndex, 1);
		const targetIndex = displayed.indexOf(layerID);
		if (targetIndex === -1) {
			resetDrag();
			return;
		}
		displayed.splice(dropBefore ? targetIndex : targetIndex + 1, 0, sourceID);

		// The list is drawn top-first, while z-index order is bottom-first.
		editor.reorderLayers([...displayed].reverse());
		resetDrag();
	}

	function resetDrag() {
		dragLayerId = null;
		dropTargetId = null;
		dropBefore = true;
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
		<div class="flex items-center gap-1">
			<span class="badge">{editor.map.layers.length}</span>
			<button
				class="icon-btn"
				class:active={newLayerOpen}
				onclick={toggleNewLayer}
				aria-expanded={newLayerOpen}
				aria-label="Toggle new layer form"
				title={newLayerOpen ? 'Hide new layer form' : 'New layer'}
			>
				<Icon icon={icons.plus} width="13" height="13" />
			</button>
		</div>
	</div>

	<div class="panel-body flex flex-col">
		<div class="overflow-hidden rounded border" style="border-color: var(--border)" role="list">
			{#each layersDescending as layer (layer.id)}
				<div
					class="group flex cursor-grab items-stretch border-b last:border-b-0 active:cursor-grabbing"
					style="border-color: var(--border); {layerRowStyle(layer)}"
					role="listitem"
					draggable={true}
					ondragstart={(event) => startDrag(event, layer.id)}
					ondragover={(event) => dragOverLayer(event, layer.id)}
					ondragleave={(event) => dragLeaveLayer(event, layer.id)}
					ondrop={(event) => dropLayer(event, layer.id)}
					ondragend={resetDrag}
				>
					<button
						class="flex min-w-0 flex-1 items-center gap-2 px-2 py-1.5 text-left"
						onclick={() => editor.selectLayer(layer.id)}
						title="Select layer"
					>
						<span
							class="shrink-0 select-none"
							style="color: var(--text-faint)"
							aria-hidden="true"
							title="Drag to reorder"
						>
							<Icon icon={icons.gripVertical} width="10" height="14" />
						</span>
						<span class="mono w-5 shrink-0 text-[10.5px]" style="color: var(--text-faint)"
							>{layer.z_index}</span
						>
						<span class="min-w-0 flex-1">
							<span
								class="mono block truncate text-[11.5px]"
								class:text-[var(--accent)]={layer.id === editor.activeLayerId}
								class:opacity-50={editor.isLayerHidden(layer.id)}>{layer.id}</span
							>
							<span class="text-[10.5px]" style="color: var(--text-faint)">
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
							<Icon icon={icons.eyeOff} width="14" height="14" />
						{:else}
							<Icon icon={icons.eye} width="14" height="14" />
						{/if}
					</button>
					<button
						class="icon-btn layer-delete mr-1 self-center disabled:cursor-not-allowed disabled:opacity-30"
						title={editor.map.layers.length <= 1
							? 'A map needs at least one layer'
							: 'Delete layer'}
						disabled={editor.map.layers.length <= 1}
						onclick={() => deleteLayer(layer)}
					>
						<Icon icon={icons.trash} width="13" height="13" />
					</button>
				</div>
			{/each}
		</div>

		{#if activeLayer}
			<div class="rounded border p-2.5" style="border-color: var(--border); background: #222222">
				<div class="mb-2 flex items-center justify-between">
					<span class="field-label" style="margin: 0">Active layer</span>
					<span class="mono text-[10.5px]" style="color: var(--text-faint)"
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
			</div>
		{/if}

		{#if newLayerOpen}
			<div>
				<div class="field-label">New layer</div>
				<div class="flex items-center gap-1.5">
					<input
						bind:this={newLayerInput}
						class="field-sm mono"
						placeholder={editor.suggestLayerId()}
						value={newLayerId}
						oninput={(event) => (newLayerId = event.currentTarget.value)}
						onkeydown={(event) => {
							if (event.key === 'Enter') addLayer();
							if (event.key === 'Escape') newLayerOpen = false;
						}}
					/>
					<select class="field-sm" bind:value={newCollision}>
						{#each COLLISION_TYPES as collision (collision)}
							<option value={collision}>{collision}</option>
						{/each}
					</select>
					<button class="btn btn-xs" onclick={addLayer}>Add</button>
				</div>
			</div>
		{/if}
	</div>
</section>

<style>
	.layer-delete:hover:not(:disabled) {
		color: var(--danger);
		background: rgba(248, 113, 113, 0.12);
	}
</style>
