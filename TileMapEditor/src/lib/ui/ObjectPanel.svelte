<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';
	import Icon from '@iconify/svelte';
	import { tick } from 'svelte';
	import { icons } from './icons';

	let { editor }: { editor: EditorState } = $props();

	let newTypeId = $state('');
	let newTypeColor = $state('#f472b6');
	let addTypeOpen = $state(false);
	let newTypeInput: HTMLInputElement | undefined = $state();
	let renameDraft = $state('');
	let colorDraft = $state('#f472b6');

	const active = $derived(editor.activeObjectType);
	const activeCount = $derived(active ? editor.objectCountForType(active.id) : 0);

	$effect(() => {
		renameDraft = active?.id ?? '';
		colorDraft = active?.color ?? '#f472b6';
	});

	function addType() {
		if (editor.addObjectType(newTypeId, newTypeColor)) {
			newTypeId = '';
			addTypeOpen = false;
		}
	}

	async function toggleAddType() {
		addTypeOpen = !addTypeOpen;
		if (addTypeOpen) {
			await tick();
			newTypeInput?.focus();
		}
	}

	function commitRename() {
		if (!active) return;
		editor.renameObjectType(active.id, renameDraft);
		renameDraft = editor.activeObjectType?.id ?? '';
	}

	function commitColor(value: string) {
		if (!active) return;
		editor.setObjectTypeColor(active.id, value);
		colorDraft = editor.activeObjectType?.color ?? value;
	}

	function enterBlur(event: KeyboardEvent) {
		if (event.key === 'Enter') (event.currentTarget as HTMLInputElement).blur();
	}
</script>

<section class="flex flex-col">
	<div class="panel-title">
		<span>Object types</span>
		<div class="flex items-center gap-1">
			<span class="badge">{editor.objectTypes.length}</span>
			<button
				class="icon-btn"
				class:active={addTypeOpen}
				onclick={toggleAddType}
				aria-expanded={addTypeOpen}
				aria-label="Toggle new object type form"
				title={addTypeOpen ? 'Hide new object type form' : 'New object type'}
			>
				<Icon icon={icons.plus} width="13" height="13" />
			</button>
		</div>
	</div>

	<div class="panel-body flex flex-col">
		<div class="overflow-hidden rounded border" style="border-color: var(--border)">
			{#each editor.objectTypes as type (type.id)}
				{@const placed = editor.objectCountForType(type.id)}
				<div
					class="flex w-full items-center gap-1.5 border-b pr-1 last:border-b-0"
					style="border-color: var(--border); background: {type.id === editor.activeObjectTypeId
						? 'rgba(111, 159, 216, 0.13)'
						: 'transparent'}; box-shadow: {type.id === editor.activeObjectTypeId
						? 'inset 2px 0 0 var(--accent)'
						: 'none'}"
				>
					<button
						class="flex min-w-0 flex-1 items-center gap-2 py-1.5 pl-2 text-left"
						onclick={() => editor.selectObjectType(type.id)}
						title="Select object type"
					>
						<span
							class="h-3.5 w-3.5 shrink-0 rotate-45 rounded-[2px] border"
							style="background: {type.color}; border-color: var(--border-strong)"
						></span>
						<span
							class="mono min-w-0 flex-1 truncate text-[11.5px]"
							class:text-[var(--accent)]={type.id === editor.activeObjectTypeId}
						>
							{type.id}
						</span>
					</button>
					<span class="badge shrink-0">{placed}</span>
					<button
						class="icon-btn object-delete shrink-0 disabled:cursor-not-allowed disabled:opacity-30"
						disabled={placed > 0}
						title={placed > 0
							? 'Erase placed objects before deleting the type'
							: `Delete type "${type.id}"`}
						onclick={() => editor.deleteObjectType(type.id)}
					>
						<Icon icon={icons.trash} width="13" height="13" />
					</button>
				</div>
			{:else}
				<p class="px-2.5 py-3 text-[11px]" style="color: var(--text-faint)">
					No object types yet. Create one with the + button.
				</p>
			{/each}
		</div>

		{#if active}
			<div class="rounded border p-2.5" style="border-color: var(--border); background: #222222">
				<div class="mb-2 flex items-center justify-between">
					<span class="field-label" style="margin: 0">Active type</span>
					<span class="mono text-[10.5px]" style="color: var(--text-faint)"
						>{activeCount} placed</span
					>
				</div>
				<div class="flex flex-row gap-3 space-y-2">
					<label class="block">
						<span class="field-label">Type ID</span>
						<input
							class="field-sm mono"
							value={renameDraft}
							oninput={(event) => (renameDraft = event.currentTarget.value)}
							onchange={commitRename}
							onkeydown={enterBlur}
						/>
					</label>
					<label class="block">
						<span class="field-label">Colour</span>
						<div class="flex items-center gap-1.5">
							<input
								type="color"
								value={colorDraft}
								oninput={(event) => commitColor(event.currentTarget.value)}
								title="Marker colour in the editor"
							/>
							<input
								class="field-sm mono min-w-0 flex-1"
								value={colorDraft}
								oninput={(event) => (colorDraft = event.currentTarget.value)}
								onchange={(event) => commitColor(event.currentTarget.value)}
								onkeydown={enterBlur}
							/>
						</div>
					</label>
				</div>
			</div>
		{/if}

		{#if addTypeOpen}
			<div>
				<div class="field-label">New object type</div>
				<div class="flex items-center gap-1.5">
					<input
						bind:this={newTypeInput}
						class="field-sm mono min-w-0 flex-1"
						placeholder="chest"
						value={newTypeId}
						oninput={(event) => (newTypeId = event.currentTarget.value)}
						onkeydown={(event) => {
							if (event.key === 'Enter') addType();
							if (event.key === 'Escape') addTypeOpen = false;
						}}
					/>
					<input type="color" bind:value={newTypeColor} title="New marker colour" />
					<button class="btn btn-xs" onclick={addType} disabled={newTypeId.trim() === ''}>
						Add
					</button>
				</div>
			</div>
		{/if}
	</div>
</section>

<style>
	.object-delete:hover:not(:disabled) {
		color: var(--danger);
		background: rgba(248, 113, 113, 0.12);
	}
</style>
