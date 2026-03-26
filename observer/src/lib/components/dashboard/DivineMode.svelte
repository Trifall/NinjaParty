<script lang="ts">
	import type { DivineResult } from '$lib/types/stronghold';

	interface Props {
		divine:
			| { isDivineModeEnabled: boolean; divineResult: DivineResult | Record<string, never> }
			| undefined;
	}

	let { divine }: Props = $props();
</script>

{#if divine?.isDivineModeEnabled && divine.divineResult && 'fossilXCoordinate' in divine.divineResult}
	{@const result = divine.divineResult}
	<div class="overflow-hidden rounded-lg border border-white/10 bg-amber-500/5 px-4 py-3">
		<div class="mb-1.5 flex items-center gap-2 font-semibold text-amber-300">
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
				<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
				<circle cx="12" cy="12" r="3" />
			</svg>
			Divine Mode
		</div>
		<div class="flex items-center gap-6 text-sm">
			<span class="text-gray-400"
				>Fossil X: <span class="font-semibold text-amber-300">{result.fossilXCoordinate}</span
				></span
			>
			<span class="text-gray-400"
				>Highroll: <span class="text-gray-200">{result.formattedHighrollCoords}</span></span
			>
			<span class="text-gray-400"
				>Safe: <span class="text-gray-200">{result.formattedSafeCoords}</span></span
			>
		</div>
	</div>
{/if}
