<script lang="ts">
	import type { BlindResult } from '$lib/types/stronghold';

	interface Props {
		blind:
			| { isBlindModeEnabled: boolean; blindResult: BlindResult | Record<string, never> }
			| undefined;
	}

	let { blind }: Props = $props();

	function radiansToDegrees(radians: number): number {
		return radians * (180 / Math.PI);
	}

	function getEvaluationText(evaluation: string): string {
		switch (evaluation) {
			case 'NOT_IN_RING':
				return 'not in any ring';
			case 'BAD':
				return 'bad';
			case 'BAD_BUT_IN_RING':
				return 'bad but in ring';
			case 'HIGHROLL_OKAY':
				return 'highroll okay';
			case 'HIGHROLL_GOOD':
				return 'highroll good';
			case 'EXCELLENT':
				return 'excellent';
			default:
				return evaluation;
		}
	}

	function getEvaluationColor(evaluation: string): string {
		switch (evaluation) {
			case 'NOT_IN_RING':
			case 'BAD':
				return 'text-red-400';
			case 'BAD_BUT_IN_RING':
			case 'HIGHROLL_OKAY':
				return 'text-amber-400';
			case 'HIGHROLL_GOOD':
			case 'EXCELLENT':
				return 'text-emerald-400';
			default:
				return 'text-gray-200';
		}
	}
</script>

{#if blind?.isBlindModeEnabled && blind.blindResult && 'xInNether' in blind.blindResult}
	{@const result = blind.blindResult}
	{@const roundedX = Math.round(result.xInNether)}
	{@const roundedZ = Math.round(result.zInNether)}
	{@const improveDeg = radiansToDegrees(result.improveDirection)}
	{@const improveBlocks = Math.round(result.improveDistance)}
	{@const highrollPct = (result.highrollProbability * 100).toFixed(1)}
	{@const threshold = Math.round(result.highrollThreshold)}
	<div class="border-t border-white/20 bg-purple-500/5 px-4 py-3">
		<div class="mb-1 flex items-center gap-2 font-semibold text-purple-300">
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
			Blind Mode
		</div>
		<div class="flex flex-col gap-1 text-sm">
			<span class="text-gray-200">
				Blind coords ({roundedX}, {roundedZ}) are
				<span class={getEvaluationColor(result.evaluation)}
					>{getEvaluationText(result.evaluation)}</span
				>.
			</span>
			<span class="text-gray-200">
				{highrollPct}% chance of &lt;{threshold} block blind
			</span>
			<span class="text-gray-200">
				Head {improveDeg.toFixed(0)} (deg), {improveBlocks} blocks away for better coords
			</span>
		</div>
	</div>
{/if}
