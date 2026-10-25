<script lang="ts">
	import Icon from '@iconify/svelte';
	import type { EditorState } from '#lib/editor-state.svelte';
	import type { EditorTool } from '#lib/renderer/draw-map';
	import { icons } from './icons';

	let { editor }: { editor: EditorState } = $props();

	const tools: {
		id: EditorTool;
		label: string;
		hint: string;
		icon: (typeof icons)[keyof typeof icons];
	}[] = [
		{
			id: 'paint',
			label: 'Paint',
			hint: 'Paint with the selected tile (B)',
			icon: icons.paintbrush
		},
		{
			id: 'erase',
			label: 'Erase',
			hint: 'Erase cells on the active layer (E)',
			icon: icons.eraser
		},
		{ id: 'object', label: 'Object', hint: 'Place object markers (O)', icon: icons.diamond },
		{ id: 'spawn', label: 'Spawn', hint: 'Place the spawn point (S)', icon: icons.crosshair }
	];
</script>

<div class="tool-grid" role="toolbar" aria-label="Tools" aria-orientation="vertical">
	{#each tools as tool (tool.id)}
		<button
			class="tool-tile p-2"
			class:active={editor.tool === tool.id}
			aria-pressed={editor.tool === tool.id}
			aria-label={tool.label}
			title={tool.hint}
			onclick={() => (editor.tool = tool.id)}
		>
			<Icon icon={tool.icon} width="16" height="16" />
			<span class="tool-tile-label">{tool.label}</span>
		</button>
	{/each}
</div>
