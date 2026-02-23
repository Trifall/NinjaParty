<script lang="ts">
	import type { EyeThrow } from '$lib/types/stronghold';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card';
	import { Eye, TriangleAlert, Info } from '@lucide/svelte';
	import { normalizeMinecraftAngle, getCorrectionColor } from '$lib/utils/stronghold';

	let { throws = [] }: { throws?: EyeThrow[] } = $props();

	const BOAT_ALLOWABLE_ERROR = 0.03;
	const BOAT_STD_DEV = 0.0007;
	const NON_BOAT_WARNING_THRESHOLD = 0.05;

	function getDerivedIncrements(correction: number | undefined): number {
		if (correction === undefined) return 0;
		return Math.round(correction * 451.9655419231479);
	}

	function getErrorStatus(eyeThrow: EyeThrow): { color: string; warning: string | null } {
		const absError = Math.abs(eyeThrow.error);
		const isBoat = eyeThrow.type === 'BOAT';

		if (isBoat) {
			if (absError > BOAT_ALLOWABLE_ERROR) {
				return { color: 'text-red-400', warning: 'Boat angle error exceeds 0.03' };
			}
			if (absError > BOAT_STD_DEV) {
				return { color: 'text-amber-400', warning: 'Error exceeds boat std dev (0.0007)' };
			}
			return { color: 'text-emerald-400', warning: null };
		}

		if (absError > NON_BOAT_WARNING_THRESHOLD) {
			return { color: 'text-amber-400', warning: 'Possible bad angle (>0.05)' };
		}
		if (absError > 0.01) {
			return { color: 'text-yellow-400', warning: null };
		}
		return { color: 'text-emerald-400', warning: null };
	}

	let warnings = $derived(throws.map(t => getErrorStatus(t)));
	let increments = $derived(throws.map(t => getDerivedIncrements(t.correction)));
</script>

<Card class="border-white/10 bg-black/40 backdrop-blur-xl">
	<CardHeader>
		<CardTitle class="flex items-center gap-2 text-xl">
			<Eye class="h-5 w-5 text-indigo-400" />
			Eye Throws
		</CardTitle>
		<CardDescription>
			Total measured throws: {throws.length}
		</CardDescription>
	</CardHeader>
	<CardContent>
		<div class="rounded-md border border-white/10 bg-black/20">
			<Table>
				<TableHeader>
					<TableRow class="hover:bg-white/5 data-[state=selected]:bg-muted">
						<TableHead class="w-[100px]">X</TableHead>
						<TableHead>Z</TableHead>
						<TableHead>Angle</TableHead>
						<TableHead>Correction</TableHead>
						<TableHead class="text-center">
							<span class="inline-flex items-center gap-1">
								Adj
								<span class="relative group cursor-help">
									<Info class="h-3.5 w-3.5 text-muted-foreground" />
									<span class="absolute top-full left-1/2 -translate-x-1/2 mt-2 hidden group-hover:block w-52 rounded bg-black/95 border border-white/10 px-2 py-1.5 text-xs text-white shadow-lg z-50 text-center">
										Assumes 30 FOV, TALL resolution, 16384px height. May be slightly off otherwise.
									</span>
								</span>
							</span>
						</TableHead>
						<TableHead>Error</TableHead>
						<TableHead class="text-right">Type</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{#if throws.length === 0}
						<TableRow class="hover:bg-transparent">
							<TableCell colspan={7} class="h-24 text-center text-muted-foreground">
								No eye throws recorded yet.
							</TableCell>
						</TableRow>
					{:else}
						{#each throws as eyeThrow, i (i)}
							{@const status = warnings[i]}
							{@const derivedIncr = increments[i]}
							<TableRow class="hover:bg-white/5 data-[state=selected]:bg-muted">
								<TableCell class="font-medium text-white">{eyeThrow.xInOverworld.toFixed(2)}</TableCell>
								<TableCell class="text-white">{eyeThrow.zInOverworld.toFixed(2)}</TableCell>
								<TableCell class="text-white">{normalizeMinecraftAngle(eyeThrow.angle).toFixed(2)}&deg;</TableCell>
								<TableCell>
									{#if eyeThrow.correction !== undefined && eyeThrow.correction !== 0}
										<span style="color: {getCorrectionColor(eyeThrow.correction)}" class="font-medium">
											{eyeThrow.correction > 0 ? '+' : ''}{eyeThrow.correction.toFixed(4)}
										</span>
									{:else}
										<span class="text-muted-foreground">0.0000</span>
									{/if}
								</TableCell>
								<TableCell class="text-center">
									{#if derivedIncr !== 0}
										<span style="color: {getCorrectionColor(derivedIncr)}" class="font-mono text-xs font-medium">
											{derivedIncr > 0 ? '+' : ''}{derivedIncr}
										</span>
									{:else}
										<span class="text-muted-foreground/50">-</span>
									{/if}
								</TableCell>
								<TableCell>
									<div class="flex items-center gap-1.5">
										{#if eyeThrow.error !== 0}
											<span class={status.color}>
												{eyeThrow.error.toFixed(5)}
											</span>
										{:else}
											<span class="text-muted-foreground">0.00000</span>
										{/if}
										{#if status.warning}
											<span class="relative group">
												<TriangleAlert class="h-4 w-4 text-amber-400" />
												<span class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block whitespace-nowrap rounded bg-black/90 px-2 py-1 text-xs text-white shadow-lg z-10">
													{status.warning}
												</span>
											</span>
										{/if}
									</div>
								</TableCell>
								<TableCell class="text-right text-muted-foreground">
									<span class="inline-flex rounded-full bg-white/10 px-2 py-0.5 text-xs font-semibold uppercase">
										{eyeThrow.type}
									</span>
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</div>
	</CardContent>
</Card>
