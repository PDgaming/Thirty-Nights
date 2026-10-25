<script lang="ts">
	import type { EditorState } from '#lib/editor-state.svelte';

	let { editor }: { editor: EditorState } = $props();

	let customFilename = $state<string | null>(null);
	const filename = $derived(customFilename ?? editor.suggestedFilename);
	const issues = $derived(editor.issues);
	const errors = $derived(issues.filter((issue) => issue.severity === 'error'));
	const warnings = $derived(issues.filter((issue) => issue.severity === 'warning'));

	function exportMap() {
		editor.exportMap(filename);
	}
</script>

<section>
	<div class="panel-title">
		<span>Validate &amp; export</span>
		{#if errors.length > 0}
			<span class="badge badge-danger">{errors.length} error{errors.length === 1 ? '' : 's'}</span>
		{:else if warnings.length > 0}
			<span class="badge badge-warn"
				>{warnings.length} warning{warnings.length === 1 ? '' : 's'}</span
			>
		{:else}
			<span class="badge badge-accent">Ready</span>
		{/if}
	</div>
	<div class="panel-body flex flex-col">
		<div class="flex flex-wrap gap-x-3 gap-y-1 text-[10.5px]" style="color: var(--text-faint)">
			<span><span class="mono">{Object.keys(editor.map.tiles).length}</span> tiles</span>
			<span><span class="mono">{editor.map.layers.length}</span> layers</span>
			<span><span class="mono">{editor.map.objects.length}</span> objects</span>
		</div>

		{#if issues.length > 0}
			<div
				class="max-h-52 overflow-y-auto rounded border"
				style="border-color: var(--border); background: #222222"
			>
				<ul class="divide-y" style="border-color: var(--border)">
					{#each issues as issue (issue.severity + issue.path + issue.message)}
						<li class="flex gap-2 px-2.5 py-2 text-[11px] leading-4">
							<span
								class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full"
								style="background: {issue.severity === 'error' ? 'var(--danger)' : 'var(--warn)'}"
							></span>
							<div class="min-w-0">
								<div class="mono text-[10.5px]" style="color: var(--text-faint)">{issue.path}</div>
								<div style="color: var(--text-dim)">{issue.message}</div>
							</div>
						</li>
					{/each}
				</ul>
			</div>
		{/if}

		<label>
			<span class="field-label">Export filename</span>
			<input
				class="field"
				value={filename}
				oninput={(event) => (customFilename = event.currentTarget.value)}
				onkeydown={(event) => {
					if (event.key === 'Enter') exportMap();
				}}
			/>
		</label>

		<button class="btn btn-primary w-full" onclick={exportMap} disabled={!editor.canExport}>
			Export JSON
		</button>
		{#if !editor.canExport}
			<p class="text-[10.5px] leading-4" style="color: var(--danger)">
				Fix the validation errors above to enable export.
			</p>
		{/if}
	</div>
</section>
