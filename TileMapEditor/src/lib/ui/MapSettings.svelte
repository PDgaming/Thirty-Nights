<script lang="ts">
	import { untrack } from 'svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import { resizeLossLabels } from '#lib/map-operations';
	import { wheelNumber } from './wheel-number';

	let { editor }: { editor: EditorState } = $props();

	let widthDraft = $state('');
	let heightDraft = $state('');

	function readMapDimensions(epoch: number) {
		return untrack(() => ({ epoch, width: editor.map.width, height: editor.map.height }));
	}

	$effect(() => {
		// Resync when the map is replaced or resized, not on every paint.
		const dimensions = readMapDimensions(editor.mapEpoch);
		widthDraft = String(dimensions.width);
		heightDraft = String(dimensions.height);
	});

	const spawnValid = $derived(
		Number.isInteger(editor.map.spawn.x) &&
			Number.isInteger(editor.map.spawn.y) &&
			editor.map.spawn.x >= 0 &&
			editor.map.spawn.y >= 0 &&
			editor.map.spawn.x < editor.map.width &&
			editor.map.spawn.y < editor.map.height
	);

	function commitText(field: 'id' | 'name', value: string) {
		if (field === 'id') editor.setMapId(value);
		else editor.setMapName(value);
	}

	function commitTileSize(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const value = Number(input.value);
		editor.setTileSize(value);
		input.value = String(editor.map.tile_size);
	}

	function applyDimension() {
		const width = Number(widthDraft);
		const height = Number(heightDraft);
		if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0) {
			editor.notify('Map dimensions must be positive integers', 'error');
			widthDraft = String(editor.map.width);
			heightDraft = String(editor.map.height);
			return;
		}
		if (width === editor.map.width && height === editor.map.height) return;

		const losses = resizeLossLabels(editor.planResize(width, height));

		if (losses.length > 0) {
			const confirmed = window.confirm(
				`Shrinking the map will permanently remove ${losses.join(', ')}. Continue?`
			);
			if (!confirmed) {
				widthDraft = String(editor.map.width);
				heightDraft = String(editor.map.height);
				return;
			}
		}

		editor.applyResize(width, height);
		widthDraft = String(editor.map.width);
		heightDraft = String(editor.map.height);
	}

	function commitSpawn(axis: 'x' | 'y', event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const value = Number(input.value);
		if (!Number.isInteger(value)) {
			input.value = String(editor.map.spawn[axis]);
			return;
		}
		editor.setSpawn(
			axis === 'x' ? value : editor.map.spawn.x,
			axis === 'y' ? value : editor.map.spawn.y
		);
	}

	function enterBlur(event: KeyboardEvent) {
		if (event.key === 'Enter') (event.currentTarget as HTMLInputElement).blur();
	}
</script>

<section class="border-b" style="border-color: var(--border)">
	<div class="panel-title">
		<span>Map</span>
		<span class="mono text-[10px]" style="color: var(--text-faint)">{editor.map.id}</span>
	</div>
	<div class="panel-body flex flex-col">
		<div class="grid grid-cols-2 gap-2">
			<label class="col-span-1">
				<span class="field-label">Map ID</span>
				<input
					class="field"
					value={editor.map.id}
					onchange={(event) => commitText('id', event.currentTarget.value)}
					onkeydown={enterBlur}
				/>
			</label>
			<label class="col-span-1">
				<span class="field-label">Tile size</span>
				<input
					class="field"
					type="number"
					use:wheelNumber
					min="1"
					value={editor.map.tile_size}
					onchange={commitTileSize}
					onkeydown={enterBlur}
				/>
			</label>
		</div>

		<label>
			<span class="field-label">Map name</span>
			<input
				class="field"
				value={editor.map.name}
				onchange={(event) => commitText('name', event.currentTarget.value)}
				onkeydown={enterBlur}
			/>
		</label>

		<div class="grid grid-cols-2 gap-2">
			<label>
				<span class="field-label">Width (cells)</span>
				<input
					class="field"
					type="number"
					use:wheelNumber
					min="1"
					value={widthDraft}
					oninput={(event) => (widthDraft = event.currentTarget.value)}
					onchange={applyDimension}
					onkeydown={enterBlur}
				/>
			</label>
			<label>
				<span class="field-label">Height (cells)</span>
				<input
					class="field"
					type="number"
					use:wheelNumber
					min="1"
					value={heightDraft}
					oninput={(event) => (heightDraft = event.currentTarget.value)}
					onchange={applyDimension}
					onkeydown={enterBlur}
				/>
			</label>
		</div>

		<div class="grid grid-cols-2 gap-2">
			<label>
				<span class="field-label">Spawn X</span>
				<input
					class="field"
					type="number"
					use:wheelNumber
					min="0"
					max={editor.map.width - 1}
					value={editor.map.spawn.x}
					style={spawnValid ? '' : 'border-color: var(--danger)'}
					onchange={(event) => commitSpawn('x', event)}
					onkeydown={enterBlur}
				/>
			</label>
			<label>
				<span class="field-label">Spawn Y</span>
				<input
					class="field"
					type="number"
					use:wheelNumber
					min="0"
					max={editor.map.height - 1}
					value={editor.map.spawn.y}
					style={spawnValid ? '' : 'border-color: var(--danger)'}
					onchange={(event) => commitSpawn('y', event)}
					onkeydown={enterBlur}
				/>
			</label>
		</div>
	</div>
</section>
