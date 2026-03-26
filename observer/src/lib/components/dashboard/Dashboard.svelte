<script lang="ts">
	import { Clock, Eye, EyeOff, House, LoaderCircle, Wifi, WifiOff } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Badge } from '$lib/components/ui/badge';
	import type { RoomConnection } from '$lib/stores/room.svelte';
	import { type ViewMode, settings } from '$lib/stores/settings.svelte';
	import { formatUptime } from '$src/lib/utils/format';
	import { embedParamsSchema } from '$components/dashboard/embed-params';
	import AnimatedGradientText from '$components/ui/AnimatedGradientText.svelte';
	import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '$components/ui/tooltip';
	import AllAdvancements from './AllAdvancements.svelte';
	import BlindMode from './BlindMode.svelte';
	import DivineMode from './DivineMode.svelte';
	import EyeThrows from './EyeThrows.svelte';
	import InformationMessages from './InformationMessages.svelte';
	import LocationsTable from './LocationsTable.svelte';
	import PlayerPosition from './PlayerPosition.svelte';
	import AddToObs from './modals/AddToObs.svelte';
	import SettingsSheet from './modals/SettingsSheet.svelte';

	let {
		roomStore,
		roomCode,
		isEmbed,
	}: { roomStore: RoomConnection; roomCode: string; isEmbed: boolean } = $props();

	let isRoomCodeHidden = $state(true);
	let showRevealConfirm = $state(false);

	let runners = $derived(roomStore.sortedRunners);

	let now = $state(Date.now());

	let embedParams = $derived(() => {
		const sp = $page.url.searchParams;
		const result = embedParamsSchema.safeParse({
			viewMode: sp.get('viewMode') ?? undefined,
			displayType: sp.get('displayType') ?? undefined,
			colorNegatives: sp.get('colorNegatives') ?? undefined,
			autoHide: sp.get('autoHide') ?? undefined,
		});
		return result.success ? result.data : embedParamsSchema.parse({});
	});

	// in embed mode, URL params are the source of truth
	// in normal mode, use the persisted settings store
	let viewMode = $derived(
		isEmbed ? (embedParams().viewMode as ViewMode) : (settings.viewMode.current as ViewMode)
	);
	let displayType = $derived(isEmbed ? embedParams().displayType : settings.displayType.current);
	let colorNegatives = $derived(
		isEmbed ? embedParams().colorNegatives : settings.colorNegatives.current
	);
	let autoHideSeconds = $derived(isEmbed ? embedParams().autoHide : settings.embedAutoHide.current);

	// auto-hide state
	let isHidden = $state(false);
	let autoHideTimer: ReturnType<typeof setTimeout> | null = null;

	// snapshot of only meaningful runner fields
	let runnerDataSnapshot = $derived(
		JSON.stringify(
			runners.map((r) => ({
				name: r.name,
				status: r.status,
				stronghold: r.stronghold,
				boat: r.boat,
				blind: r.blind,
				divine: r.divine,
				allAdvancements: r.allAdvancements,
			}))
		)
	);

	$effect(() => {
		console.log('Dashboard render:', {
			status: roomStore.status,
			runnersLength: runners.length,
			runners: runners.map((r) => r.name),
		});
	});

	$effect(() => {
		const interval = setInterval(() => {
			now = Date.now();
		}, 1000);
		return () => clearInterval(interval);
	});

	$effect(() => {
		// oxlint-disable-next-line no-unused-expressions
		runnerDataSnapshot;

		if (autoHideSeconds === 0) {
			if (autoHideTimer) clearTimeout(autoHideTimer);
			autoHideTimer = null;
			isHidden = false;
			return;
		}

		isHidden = false;
		if (autoHideTimer) clearTimeout(autoHideTimer);
		autoHideTimer = setTimeout(() => {
			isHidden = true;
		}, autoHideSeconds * 1000);

		return () => {
			if (autoHideTimer) {
				clearTimeout(autoHideTimer);
				autoHideTimer = null;
			}
		};
	});

	let connectionColor = $derived(
		roomStore.status === 'connected'
			? 'bg-emerald-500 text-white border-emerald-500'
			: roomStore.status === 'reconnecting'
				? 'bg-amber-500 text-white border-amber-500'
				: roomStore.status === 'connecting'
					? 'bg-blue-500 text-white border-blue-500'
					: 'bg-red-500 text-white border-red-500'
	);

	function getBoatColor(boatState: string | undefined): string {
		if (!boatState) return 'hidden';
		switch (boatState) {
			case 'VALID':
				return 'text-emerald-500 border-emerald-500';
			case 'MEASURING':
				return 'text-blue-500 border-blue-500';
			case 'ERROR':
				return 'text-red-500 border-red-500';
			default:
				return 'hidden';
		}
	}
</script>

<TooltipProvider>
	<div
		class="mx-auto w-full max-w-3xl transition-opacity duration-700 {isEmbed && isHidden
			? 'opacity-0'
			: 'opacity-100'}"
	>
		<!-- Header -->
		<div
			class="flex items-center justify-between rounded-t-lg border border-b-0 border-white/20 bg-black backdrop-blur-xl {viewMode ===
			'ultra-compact'
				? 'px-2 pb-0.5 pt-1'
				: 'px-5 py-3'}"
		>
			<div class="flex items-center gap-2">
				<span
					class="flex items-center gap-1 {viewMode === 'ultra-compact'
						? 'text-xs'
						: 'text-sm text-gray-800 dark:text-gray-200 md:text-lg'}"
				>
					<AnimatedGradientText
						noBorder
						className="p-0"
						textClassName="{viewMode === 'ultra-compact'
							? 'text-xs'
							: 'text-sm'} font-bold mb-0 md:text-lg pb-0"
					>
						NinjaParty
					</AnimatedGradientText>
				</span>

				{#if !isEmbed}
					<span class="text-gray-400">|</span>
					<span class="{viewMode === 'ultra-compact' ? 'text-xs' : 'text-base'} text-gray-400"
						>Room:</span
					>
					{#if isRoomCodeHidden}
						<span
							class="{viewMode === 'ultra-compact'
								? 'text-xs'
								: 'text-base'} font-bold uppercase tracking-wider text-emerald-400">******</span
						>
					{:else}
						<span
							class="{viewMode === 'ultra-compact'
								? 'text-xs'
								: 'text-base'} font-bold uppercase tracking-wider text-emerald-400">{roomCode}</span
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
							<EyeOff class={viewMode === 'ultra-compact' ? 'mt-0.5 h-3 w-3' : 'mt-1 h-4 w-4'} />
						{:else}
							<Eye class={viewMode === 'ultra-compact' ? 'mt-0.5 h-3 w-3' : 'mt-1 h-4 w-4'} />
						{/if}
					</button>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				<Badge
					variant="outline"
					class="flex gap-2 {viewMode === 'ultra-compact'
						? 'px-1.5 py-0.5 text-xs'
						: 'px-3 py-1.5 text-sm'} {connectionColor}"
				>
					{#if roomStore.status === 'connected'}
						<Wifi class={viewMode === 'ultra-compact' ? 'h-3 w-3' : 'h-4 w-4'} /> Live
					{:else if roomStore.status === 'reconnecting'}
						<LoaderCircle
							class="{viewMode === 'ultra-compact' ? 'h-3 w-3' : 'h-4 w-4'} animate-spin"
						/> Reconnecting
					{:else if roomStore.status === 'connecting'}
						<LoaderCircle
							class="{viewMode === 'ultra-compact' ? 'h-3 w-3' : 'h-4 w-4'} animate-spin"
						/> Connecting
					{:else}
						<WifiOff class={viewMode === 'ultra-compact' ? 'h-3 w-3' : 'h-4 w-4'} /> Offline
					{/if}
				</Badge>
				<Badge
					variant="outline"
					class="flex gap-2 border-white/20 bg-black/40 {viewMode === 'ultra-compact'
						? 'px-1.5 py-0.5 text-xs'
						: 'px-3 py-1.5 text-sm'}"
				>
					<span class="text-gray-400">{runners.length} Runner{runners.length !== 1 ? 's' : ''}</span
					>
				</Badge>
				{#if viewMode !== 'ultra-compact'}
					<Badge
						variant="outline"
						class="flex gap-2 border-white/20 bg-black/40 px-3 py-1.5 text-sm"
					>
						<Tooltip delayDuration={0}>
							<TooltipTrigger class="flex items-center gap-2">
								<Clock class="h-4 w-4" />
								<span class="text-gray-400"> {formatUptime(roomStore.createdAt, now)}</span>
							</TooltipTrigger>
							<TooltipContent side="top">Room Uptime</TooltipContent>
						</Tooltip>
					</Badge>
				{/if}
				{#if !isEmbed}
					<SettingsSheet compact={viewMode === 'ultra-compact'} />
				{/if}
			</div>
		</div>

		<!-- Main Panel - Multi Runner Scrollable List -->
		<div
			class="rounded-b-lg border border-t-0 border-white/20 bg-[#1a1a1a] text-sm text-gray-200 backdrop-blur-xl"
		>
			{#if roomStore.roomClosed}
				<div class="flex h-48 flex-col items-center justify-center gap-3 text-red-400">
					<p class="text-base">This room has been closed</p>
				</div>
			{:else if roomStore.status === 'connected' && runners.length === 0}
				<div class="flex h-48 flex-col items-center justify-center gap-3 text-muted-foreground">
					<LoaderCircle class="h-8 w-8 animate-spin text-emerald-500" />
					<p>Waiting for runner data...</p>
				</div>
			{:else}
				<div
					class="flex flex-col {viewMode === 'ultra-compact' ? 'gap-1.5 p-2 py-0.5' : 'gap-4 p-4'}"
				>
					{#each runners as runner (runner.name)}
						{#if viewMode === 'ultra-compact'}
							<!-- Ultra-Compact: no card header, name inline with data -->
							<div
								class="rounded-md border border-white/10 bg-black/40 {runner.status === 'offline'
									? 'opacity-60'
									: ''}"
							>
								{#if !runner.stronghold}
									<div class="flex items-center gap-3 px-3 py-1.5 text-xs">
										<div class="flex items-center gap-1.5">
											<span class="relative flex h-2 w-2">
												{#if runner.status === 'online'}
													<span
														class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
													></span>
												{/if}
												<span
													class="relative inline-flex h-2 w-2 rounded-full {runner.status ===
													'online'
														? 'bg-emerald-500'
														: 'bg-red-500'}"
												></span>
											</span>
											<span class="font-semibold text-white">{runner.name}</span>
											{#if runner.status === 'offline'}
												<span class="text-xs text-red-400">(Offline)</span>
											{/if}
										</div>
										{#if runner.boat && runner.boat.boatState !== 'NONE'}
											<img
												src="/images/boat_{runner.boat.boatState === 'VALID'
													? 'green'
													: runner.boat.boatState === 'MEASURING'
														? 'blue'
														: 'red'}.png"
												alt={runner.boat.boatState}
												class="h-3 w-3 object-contain"
											/>
										{/if}
										<span class="ml-auto text-muted-foreground">Waiting for data...</span>
									</div>
								{:else}
									<div class="flex items-center gap-2 border-b border-white/5 px-3 py-1">
										<span class="relative flex h-2 w-2 shrink-0">
											{#if runner.status === 'online'}
												<span
													class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
												></span>
											{/if}
											<span
												class="relative inline-flex h-2 w-2 rounded-full {runner.status === 'online'
													? 'bg-emerald-500'
													: 'bg-red-500'}"
											></span>
										</span>
										<span class="text-sm font-semibold text-white">{runner.name}</span>
										{#if runner.status === 'offline'}
											<span class="text-xs text-red-400">(Offline)</span>
										{/if}
										<div class="ml-auto flex items-center gap-2">
											<PlayerPosition
												pos={runner.stronghold.playerPosition}
												inline={true}
												{colorNegatives}
											/>
											{#if runner.boat && runner.boat.boatState !== 'NONE'}
												<img
													src="/images/boat_{runner.boat.boatState === 'VALID'
														? 'green'
														: runner.boat.boatState === 'MEASURING'
															? 'blue'
															: 'red'}.png"
													alt={runner.boat.boatState}
													class="h-3 w-3 object-contain"
												/>
											{/if}
										</div>
									</div>
									<LocationsTable
										resultType={runner.stronghold.resultType}
										predictions={runner.stronghold.predictions?.slice(0, 1) || []}
										pos={runner.stronghold.playerPosition}
										eyeThrows={runner.stronghold.eyeThrows || []}
										{viewMode}
										{displayType}
										{colorNegatives}
									/>
								{/if}
								<InformationMessages
									messages={runner.informationMessages?.informationMessages}
									{viewMode}
								/>
							</div>
						{:else}
							<!-- Standard card layout (max / compact) -->
							<div
								class="rounded-lg border border-white/10 bg-black/40 p-4 {runner.status ===
								'offline'
									? 'opacity-60'
									: ''}"
							>
								<!-- Runner Header -->
								<div class="mb-1 flex items-center justify-between">
									<div class="flex items-center gap-2">
										<span class="relative flex h-3 w-3">
											{#if runner.status === 'online'}
												<span
													class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
												></span>
											{/if}
											<span
												class="relative inline-flex h-3 w-3 rounded-full {runner.status === 'online'
													? 'bg-emerald-500'
													: 'bg-red-500'}"
											></span>
										</span>
										<span class="text-lg font-semibold text-white">{runner.name}</span>
										{#if runner.status === 'offline'}
											<span class="text-xs text-red-400">(Offline)</span>
										{/if}
									</div>
									{#if runner.boat && runner.boat.boatState !== 'NONE'}
										<Badge
											variant="outline"
											class="flex items-center gap-2 border-white/20 bg-black/40 px-2 py-1 text-xs {getBoatColor(
												runner.boat.boatState
											)}"
										>
											{#if runner.boat.boatState === 'VALID'}
												<img
													src="/images/boat_green.png"
													alt="Valid"
													class="h-3 w-3 object-contain"
												/>
											{:else if runner.boat.boatState === 'MEASURING'}
												<img
													src="/images/boat_blue.png"
													alt="Measuring"
													class="h-3 w-3 object-contain"
												/>
											{:else if runner.boat.boatState === 'ERROR'}
												<img
													src="/images/boat_red.png"
													alt="Error"
													class="h-3 w-3 object-contain"
												/>
											{/if}
											{runner.boat.boatState}
										</Badge>
									{/if}
								</div>

								<!-- Runner Data -->
								<div class="flex flex-col gap-1">
									{#if !runner.stronghold}
										<div class="flex h-20 items-center justify-center text-muted-foreground">
											<p>Waiting for stronghold data...</p>
										</div>
									{:else if viewMode === 'compact'}
										<LocationsTable
											resultType={runner.stronghold.resultType}
											predictions={runner.stronghold.predictions?.slice(0, 1) || []}
											pos={runner.stronghold.playerPosition}
											eyeThrows={runner.stronghold.eyeThrows || []}
											{viewMode}
											{displayType}
											{colorNegatives}
										/>
										<PlayerPosition pos={runner.stronghold.playerPosition} {colorNegatives} />
									{:else}
										<LocationsTable
											resultType={runner.stronghold.resultType}
											predictions={runner.stronghold.predictions?.slice(0, 5) || []}
											pos={runner.stronghold.playerPosition}
											{viewMode}
											{displayType}
											{colorNegatives}
										/>
										<PlayerPosition pos={runner.stronghold.playerPosition} {colorNegatives} />
										<EyeThrows throws={runner.stronghold.eyeThrows || []} />
										<BlindMode blind={runner.blind ?? undefined} />
										<DivineMode divine={runner.divine ?? undefined} />
										<AllAdvancements allAdvancements={runner.allAdvancements ?? undefined} />
									{/if}
									<InformationMessages
										messages={runner.informationMessages?.informationMessages}
										{viewMode}
									/>
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</div>

		<!-- Add to OBS Button -->
		{#if !isEmbed}
			<div class="flex w-full items-center justify-center gap-2">
				<AddToObs {roomCode} />
				<div class="mt-4 flex justify-center">
					<button type="button" onclick={() => goto('/')} class="group relative rounded-lg p-[2px]">
						<div
							class="animate-gradient-rotate absolute inset-0 rounded-lg bg-gradient-to-r from-emerald-500 via-cyan-500 to-emerald-500"
						></div>
						<div
							class="relative flex items-center gap-2 rounded-lg border-0 bg-black px-4 py-2 text-sm font-medium text-white transition-all hover:bg-black/60 hover:text-white"
						>
							<House class="h-4 w-4" />
							Return to Home
						</div>
					</button>
				</div>
			</div>
		{/if}

		<!-- Reveal Room Code Dialog -->
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
							Are you sure you want to reveal the room code? This may be visible if you're
							streaming.
						</AlertDialog.Description>
					</AlertDialog.Header>
					<AlertDialog.Footer>
						<AlertDialog.Cancel onclick={() => (showRevealConfirm = false)}
							>Cancel</AlertDialog.Cancel
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
</TooltipProvider>
