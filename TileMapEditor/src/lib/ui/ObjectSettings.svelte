<script lang="ts">
	import { untrack } from 'svelte';
	import {
		defaultValueForAttribute,
		type EditorState,
		type ObjectAttributeDef
	} from '#lib/editor-state.svelte';

	let { editor, objectId }: { editor: EditorState; objectId: string } = $props();

	const object = untrack(() => editor.map.objects.find((item) => item.id === objectId));
	const attributes = untrack(
		(): ObjectAttributeDef[] =>
			editor.objectTypes.find((type) => type.id === object?.type)?.attributes ?? []
	);

	let x = $state(untrack(() => object?.x ?? 0));
	let y = $state(untrack(() => object?.y ?? 0));
	let values = $state<Record<string, unknown>>(
		untrack(() => {
			const next: Record<string, unknown> = {};
			for (const attribute of attributes) {
				next[attribute.name] = initialValue(attribute);
			}
			return next;
		})
	);
	let error = $state('');

	function initialValue(attribute: ObjectAttributeDef): unknown {
		const current = object?.properties[attribute.name];
		const matches =
			(attribute.type === 'number' && typeof current === 'number') ||
			(attribute.type === 'boolean' && typeof current === 'boolean') ||
			(attribute.type === 'string' && typeof current === 'string');
		return matches ? current : defaultValueForAttribute(attribute.type);
	}

	function persist() {
		if (!object) return;

		const nextX = Number(x);
		const nextY = Number(y);
		if (!Number.isInteger(nextX) || !Number.isInteger(nextY)) {
			error = 'Position must use whole numbers.';
			return;
		}

		const properties = { ...object.properties };
		for (const attribute of attributes) {
			const value = values[attribute.name];
			if (attribute.type === 'number' && (typeof value !== 'number' || !Number.isFinite(value))) {
				error = `"${attribute.name}" must be a number.`;
				return;
			}
			if (attribute.type === 'boolean' && typeof value !== 'boolean') {
				error = `"${attribute.name}" must be true or false.`;
				return;
			}
			if (attribute.type === 'string' && typeof value !== 'string') {
				error = `"${attribute.name}" must be text.`;
				return;
			}
			properties[attribute.name] = value;
		}

		if (!editor.updateObject(object.id, { x: nextX, y: nextY, properties })) {
			error = 'Could not save the object. Check the position and try again.';
			return;
		}
		error = '';
	}

	function beginEdit() {
		editor.beginUndoGroup();
	}

	function endEdit() {
		editor.endUndoGroup();
	}

	function updatePosition(axis: 'x' | 'y', event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		if (input.value.trim() === '') return;
		const value = Number(input.value);
		if (!Number.isInteger(value)) {
			error = 'Position must use whole numbers.';
			return;
		}
		if (axis === 'x') x = value;
		else y = value;
		persist();
	}

	function updateString(name: string, event: Event) {
		values[name] = (event.currentTarget as HTMLInputElement).value;
		persist();
	}

	function updateNumber(name: string, event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		if (input.value.trim() === '') return;
		const value = Number(input.value);
		if (!Number.isFinite(value)) {
			error = `"${name}" must be a number.`;
			return;
		}
		values[name] = value;
		persist();
	}

	function updateBoolean(name: string, event: Event) {
		values[name] = (event.currentTarget as HTMLSelectElement).value === 'true';
		persist();
	}
</script>

{#if object}
	<div class="panel-body flex flex-col gap-3">
		<div class="flex flex-wrap items-center gap-1.5 text-[11px]">
			<span class="mono" style="color: var(--text)">{object.id}</span>
			<span class="badge">{object.type}</span>
			<span class="mono" style="color: var(--text-faint)">layer: {object.layer}</span>
		</div>

		<div class="flex flex-col gap-3">
			<div>
				<div class="field-label">Position</div>
				<div class="grid grid-cols-2 gap-2">
					<label class="block">
						<span class="field-label">X</span>
						<input
							class="field-sm mono"
							type="number"
							step="1"
							value={x}
							oninput={(event) => updatePosition('x', event)}
							onfocus={beginEdit}
							onblur={endEdit}
						/>
					</label>
					<label class="block">
						<span class="field-label">Y</span>
						<input
							class="field-sm mono"
							type="number"
							step="1"
							value={y}
							oninput={(event) => updatePosition('y', event)}
							onfocus={beginEdit}
							onblur={endEdit}
						/>
					</label>
				</div>
			</div>

			<div class="divider"></div>

			<div>
				<div class="mb-1.5 flex items-center justify-between">
					<span class="field-label" style="margin: 0">Properties</span>
					<span class="badge">{attributes.length}</span>
				</div>
				{#if attributes.length === 0}
					<p class="text-[11px]" style="color: var(--text-faint)">
						This object type has no attributes yet.
					</p>
				{:else}
					<div class="flex flex-col gap-2">
						{#each attributes as attribute (attribute.name)}
							<label class="block">
								<span class="mb-0.5 flex items-center justify-between gap-2">
									<span class="field-label mono" style="margin: 0">{attribute.name}</span>
									<span class="mono text-[10px]" style="color: var(--text-faint)">
										{attribute.type}
									</span>
								</span>
								{#if attribute.type === 'boolean'}
									<select
										class="field-sm"
										value={String(values[attribute.name])}
										onchange={(event) => updateBoolean(attribute.name, event)}
									>
										<option value="true">true</option>
										<option value="false">false</option>
									</select>
								{:else if attribute.type === 'number'}
									<input
										class="field-sm mono"
										type="number"
										value={typeof values[attribute.name] === 'number'
											? (values[attribute.name] as number)
											: 0}
										oninput={(event) => updateNumber(attribute.name, event)}
										onfocus={beginEdit}
										onblur={endEdit}
									/>
								{:else}
									<input
										class="field-sm mono"
										value={String(values[attribute.name] ?? '')}
										oninput={(event) => updateString(attribute.name, event)}
										onfocus={beginEdit}
										onblur={endEdit}
									/>
								{/if}
							</label>
						{/each}
					</div>
				{/if}
			</div>

			{#if error}
				<p class="text-[11px] leading-4" style="color: var(--danger)" role="alert">{error}</p>
			{/if}
		</div>
	</div>
{:else}
	<div class="panel-body flex flex-col gap-2">
		<p class="text-[11px]" style="color: var(--text-faint)">This object no longer exists.</p>
		<div class="flex justify-end">
			<button type="button" class="btn" onclick={() => (editor.editingObjectId = null)}>
				Clear
			</button>
		</div>
	</div>
{/if}
