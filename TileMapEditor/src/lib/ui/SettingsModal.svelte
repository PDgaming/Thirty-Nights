<script lang="ts">
	import { onMount, tick, untrack } from 'svelte';
	import hljs from 'highlight.js/lib/core';
	import jsonLanguage from 'highlight.js/lib/languages/json';
	import type { EditorState } from '#lib/editor-state.svelte';
	import { MAX_AUTOSAVE_DELAY_MS, MIN_AUTOSAVE_DELAY_MS } from '#lib/editor-state.svelte';
	import { DEFAULT_MAP_SCHEMA, parseMapSchema } from '#lib/map-schema';

	hljs.registerLanguage('json', jsonLanguage);

	let { editor, onclose }: { editor: EditorState; onclose: () => void } = $props();

	let seconds = $state(untrack(() => editor.autosaveDelayMs) / 1000);
	let secondsInput: HTMLInputElement | undefined = $state();
	let schemaText = $state(
		JSON.stringify(
			untrack(() => editor.mapSchema),
			null,
			2
		)
	);
	let schemaError = $state('');
	let schemaFileInput: HTMLInputElement | undefined = $state();
	let codePreview: HTMLPreElement | undefined = $state();
	let codeInput: HTMLTextAreaElement | undefined = $state();

	const highlighted = $derived(highlightJson(schemaText));

	function highlightJson(text: string): string {
		try {
			return hljs.highlight(text, { language: 'json', ignoreIllegals: true }).value;
		} catch {
			return escapeHtml(text);
		}
	}

	function escapeHtml(text: string): string {
		return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	}

	onMount(() => {
		secondsInput?.focus();
		secondsInput?.select();
	});

	function syncCodeScroll() {
		if (!codePreview || !codeInput) return;
		codePreview.scrollTop = codeInput.scrollTop;
		codePreview.scrollLeft = codeInput.scrollLeft;
	}

	/** Keep the editor usable for JSON: Tab indents instead of leaving the field. */
	async function handleCodeKeydown(event: KeyboardEvent) {
		if (event.key !== 'Tab' || !codeInput) return;
		event.preventDefault();

		const { selectionStart, selectionEnd, value } = codeInput;
		const indent = '  ';
		schemaText = value.slice(0, selectionStart) + indent + value.slice(selectionEnd);

		// Let Svelte apply the new value before restoring the caret.
		await tick();
		if (!codeInput) return;
		codeInput.selectionStart = selectionStart + indent.length;
		codeInput.selectionEnd = selectionStart + indent.length;
		syncCodeScroll();
	}

	function resetSchema() {
		schemaText = JSON.stringify(DEFAULT_MAP_SCHEMA, null, 2);
		schemaError = '';
	}

	async function handleSchemaFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		try {
			schemaText = await file.text();
			schemaError = '';
		} catch {
			schemaError = 'Could not read the schema file.';
		}
	}

	function save() {
		const delay = Number(seconds);
		if (!Number.isFinite(delay) || delay <= 0) {
			editor.notify('Autosave delay must be a positive number of seconds', 'error');
			return;
		}

		let schema;
		try {
			schema = parseMapSchema(schemaText);
		} catch (error) {
			schemaError = error instanceof Error ? error.message : 'Could not parse the schema.';
			return;
		}

		editor.setAutosaveDelay(delay * 1000);
		editor.setMapSchema(schema);
		onclose();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			onclose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="modal-backdrop">
	<button type="button" class="modal-dismiss" aria-label="Close settings" onclick={onclose}
	></button>
	<div class="modal-card" role="dialog" aria-modal="true" aria-labelledby="settings-title">
		<div id="settings-title" class="modal-title">Editor settings</div>

		<form
			class="flex flex-col gap-3"
			onsubmit={(event) => {
				event.preventDefault();
				save();
			}}
		>
			<label>
				<span class="field-label">Autosave delay (seconds)</span>
				<input
					bind:this={secondsInput}
					class="field"
					type="number"
					min={MIN_AUTOSAVE_DELAY_MS / 1000}
					max={MAX_AUTOSAVE_DELAY_MS / 1000}
					step="1"
					bind:value={seconds}
				/>
			</label>
			<p class="text-[10.5px] leading-4" style="color: var(--text-faint)">
				Edits are written to browser storage at most once per delay, from{' '}
				{MIN_AUTOSAVE_DELAY_MS / 1000} to {MAX_AUTOSAVE_DELAY_MS / 1000} seconds. Pending changes are
				always flushed when the page is hidden or closed, and Ctrl+S saves immediately.
			</p>

			<div class="divider"></div>

			<div class="flex items-center justify-between gap-2">
				<span class="field-label" style="margin: 0">Map JSON schema</span>
				<div class="flex items-center gap-1.5">
					<input
						bind:this={schemaFileInput}
						type="file"
						accept=".json,application/json"
						class="hidden"
						onchange={handleSchemaFile}
					/>
					<button type="button" class="btn btn-xs" onclick={() => schemaFileInput?.click()}>
						Upload
					</button>
					<button type="button" class="btn btn-xs" onclick={resetSchema}>Reset</button>
				</div>
			</div>

			<div class="code-editor">
				<pre bind:this={codePreview} aria-hidden="true"><code>{@html highlighted}{'\n'}</code></pre>
				<textarea
					bind:this={codeInput}
					bind:value={schemaText}
					spellcheck="false"
					autocomplete="off"
					autocapitalize="off"
					wrap="off"
					aria-label="Map JSON schema"
					oninput={() => (schemaError = '')}
					onkeydown={handleCodeKeydown}
					onscroll={syncCodeScroll}></textarea>
			</div>

			<p class="text-[10.5px] leading-4" style="color: var(--text-faint)">
				The exported JSON follows this schema's property order. Extra properties are exported with
				their <span class="mono">default</span> value. The editor's fields (format version, id,
				name, tile size, dimensions, tiles, layers, objects, spawn) must stay in
				<span class="mono">properties</span> and <span class="mono">required</span>.
			</p>

			{#if schemaError}
				<p class="text-[11px] leading-4" style="color: var(--danger)" role="alert">{schemaError}</p>
			{/if}

			<div class="mt-1 flex justify-end gap-2">
				<button type="button" class="btn" onclick={onclose}>Cancel</button>
				<button type="submit" class="btn btn-primary">Save</button>
			</div>
		</form>
	</div>
</div>
