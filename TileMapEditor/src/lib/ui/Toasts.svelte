<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import type { EditorState } from '#lib/editor-state.svelte';

	let { editor }: { editor: EditorState } = $props();
</script>

<div class="toast-wrap" aria-live="polite" aria-atomic="false">
	{#if editor.notice}
		{#key editor.notice}
			<div
				class="toast {editor.notice.kind}"
				role="status"
				in:fly={{ y: 7, duration: 170 }}
				out:fade={{ duration: 120 }}
			>
				<span class="min-w-0 truncate" title={editor.notice.text}>{editor.notice.text}</span>
				<button
					class="toast-dismiss"
					onclick={() => editor.dismissNotice()}
					title="Dismiss"
					aria-label="Dismiss notification"
				>
					×
				</button>
			</div>
		{/key}
	{/if}
</div>
