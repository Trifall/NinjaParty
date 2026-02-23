<script lang="ts">
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<svelte:head>
	{#if page.url.searchParams.has('embed')}
		<style>
			body {
				background: transparent !important;
			}
		</style>
	{/if}
</svelte:head>

<div class="flex-1">
	{@render children()}
</div>
