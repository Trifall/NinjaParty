<script lang="ts">
	import type { EyeThrow } from '$lib/types/stronghold';
	import { getCorrectionColor, normalizeMinecraftAngle } from '$lib/utils/stronghold';
	import {
		BOAT_ALLOWABLE_ERROR,
		BOAT_STD_DEV,
		NON_BOAT_WARNING_THRESHOLD,
		getDerivedIncrements,
	} from '$src/lib/utils/constants';

	interface Props {
		throws: EyeThrow[];
	}

	let { throws }: Props = $props();

	function getErrorStatus(eyeThrow: { error: number; type: string }): {
		color: string;
		warning: string | null;
	} {
		const absError = Math.abs(eyeThrow.error);
		const isBoat = eyeThrow.type === 'BOAT';

		if (isBoat) {
			if (absError > BOAT_ALLOWABLE_ERROR)
				return { color: 'text-red-400', warning: 'Boat angle error exceeds 0.03' };
			if (absError > BOAT_STD_DEV)
				return { color: 'text-amber-400', warning: 'Error exceeds boat std dev (0.0007)' };
			return { color: 'text-emerald-400', warning: null };
		}

		if (absError > NON_BOAT_WARNING_THRESHOLD)
			return { color: 'text-amber-400', warning: 'Possible bad angle (>0.05)' };
		if (absError > 0.01) return { color: 'text-yellow-400', warning: null };
		return { color: 'text-emerald-400', warning: null };
	}
</script>

{#if throws.length > 0}
	<div class="border-t border-white/20">
		<div class="flex items-center justify-between bg-white/[0.02] px-4 py-2">
			<span class="font-semibold tracking-wide text-gray-300">Ender eye throws</span>
			<span class="text-sm font-semibold text-gray-300">{throws.length} total</span>
		</div>
		<table class="w-full border-collapse text-base">
			<thead>
				<tr class="border-b border-white/20 text-gray-400">
					<th class="px-4 py-1.5 text-center font-normal">X</th>
					<th class="px-4 py-1.5 text-center font-normal">Z</th>
					<th class="px-4 py-1.5 text-center font-normal">Angle</th>
					<th class="px-4 py-1.5 text-center font-normal">Corr</th>
					<th class="px-4 py-1.5 text-center font-normal">
						<span class="inline-flex items-center gap-1">
							Adj
							<span class="group relative cursor-help">
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
								<span
									class="absolute left-1/2 top-full z-50 mt-1 hidden w-52 -translate-x-1/2 rounded border border-white/10 bg-black/95 px-3 py-2 text-center text-xs text-white shadow-lg group-hover:block"
								>
									Assumes 30 FOV, TALL resolution, 16384px height. May be slightly off.
								</span>
							</span>
						</span>
					</th>
					<th class="px-4 py-1.5 text-center font-normal">Error</th>
					<th class="px-4 py-1.5 text-center font-normal">Type</th>
				</tr>
			</thead>
			<tbody>
				{#each throws as t (t.xInOverworld + '-' + t.zInOverworld + '-' + t.angle)}
					{@const status = getErrorStatus(t)}
					{@const incr = getDerivedIncrements(t.correction)}
					<tr class="border-b border-white/5 transition-colors hover:bg-white/5">
						<td class="px-4 py-1.5 text-center text-gray-200">{t.xInOverworld.toFixed(2)}</td>
						<td class="px-4 py-1.5 text-center text-gray-200">{t.zInOverworld.toFixed(2)}</td>
						<td class="px-4 py-1.5 text-center text-gray-200"
							>{normalizeMinecraftAngle(t.angle).toFixed(2)}°</td
						>
						<td class="px-4 py-1.5 text-center">
							{#if t.correction !== undefined && t.correction !== 0}
								<span style="color: {getCorrectionColor(t.correction)}"
									>{t.correction > 0 ? '+' : ''}{t.correction.toFixed(4)}</span
								>
							{:else}
								<span class="text-gray-500">0.0000</span>
							{/if}
						</td>
						<td class="px-4 py-1.5 text-center">
							{#if incr !== 0}
								<span style="color: {getCorrectionColor(incr)}">{incr > 0 ? '+' : ''}{incr}</span>
							{:else}
								<span class="text-gray-500">-</span>
							{/if}
						</td>
						<td class="px-4 py-1.5 text-center">
							<span class="inline-flex items-center gap-1">
								<span class={status.color}>{t.error.toFixed(5)}</span>
								{#if status.warning}
									<span class="group relative">
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
										<span
											class="absolute right-0 top-full z-50 mt-1 hidden whitespace-nowrap rounded border border-white/10 bg-black/95 px-2 py-1 text-xs text-white shadow-lg group-hover:block"
										>
											{status.warning}
										</span>
									</span>
								{/if}
							</span>
						</td>
						<td class="px-4 py-1.5 text-center">
							<span class="text-xs uppercase text-gray-400">{t.type}</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
