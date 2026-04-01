<script lang="ts">
	import { tick } from 'svelte';
	import { CircleAlert, Info, TriangleAlert } from '@lucide/svelte';
	import type { ViewMode } from '$lib/stores/settings.svelte';
	import type { InformationMessage } from '$lib/types/stronghold';
	import { sanitize } from '$src/lib/utils/sanitize';
	import { Tooltip, TooltipContent, TooltipTrigger } from '$components/ui/tooltip';

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

	let truncatedMessages = $state<Record<number, boolean>>({});

	function detectTruncation(node: HTMLElement, index: number) {
		const check = () => {
			truncatedMessages[index] = node.scrollWidth > node.clientWidth;
		};

		tick().then(check);

		const ro = new ResizeObserver(check);
		ro.observe(node);

		return {
			destroy() {
				ro.disconnect();
			},
		};
	}
</script>

{#if messages && messages.length > 0}
	{#if viewMode === 'ultra-compact' || viewMode === 'grid'}
		<div class="flex flex-col gap-1 border-t border-white/5 bg-white/[0.02] px-3 py-1.5 text-xs">
			{#each messages as msg, i}
				<div class="flex items-center gap-1.5 overflow-hidden">
					<div
						class="flex shrink-0 items-center gap-1 font-semibold {msg.severity === 'ERROR'
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
						<span>{formatType(msg.type)}:</span>
					</div>
					<Tooltip delayDuration={0} open={truncatedMessages[i] ? undefined : false}>
						<TooltipTrigger
							class="min-w-0 flex-1 {truncatedMessages[i] ? 'cursor-help' : 'cursor-default'} truncate border-none bg-transparent text-left text-gray-300 outline-none"
						>
							<div class="w-full truncate" use:detectTruncation={i}>{@html sanitize(msg.message)}</div>
						</TooltipTrigger>
						<TooltipContent class="max-w-[16rem] text-sm md:max-w-sm">
							<p>{@html sanitize(msg.message)}</p>
						</TooltipContent>
					</Tooltip>
				</div>
			{/each}
		</div>
	{:else}
		<div class="flex flex-col overflow-hidden rounded-lg border border-white/10 bg-white/[0.02]">
			{#each messages as msg, i}
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
					<div class="flex min-w-0 flex-1 flex-col gap-0.5">
						<span
							class="font-semibold {msg.severity === 'ERROR'
								? 'text-red-400/90'
								: msg.severity === 'WARNING'
									? 'text-amber-400/90'
									: 'text-blue-400/90'}">{formatType(msg.type)}</span
						>
						{#if viewMode === 'compact'}
							<Tooltip delayDuration={0} open={truncatedMessages[i] ? undefined : false}>
								<TooltipTrigger
									class="w-full {truncatedMessages[i] ? 'cursor-help' : 'cursor-default'} truncate border-none bg-transparent text-left leading-snug text-gray-300 outline-none"
								>
									<div class="w-full truncate" use:detectTruncation={i}>{@html sanitize(msg.message)}</div>
								</TooltipTrigger>
								<TooltipContent class="max-w-[16rem] text-sm md:max-w-sm">
									<p>{@html sanitize(msg.message)}</p>
								</TooltipContent>
							</Tooltip>
						{:else}
							<span class="leading-snug text-gray-300">{@html sanitize(msg.message)}</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
{/if}
