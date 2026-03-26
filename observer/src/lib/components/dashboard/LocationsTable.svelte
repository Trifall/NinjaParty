<script lang="ts">
	import {
		type CoordinateDisplayType,
		VIEW_MODE,
		type ViewMode,
		settings,
	} from '$lib/stores/settings.svelte';
	import type {
		ChunkPrediction,
		EyeThrow,
		PlayerPosition,
		ResultType,
	} from '$lib/types/stronghold';
	import {
		chunkToNether,
		getAngleDiffColor,
		getAngleDifference,
		getAngleToStructure,
		getCertaintyColor,
		normalizeMinecraftAngle,
	} from '$lib/utils/stronghold';
	import {
		BOAT_ALLOWABLE_ERROR,
		BOAT_STD_DEV,
		NON_BOAT_WARNING_THRESHOLD,
	} from '$src/lib/utils/constants';
	import { Tooltip, TooltipContent, TooltipTrigger } from '$components/ui/tooltip';

	interface Props {
		resultType: ResultType;
		predictions: ChunkPrediction[];
		pos: PlayerPosition | undefined;
		eyeThrows?: EyeThrow[];
		viewMode?: ViewMode;
		// optional overrides — provided by Dashboard in embed mode
		// falls back to the settings store when not set (normal mode)
		displayType?: CoordinateDisplayType;
		colorNegatives?: boolean;
	}

	let {
		resultType,
		predictions,
		pos,
		eyeThrows = [],
		viewMode = VIEW_MODE.MAX,
		displayType: displayTypeProp,
		colorNegatives: colorNegativesProp,
	}: Props = $props();

	let isCompact = $derived(viewMode === VIEW_MODE.COMPACT || viewMode === VIEW_MODE.ULTRA_COMPACT);
	let isUltraCompact = $derived(viewMode === VIEW_MODE.ULTRA_COMPACT);

	// use prop if provided, otherwise fall back to store
	let displayType = $derived(displayTypeProp ?? settings.displayType.current);
	let colorNegatives = $derived(colorNegativesProp ?? settings.colorNegatives.current);

	function getPredictionDisplayCoords(chunkX: number, chunkZ: number): [number, number] {
		if (displayType === 'FOUR_FOUR') return [16 * chunkX + 4, 16 * chunkZ + 4];
		if (displayType === 'EIGHT_EIGHT') return [16 * chunkX + 8, 16 * chunkZ + 8];
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

	function getMaxErrorInfo() {
		if (eyeThrows.length === 0) return { error: null, warning: null, color: null };

		let maxError = 0;
		let maxErrorType = '';
		for (const t of eyeThrows) {
			if (Math.abs(t.error) > Math.abs(maxError)) {
				maxError = t.error;
				maxErrorType = t.type;
			}
		}

		const absError = Math.abs(maxError);
		const isBoat = maxErrorType === 'BOAT';
		let warning = null;
		let color = 'text-gray-200';

		if (isBoat) {
			if (absError > BOAT_ALLOWABLE_ERROR) {
				warning = 'Boat angle error exceeds 0.03';
				color = 'text-red-400';
			} else if (absError > BOAT_STD_DEV) {
				warning = 'Error exceeds boat std dev (0.0007)';
				color = 'text-amber-400';
			} else {
				color = 'text-emerald-400';
			}
		} else {
			if (absError > NON_BOAT_WARNING_THRESHOLD) {
				warning = 'Possible bad angle (>0.05)';
				color = 'text-amber-400';
			} else if (absError > 0.01) {
				color = 'text-yellow-400';
			} else {
				color = 'text-emerald-400';
			}
		}

		return { error: maxError, warning, color };
	}

	let maxErrorInfo = $derived(getMaxErrorInfo());
</script>

{#if resultType === 'FAILED'}
	<div
		class="flex items-center justify-center gap-3 overflow-hidden rounded-lg border border-white/10 bg-red-500/5 px-4 py-3 text-center text-red-400"
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
	<div class={isUltraCompact ? '' : 'overflow-hidden rounded-lg border border-white/10 bg-white/[0.02]'}>
		<table class="w-full border-collapse text-base">
			<thead>
			<tr class="border-b border-white/20 text-gray-400 {isUltraCompact ? 'text-xs' : ''}">
				<th class="w-[22%] px-3 {isUltraCompact ? 'py-0' : 'py-2'} text-center font-normal"
					>Location</th
				>
				<th class="w-[12%] px-3 {isUltraCompact ? 'py-0' : 'py-2'} text-center font-normal">%</th>
				<th class="w-[12%] px-3 {isUltraCompact ? 'py-0' : 'py-2'} text-center font-normal"
					>Dist.</th
				>
				<th class="w-[14%] px-3 {isUltraCompact ? 'py-0' : 'py-2'} text-center font-normal"
					>Nether</th
				>
				<th class="w-[18%] px-3 {isUltraCompact ? 'py-0' : 'py-2'} text-center font-normal"
					>Angle</th
				>
				{#if isCompact && !isUltraCompact}
					<th class="w-[12%] px-3 py-2 text-center font-normal">
						<span class="inline-flex items-center gap-1">
							Error
							<Tooltip delayDuration={0}>
								<TooltipTrigger class="cursor-help flex items-center">
									<svg
										class="h-4 w-4 text-gray-500"
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
										stroke-linecap="round"
										stroke-linejoin="round"
									>
										<circle cx="12" cy="12" r="10" />
										<path d="M12 16v-4" />
										<path d="M12 8h.01" />
									</svg>
								</TooltipTrigger>
								<TooltipContent>
									<p>Displays the LARGEST error from the throws of this calculation.</p>
								</TooltipContent>
							</Tooltip>
						</span>
					</th>
				{/if}
			</tr>
		</thead>
		<tbody>
			{#each predictions as pred (pred.chunkX + '-' + pred.chunkZ)}
				{@const coords = getPredictionDisplayCoords(pred.chunkX, pred.chunkZ)}
				{@const nether = chunkToNether(pred.chunkX, pred.chunkZ)}
				{@const travel = getTravelAngleInfo(pred.chunkX, pred.chunkZ)}
				{@const certaintyCol = getCertaintyColor(pred.certainty)}
				<tr
					class="border-b border-white/5 transition-colors hover:bg-white/5 {isUltraCompact
						? 'text-xs'
						: ''}"
				>
					<td class="px-3 {isUltraCompact ? 'py-0.5' : 'py-1.5'} text-center">
						(<span class={colorNegatives && coords[0] < 0 ? 'text-red-400' : 'text-gray-200'}
							>{coords[0]}</span
						>,
						<span class={colorNegatives && coords[1] < 0 ? 'text-red-400' : 'text-gray-200'}
							>{coords[1]}</span
						>)
					</td>
					<td
						class="px-3 {isUltraCompact ? 'py-0.5' : 'py-1.5'} text-center font-semibold"
						style="color: {certaintyCol}"
					>
						{(pred.certainty * 100).toFixed(1)}%
					</td>
					<td class="px-3 {isUltraCompact ? 'py-0.5' : 'py-1.5'} text-center text-gray-200">
						{pos?.isInNether
							? Math.floor(pred.overworldDistance / 8)
							: Math.floor(pred.overworldDistance)}
					</td>
					<td class="px-3 {isUltraCompact ? 'py-0.5' : 'py-1.5'} text-center">
						(<span class={colorNegatives && nether[0] < 0 ? 'text-red-400' : 'text-gray-300'}
							>{nether[0]}</span
						>,
						<span class={colorNegatives && nether[1] < 0 ? 'text-red-400' : 'text-gray-300'}
							>{nether[1]}</span
						>)
					</td>
					<td class="whitespace-nowrap px-3 {isUltraCompact ? 'py-0.5' : 'py-1.5'} text-center">
						{#if travel.angle !== null && travel.diff !== null}
							<span class="text-gray-200">{normalizeMinecraftAngle(travel.angle).toFixed(2)}</span>
							{#if !isUltraCompact}
								<span style="color: {getAngleDiffColor(travel.diff)}" class="ml-1.5"
									>({travel.diff > 0 ? '→' : '←'} {Math.abs(travel.diff).toFixed(1)})</span
								>
							{/if}
						{:else}
							<span class="text-gray-500">-</span>
						{/if}
					</td>
					{#if isCompact && !isUltraCompact}
						<td class="whitespace-nowrap px-3 py-1.5 text-center">
							{#if maxErrorInfo.error !== null}
								<span class="inline-flex items-center gap-1">
									<span class={maxErrorInfo.color}>{maxErrorInfo.error.toFixed(5)}</span>
									{#if maxErrorInfo.warning}
										<Tooltip delayDuration={0}>
											<TooltipTrigger class="cursor-help flex items-center">
												<svg
													class="h-4 w-4 text-amber-400"
													xmlns="http://www.w3.org/2000/svg"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													stroke-width="2"
													stroke-linecap="round"
													stroke-linejoin="round"
												>
													<path
														d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"
													/>
													<line x1="12" y1="9" x2="12" y2="13" />
													<line x1="12" y1="17" x2="12.01" y2="17" />
												</svg>
											</TooltipTrigger>
											<TooltipContent>
												<p>{maxErrorInfo.warning}</p>
											</TooltipContent>
										</Tooltip>
									{/if}
								</span>
							{:else}
								<span class="text-gray-500">-</span>
							{/if}
						</td>
					{/if}
				</tr>
			{/each}
		</tbody>
	</table>
	</div>
{:else if resultType !== 'FAILED'}
	<div class="flex h-16 items-center justify-center text-muted-foreground {isUltraCompact ? '' : 'overflow-hidden rounded-lg border border-white/10 bg-white/[0.02]'}">
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
