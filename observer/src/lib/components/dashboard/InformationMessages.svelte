<script lang="ts">
	import { CircleAlert, Info, TriangleAlert } from '@lucide/svelte';
	import type { ViewMode } from '$lib/stores/settings.svelte';
	import type { InformationMessage } from '$lib/types/stronghold';

	interface Props {
		messages: InformationMessage[] | undefined;
		viewMode: ViewMode;
	}

	let { messages, viewMode }: Props = $props();

	function formatType(type: string): string {
		return type
			.split('_')
			.map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
			.join(' ');
	}
</script>

{#if messages && messages.length > 0}
	{#if viewMode === 'ultra-compact'}
		<div class="flex flex-col gap-1.5 border-t border-white/5 bg-white/[0.02] px-3 py-1.5 text-xs">
			{#each messages as msg}
				<div class="flex flex-col gap-0.5">
					<div
						class="flex items-center gap-1 font-semibold {msg.severity === 'ERROR'
							? 'text-red-400'
							: msg.severity === 'WARNING'
								? 'text-amber-400'
								: 'text-blue-400'}"
					>
						{#if msg.severity === 'ERROR'}
							<CircleAlert class="h-3 w-3 shrink-0" />
						{:else if msg.severity === 'WARNING'}
							<TriangleAlert class="h-3 w-3 shrink-0" />
						{:else}
							<Info class="h-3 w-3 shrink-0" />
						{/if}
						<span>{formatType(msg.type)}</span>
					</div>
					<span class="leading-snug text-gray-300">{@html msg.message}</span>
				</div>
			{/each}
		</div>
	{:else}
		<div class="flex flex-col border-t border-white/20 bg-white/[0.02]">
			{#each messages as msg}
				<div
					class="flex items-start gap-2 border-b border-white/5 px-4 py-2 last:border-b-0 {viewMode ===
					'compact'
						? 'py-1.5 text-xs'
						: 'py-2.5 text-sm'}"
				>
					<div
						class="mt-0.5 shrink-0 {msg.severity === 'ERROR'
							? 'text-red-400'
							: msg.severity === 'WARNING'
								? 'text-amber-400'
								: 'text-blue-400'}"
					>
						{#if msg.severity === 'ERROR'}
							<CircleAlert class={viewMode === 'compact' ? 'h-3.5 w-3.5' : 'h-4 w-4'} />
						{:else if msg.severity === 'WARNING'}
							<TriangleAlert class={viewMode === 'compact' ? 'h-3.5 w-3.5' : 'h-4 w-4'} />
						{:else}
							<Info class={viewMode === 'compact' ? 'h-3.5 w-3.5' : 'h-4 w-4'} />
						{/if}
					</div>
					<div class="flex flex-col gap-0.5">
						<span
							class="font-semibold {msg.severity === 'ERROR'
								? 'text-red-400/90'
								: msg.severity === 'WARNING'
									? 'text-amber-400/90'
									: 'text-blue-400/90'}">{formatType(msg.type)}</span
						>
						<span class="leading-snug text-gray-300">{@html msg.message}</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
{/if}
