<script lang="ts">
	import { Eye, EyeOff, Home, House, LoaderCircle, Wifi, WifiOff } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Badge } from '$lib/components/ui/badge';
	import type { RoomConnection } from '$lib/stores/room.svelte';
	import AnimatedGradientText from '$components/ui/AnimatedGradientText.svelte';
	import AllAdvancements from './AllAdvancements.svelte';
	import BlindMode from './BlindMode.svelte';
	import DivineMode from './DivineMode.svelte';
	import EyeThrows from './EyeThrows.svelte';
	import LocationsTable from './LocationsTable.svelte';
	import PlayerPosition from './PlayerPosition.svelte';
	import AddToObs from './modals/AddToObs.svelte';
	import SettingsSheet from './modals/SettingsSheet.svelte';

	let { roomStore, roomCode, isEmbed }: { roomStore: RoomConnection; roomCode: string; isEmbed: boolean } = $props();

	let isRoomCodeHidden = $state(true);
	let showRevealConfirm = $state(false);

	let payload = $derived(roomStore.stronghold);
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
</script>

<div class="mx-auto w-full max-w-3xl">
	<!-- Header -->
	<div
		class="flex items-center justify-between rounded-t-lg border border-b-0 border-white/20 bg-black px-5 py-3 backdrop-blur-xl"
	>
		<div class="flex items-center gap-2">
			<span class="flex items-center gap-1 text-sm text-gray-800 dark:text-gray-200 md:text-lg">
				<AnimatedGradientText
					noBorder
					className="p-0"
					spanClassName="text-sm font-bold mb-0 md:text-lg pb-0"
				>
					Ninjabrain Party
				</AnimatedGradientText>
			</span>

			{#if !isEmbed}
				<span class="text-gray-400">|</span>
				<span class="text-base text-gray-400">Room:</span>
				{#if isRoomCodeHidden}
					<span class="text-base font-bold uppercase tracking-wider text-emerald-400">******</span>
				{:else}
					<span class="text-base font-bold uppercase tracking-wider text-emerald-400"
						>{roomCode}</span
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
						<EyeOff class="mt-1 h-4 w-4" />
					{:else}
						<Eye class="mt-1 h-4 w-4" />
					{/if}
				</button>
			{/if}
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
			{#if !isEmbed}
				<SettingsSheet />
			{/if}
		</div>
	</div>

	{#if isEmbed && roomStore.roomClosed}
		<div class="mt-4 rounded-lg border border-red-500/50 bg-red-500/10 px-4 py-3 text-center text-red-400">
			This room has been closed
		</div>
	{/if}

	<!-- Main Panel -->
	<div
		class="rounded-b-lg border border-t-0 border-white/20 bg-[#1a1a1a] text-sm text-gray-200 backdrop-blur-xl"
	>
		{#if roomStore.status === 'connected' && !payload}
			<div class="flex h-48 flex-col items-center justify-center gap-3 text-muted-foreground">
				<LoaderCircle class="h-8 w-8 animate-spin text-emerald-500" />
				<p>Waiting for Ninjabrain Bot data...</p>
			</div>
		{:else if payload}
			<LocationsTable
				resultType={payload.resultType}
				predictions={payload.predictions?.slice(0, 5) || []}
				{pos}
			/>

			<PlayerPosition {pos} />

			<EyeThrows {throws} />

			<BlindMode blind={roomStore.blind ?? undefined} />

			<DivineMode divine={roomStore.divine ?? undefined} />

			<AllAdvancements allAdvancements={roomStore.allAdvancements ?? undefined} />
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
