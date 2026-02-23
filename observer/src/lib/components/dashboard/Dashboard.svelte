<script lang="ts">
	import {
		CircleAlert,
		Eye,
		EyeOff,
		Info,
		LoaderCircle,
		MapPin,
		TriangleAlert,
		Wifi,
		WifiOff,
	} from '@lucide/svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Badge } from '$lib/components/ui/badge';
	import type { RoomConnection } from '$lib/stores/room.svelte';
	import { settings } from '$lib/stores/settings.svelte';
	import {
		chunkToNether,
		getAngleDiffColor,
		getAngleDifference,
		getAngleToStructure,
		getCertaintyColor,
		getCorrectionColor,
		normalizeMinecraftAngle,
	} from '$lib/utils/stronghold';
	import SettingsSheet from './SettingsSheet.svelte';

	let { roomStore, roomCode }: { roomStore: RoomConnection; roomCode: string } = $props();

	let isRoomCodeHidden = $state(false);
	let showRevealConfirm = $state(false);

	let payload = $derived(roomStore.stronghold);
	let predictions = $derived(payload?.predictions?.slice(0, 5) || []);
	let pos = $derived(payload?.playerPosition);
	let throws = $derived(payload?.eyeThrows || []);

	let connectionColor = $derived(
		roomStore.status === 'connected'
			? 'bg-emerald-500 text-white border-emerald-500'
			: roomStore.status === 'reconnecting'
				? 'bg-amber-500 text-white border-amber-500'
				: roomStore.status === 'connecting'
					? 'bg-blue-500 text-white border-blue-500'
					: 'bg-red-500 text-white border-red-500'
	);

	let runnerColor = $derived(
		roomStore.runnerStatus === 'online'
			? 'text-emerald-500 border-emerald-500'
			: roomStore.runnerStatus === 'offline'
				? 'text-red-500 border-red-500'
				: 'text-gray-400 border-gray-400'
	);

	let boatColor = $derived(
		roomStore.boat?.boatState === 'VALID'
			? 'text-emerald-500 border-emerald-500'
			: roomStore.boat?.boatState === 'MEASURING'
				? 'text-blue-500 border-blue-500'
				: roomStore.boat?.boatState === 'ERROR'
					? 'text-red-500 border-red-500'
					: 'hidden'
	);

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

	const BOAT_ALLOWABLE_ERROR = 0.03;
	const BOAT_STD_DEV = 0.0007;
	const NON_BOAT_WARNING_THRESHOLD = 0.05; // guessing this

	function getDerivedIncrements(correction: number | undefined): number {
		/* this data is not available in the API for Ninjabrain v1.5.1 release, so we are estimating it

			formula uses Math.cos(beta * toRad)
				- verticalAngle = roughly -32.11166deg from my limited testing
				- cos(-32.11166deg) = cos(32.11166deg) = ~0.8466
					-> so correction per increment = 0.001874 / 0.8466 = ~0.002213 deg
					-> correctionIncrements = round(correction / 0.002213) = round(correction * 451.9)

			this is a rough approximation, we don't have access to the exact data
		 */
		if (correction === undefined) return 0;
		return Math.round(correction * 451.9);
	}

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

	let infoMessages = $derived(roomStore.informationMessages?.informationMessages || []);
</script>

<div class="mx-auto w-full max-w-3xl">
	<!-- Header -->
	<div
		class="flex items-center justify-between rounded-t-lg border border-b-0 border-white/20 bg-black/60 px-5 py-3 backdrop-blur-xl"
	>
		<div class="flex items-center gap-4">
			<span class="text-base text-gray-400">Room:</span>
			{#if isRoomCodeHidden}
				<span class="text-base font-bold uppercase tracking-wider text-emerald-400">******</span>
			{:else}
				<span class="text-base font-bold uppercase tracking-wider text-emerald-400">{roomCode}</span
				>
			{/if}
			<button
				type="button"
				onclick={() => {
					if (isRoomCodeHidden) {
						showRevealConfirm = true;
					} else {
						isRoomCodeHidden = true;
					}
				}}
				class="text-gray-400 transition-colors hover:text-gray-200"
				title={isRoomCodeHidden ? 'Reveal room code' : 'Hide room code'}
			>
				{#if isRoomCodeHidden}
					<EyeOff class="h-4 w-4" />
				{:else}
					<Eye class="h-4 w-4" />
				{/if}
			</button>
		</div>
		<div class="flex items-center gap-3">
			<Badge variant="outline" class="flex gap-2 px-3 py-1.5 text-sm {connectionColor}">
				{#if roomStore.status === 'connected'}
					<Wifi class="h-4 w-4" /> Live
				{:else if roomStore.status === 'reconnecting'}
					<LoaderCircle class="h-4 w-4 animate-spin" /> Reconnecting
				{:else if roomStore.status === 'connecting'}
					<LoaderCircle class="h-4 w-4 animate-spin" /> Connecting
				{:else}
					<WifiOff class="h-4 w-4" /> Offline
				{/if}
			</Badge>
			<Badge variant="outline" class="flex gap-2 px-3 py-1.5 text-sm {runnerColor}">
				<span class="relative flex h-2.5 w-2.5">
					{#if roomStore.runnerStatus === 'online'}
						<span
							class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
						></span>
						<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-emerald-500"></span>
					{:else}
						<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-red-500"></span>
					{/if}
				</span>
				Runner
			</Badge>
			{#if roomStore.boat && roomStore.boat.boatState !== 'NONE'}
				<Badge
					variant="outline"
					class="flex items-center gap-2 border-white/20 bg-black/40 px-3 py-1.5 text-sm {boatColor}"
				>
					{#if roomStore.boat.boatState === 'VALID'}
						<img
							src="/images/boat_green.png"
							alt="Valid"
							class="h-4 w-4 object-contain drop-shadow"
						/>
					{:else if roomStore.boat.boatState === 'MEASURING'}
						<img
							src="/images/boat_blue.png"
							alt="Measuring"
							class="h-4 w-4 object-contain drop-shadow"
						/>
					{:else if roomStore.boat.boatState === 'ERROR'}
						<img
							src="/images/boat_red.png"
							alt="Error"
							class="h-4 w-4 object-contain drop-shadow"
						/>
					{/if}
					{roomStore.boat.boatState}
				</Badge>
			{/if}
			<SettingsSheet />
		</div>
	</div>

	<!-- Main Panel -->
	<div
		class="rounded-b-lg border border-t-0 border-white/20 bg-[#1a1a1a] text-sm text-gray-200 backdrop-blur-xl"
	>
		{#if roomStore.status === 'connected' && !payload && !infoMessages.length}
			<div class="flex h-48 flex-col items-center justify-center gap-3 text-muted-foreground">
				<LoaderCircle class="h-8 w-8 animate-spin text-emerald-500" />
				<p>Waiting for Ninjabrain Bot data...</p>
			</div>
		{:else}
			<!-- Info Messages -->
			<!-- Currently disabled: this is not available in Ninjabrain v1.5.1 release -->

			<!-- {#each infoMessages as msg (msg.message + msg.severity)}
				<div class="flex items-start gap-3 px-4 py-3 {msg.severity === 'ERROR' ? 'bg-red-500/10 border-red-500/30 text-red-200' : msg.severity === 'WARNING' ? 'bg-amber-500/10 border-amber-500/30 text-amber-200' : 'bg-blue-500/10 border-blue-500/30 text-blue-200'} border-b border-white/20">
					{#if msg.severity === 'ERROR'}
						<TriangleAlert class="h-5 w-5 shrink-0 text-red-400" />
					{:else if msg.severity === 'WARNING'}
						<TriangleAlert class="h-5 w-5 shrink-0 text-amber-400" />
					{:else}
						<Info class="h-5 w-5 shrink-0 text-blue-400" />
					{/if}
					<span class="leading-snug">{@html msg.message}</span>
				</div>
			{/each}
			-->

			{#if payload}
				<!-- Failed State Banner -->
				{#if payload.resultType === 'FAILED'}
					<div
						class="flex items-center justify-center gap-3 border-b border-white/20 bg-red-500/5 px-4 py-3 text-center text-red-400"
					>
						<CircleAlert class="h-5 w-5 shrink-0" />
						<p>Could not determine stronghold chunk. You likely misread one of the eyes.</p>
					</div>
				{/if}

				<!-- Locations Table (only if not FAILED and has predictions) -->
				{#if payload.resultType !== 'FAILED' && predictions.length > 0}
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
				{:else if payload.resultType !== 'FAILED'}
					<div class="flex h-16 items-center justify-center text-muted-foreground">
						<CircleAlert class="mr-2 h-4 w-4" /> No predictions available.
					</div>
				{/if}

				<!-- Player Position (horizontal) - always show if available -->
				{#if pos}
					{@const displayX =
						pos.isInNether && pos.xInOverworld !== undefined
							? pos.xInOverworld / 8
							: pos.xInOverworld}
					{@const displayZ =
						pos.isInNether && pos.zInOverworld !== undefined
							? pos.zInOverworld / 8
							: pos.zInOverworld}
					<div
						class="flex items-center justify-between border-t border-white/20 bg-white/[0.02] px-4 py-2.5 text-base"
					>
						<div class="flex items-center gap-5">
							<span class="flex items-center gap-1.5 text-gray-400"
								><MapPin class="h-4 w-4" /> Position:</span
							>
							<span class="text-gray-200">
								(<span
									class={settings.colorNegatives && displayX !== undefined && displayX < 0
										? 'text-red-400'
										: ''}>{displayX?.toFixed(2) ?? '?'}</span
								>,
								<span
									class={settings.colorNegatives && displayZ !== undefined && displayZ < 0
										? 'text-red-400'
										: ''}>{displayZ?.toFixed(2) ?? '?'}</span
								>)
							</span>
						</div>
						<div class="flex items-center gap-5">
							<span class="text-gray-400">Angle:</span>
							<span class="text-gray-200"
								>{pos.horizontalAngle !== undefined
									? normalizeMinecraftAngle(pos.horizontalAngle).toFixed(2) + '°'
									: '-'}</span
							>
						</div>
						<div class="flex items-center gap-5">
							<span class="text-gray-400">Dimension:</span>
							<span class={pos.isInNether ? 'text-red-400' : 'text-emerald-400'}>
								{pos.isInNether ? 'Nether' : pos.isInOverworld ? 'Overworld' : '?'}
							</span>
						</div>
					</div>
				{/if}

				<!-- Eye Throws Section - always show if available -->
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
												<Info class="h-4 w-4 text-gray-500" />
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
										<td class="px-4 py-1.5 text-center text-gray-200"
											>{t.xInOverworld.toFixed(2)}</td
										>
										<td class="px-4 py-1.5 text-center text-gray-200"
											>{t.zInOverworld.toFixed(2)}</td
										>
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
												<span style="color: {getCorrectionColor(incr)}"
													>{incr > 0 ? '+' : ''}{incr}</span
												>
											{:else}
												<span class="text-gray-500">-</span>
											{/if}
										</td>
										<td class="px-4 py-1.5 text-center">
											<span class="inline-flex items-center gap-1">
												<span class={status.color}>{t.error.toFixed(5)}</span>
												{#if status.warning}
													<span class="group relative">
														<TriangleAlert class="h-4 w-4 text-amber-400" />
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

				<!-- Blind Mode -->
				{#if roomStore.blind?.isBlindModeEnabled && roomStore.blind.blindResult && 'xInNether' in roomStore.blind.blindResult}
					{@const blind = roomStore.blind.blindResult}
					<div class="border-t border-white/20 bg-purple-500/5 px-4 py-3">
						<div class="mb-1.5 flex items-center gap-2 font-semibold text-purple-300">
							<Eye class="h-4 w-4" /> Blind Mode
						</div>
						<div class="flex items-center gap-8 text-sm">
							<span class="text-gray-400"
								>Nether: <span class="font-semibold text-purple-300"
									>({blind.xInNether}, {blind.zInNether})</span
								></span
							>
							<span class="text-gray-400"
								>Avg: <span class="text-gray-200">{blind.averageDistance.toFixed(0)}</span></span
							>
							<span class="text-gray-400"
								>Highroll: <span class="text-gray-200"
									>{(blind.highrollProbability * 100).toFixed(1)}%</span
								></span
							>
							<span class="text-gray-400"
								>Eval: <span class="text-amber-300">{blind.evaluation.replace(/_/g, ' ')}</span
								></span
							>
						</div>
					</div>
				{/if}

				<!-- Divine Mode -->
				{#if roomStore.divine?.isDivineModeEnabled && roomStore.divine.divineResult && 'fossilXCoordinate' in roomStore.divine.divineResult}
					{@const divine = roomStore.divine.divineResult}
					<div class="border-t border-white/20 bg-amber-500/5 px-4 py-3">
						<div class="mb-1.5 flex items-center gap-2 font-semibold text-amber-300">
							<Eye class="h-4 w-4" /> Divine Mode
						</div>
						<div class="flex items-center gap-6 text-sm">
							<span class="text-gray-400"
								>Fossil X: <span class="font-semibold text-amber-300"
									>{divine.fossilXCoordinate}</span
								></span
							>
							<span class="text-gray-400"
								>Highroll: <span class="text-gray-200">{divine.formattedHighrollCoords}</span></span
							>
							<span class="text-gray-400"
								>Safe: <span class="text-gray-200">{divine.formattedSafeCoords}</span></span
							>
						</div>
					</div>
				{/if}

				<!-- All Advancements -->
				{#if roomStore.allAdvancements?.isAllAdvancementsModeEnabled}
					<div class="border-t border-white/20 bg-cyan-500/5 px-4 py-3">
						<div class="mb-1.5 flex items-center gap-2 font-semibold text-cyan-300">
							<MapPin class="h-4 w-4" /> All Advancements
						</div>
						<div class="flex flex-wrap gap-x-6 gap-y-1 text-sm">
							{#each [{ label: 'Spawn', data: roomStore.allAdvancements.spawn }, { label: 'Outpost', data: roomStore.allAdvancements.outpost }, { label: 'Monument', data: roomStore.allAdvancements.monument }, { label: 'Stronghold', data: roomStore.allAdvancements.stronghold }, { label: 'Deep Dark', data: roomStore.allAdvancements.deepDark }] as item (item.label)}
								<span class="text-gray-400"
									>{item.label}:
									{#if item.data && 'xInOverworld' in item.data}
										<span class="text-cyan-300"
											>({item.data.xInOverworld?.toFixed(0) ?? '?'}, {item.data.zInOverworld?.toFixed(
												0
											) ?? '?'})</span
										>
									{:else}
										<span class="text-gray-500">-</span>
									{/if}
								</span>
							{/each}
						</div>
					</div>
				{/if}
			{/if}
		{/if}
	</div>

	<AlertDialog.Root
		open={showRevealConfirm}
		onOpenChange={(open) => {
			showRevealConfirm = open;
		}}
	>
		<AlertDialog.Portal>
			<AlertDialog.Overlay
				class="fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
			/>
			<AlertDialog.Content
				class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg"
			>
				<AlertDialog.Header>
					<AlertDialog.Title>Reveal Room Code?</AlertDialog.Title>
					<AlertDialog.Description>
						Are you sure you want to reveal the room code? This may be visible if you're streaming.
					</AlertDialog.Description>
				</AlertDialog.Header>
				<AlertDialog.Footer>
					<AlertDialog.Cancel onclick={() => (showRevealConfirm = false)}>Cancel</AlertDialog.Cancel
					>
					<AlertDialog.Action
						onclick={() => {
							isRoomCodeHidden = false;
							showRevealConfirm = false;
						}}>Reveal</AlertDialog.Action
					>
				</AlertDialog.Footer>
			</AlertDialog.Content>
		</AlertDialog.Portal>
	</AlertDialog.Root>
</div>
