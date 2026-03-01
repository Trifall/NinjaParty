<script lang="ts">
	import type { AllAdvancementsPayload } from '$lib/types/stronghold';

	interface Props {
		allAdvancements: AllAdvancementsPayload | undefined;
	}

	let { allAdvancements }: Props = $props();
</script>

{#if allAdvancements?.isAllAdvancementsModeEnabled}
	<div class="border-t border-white/20 bg-cyan-500/5 px-4 py-3">
		<div class="mb-1.5 flex items-center gap-2 font-semibold text-cyan-300">
			<svg
				class="h-4 w-4"
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
				<circle cx="12" cy="10" r="3" />
			</svg>
			All Advancements
		</div>
		<div class="flex flex-wrap gap-x-6 gap-y-1 text-sm">
			{#each [{ label: 'Spawn', data: allAdvancements.spawn }, { label: 'Outpost', data: allAdvancements.outpost }, { label: 'Monument', data: allAdvancements.monument }, { label: 'Stronghold', data: allAdvancements.stronghold }, { label: 'Deep Dark', data: allAdvancements.deepDark }] as item (item.label)}
				<span class="text-gray-400"
					>{item.label}:
					{#if item.data && 'xInOverworld' in item.data}
						<span class="text-cyan-300"
							>({item.data.xInOverworld?.toFixed(0) ?? '?'}, {item.data.zInOverworld?.toFixed(0) ??
								'?'})</span
						>
					{:else}
						<span class="text-gray-500">-</span>
					{/if}
				</span>
			{/each}
		</div>
	</div>
{/if}
