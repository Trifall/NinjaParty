<script lang="ts">
	import type { ChunkPrediction, PlayerPosition, ResultType } from '$lib/types/stronghold';
	import { settings } from '$lib/stores/settings.svelte';
	import {
		chunkToNether,
		getAngleDiffColor,
		getAngleDifference,
		getAngleToStructure,
		getCertaintyColor,
		normalizeMinecraftAngle,
	} from '$lib/utils/stronghold';

	interface Props {
		resultType: ResultType;
		predictions: ChunkPrediction[];
		pos: PlayerPosition | undefined;
	}

	let { resultType, predictions, pos }: Props = $props();

	function getPredictionDisplayCoords(chunkX: number, chunkZ: number): [number, number] {
		if (settings.displayType === 'FOUR_FOUR') return [16 * chunkX + 4, 16 * chunkZ + 4];
		if (settings.displayType === 'EIGHT_EIGHT') return [16 * chunkX + 8, 16 * chunkZ + 8];
		return [chunkX, chunkZ];
	}

	function getTravelAngleInfo(chunkX: number, chunkZ: number) {
		if (
			pos?.xInOverworld === undefined ||
			pos?.zInOverworld === undefined ||
			pos?.horizontalAngle === undefined
		) {
			return { angle: null, diff: null };
		}
		const coords = getPredictionDisplayCoords(chunkX, chunkZ);
		const angle = getAngleToStructure(pos.xInOverworld, pos.zInOverworld, coords[0], coords[1]);
		const diff = getAngleDifference(angle, pos.horizontalAngle);
		return { angle, diff };
	}
</script>

{#if resultType === 'FAILED'}
	<div
		class="flex items-center justify-center gap-3 border-b border-white/20 bg-red-500/5 px-4 py-3 text-center text-red-400"
	>
		<svg
			class="h-5 w-5 shrink-0"
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<circle cx="12" cy="12" r="10" />
			<line x1="12" y1="8" x2="12" y2="12" />
			<line x1="12" y1="16" x2="12.01" y2="16" />
		</svg>
		<p>Could not determine stronghold chunk. You likely misread one of the eyes.</p>
	</div>
{/if}

{#if resultType !== 'FAILED' && predictions.length > 0}
	<table class="w-full border-collapse text-base">
		<thead>
			<tr class="border-b border-white/20 text-gray-400">
				<th class="w-[22%] px-3 py-2 text-center font-normal">Location</th>
				<th class="w-[12%] px-3 py-2 text-center font-normal">%</th>
				<th class="w-[12%] px-3 py-2 text-center font-normal">Dist.</th>
				<th class="w-[14%] px-3 py-2 text-center font-normal">Nether</th>
				<th class="w-[18%] px-3 py-2 text-center font-normal">Angle</th>
			</tr>
		</thead>
		<tbody>
			{#each predictions as pred (pred.chunkX + '-' + pred.chunkZ)}
				{@const coords = getPredictionDisplayCoords(pred.chunkX, pred.chunkZ)}
				{@const nether = chunkToNether(pred.chunkX, pred.chunkZ)}
				{@const travel = getTravelAngleInfo(pred.chunkX, pred.chunkZ)}
				{@const certaintyCol = getCertaintyColor(pred.certainty)}
				<tr class="border-b border-white/5 transition-colors hover:bg-white/5">
					<td class="px-3 py-1.5 text-center">
						(<span
							class={settings.colorNegatives && coords[0] < 0
								? 'text-red-400'
								: 'text-gray-200'}>{coords[0]}</span
						>,
						<span
							class={settings.colorNegatives && coords[1] < 0
								? 'text-red-400'
								: 'text-gray-200'}>{coords[1]}</span
						>)
					</td>
					<td class="px-3 py-1.5 text-center font-semibold" style="color: {certaintyCol}">
						{(pred.certainty * 100).toFixed(1)}%
					</td>
					<td class="px-3 py-1.5 text-center text-gray-200">
						{pos?.isInNether
							? Math.floor(pred.overworldDistance / 8)
							: Math.floor(pred.overworldDistance)}
					</td>
					<td class="px-3 py-1.5 text-center">
						(<span
							class={settings.colorNegatives && nether[0] < 0
								? 'text-red-400'
								: 'text-gray-300'}>{nether[0]}</span
						>,
						<span
							class={settings.colorNegatives && nether[1] < 0
								? 'text-red-400'
								: 'text-gray-300'}>{nether[1]}</span
						>)
					</td>
					<td class="whitespace-nowrap px-3 py-1.5 text-center">
						{#if travel.angle !== null && travel.diff !== null}
							<span class="text-gray-200"
								>{normalizeMinecraftAngle(travel.angle).toFixed(2)}</span
							>
							<span style="color: {getAngleDiffColor(travel.diff)}" class="ml-1.5"
								>({travel.diff > 0 ? '→' : '←'} {Math.abs(travel.diff).toFixed(1)})</span
							>
						{:else}
							<span class="text-gray-500">-</span>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else if resultType !== 'FAILED'}
	<div class="flex h-16 items-center justify-center text-muted-foreground">
		<svg
			class="mr-2 h-4 w-4"
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<circle cx="12" cy="12" r="10" />
			<line x1="12" y1="8" x2="12" y2="12" />
			<line x1="12" y1="16" x2="12.01" y2="16" />
		</svg>
		No predictions available.
	</div>
{/if}
