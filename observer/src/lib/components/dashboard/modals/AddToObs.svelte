<script lang="ts">
	import { Monitor } from '@lucide/svelte';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { serializeRunnerMoves } from '$lib/stores/room-preferences.svelte';
	import { VIEW_MODE, settings } from '$lib/stores/settings.svelte';

	interface Props {
		roomCode: string;
		defaultRunnerOrder: string[];
		runnerOrder: string[];
	}

	let { roomCode, defaultRunnerOrder, runnerOrder }: Props = $props();

	let showObsModal = $state(false);
	let copiedLink = $state(false);

	let obsUrl = $derived.by(() => {
		const params = new SvelteURLSearchParams({ embed: '1' });

		const viewMode = settings.viewMode.current;
		if (viewMode !== VIEW_MODE.MAX) {
			params.set('viewMode', viewMode);
		}

		const displayType = settings.displayType.current;
		if (displayType !== 'FOUR_FOUR') {
			params.set('displayType', displayType);
		}

		if (!settings.colorNegatives.current) {
			params.set('colorNegatives', 'false');
		}

		const autoHide = settings.embedAutoHide.current;
		if (autoHide > 0) {
			params.set('autoHide', String(autoHide));
		}

		// encode only the deltas from the default live ordering
		const runnerMoves = serializeRunnerMoves(defaultRunnerOrder, runnerOrder);
		if (runnerMoves) {
			params.set('runnerMoves', runnerMoves);
		}

		// reorder offline runners, only include if enabled
		if (settings.reorderOfflineRunners.current) {
			params.set('reorderOfflineRunners', 'true');
		}

		return `/room/${roomCode}?${params.toString()}`;
	});

	function copyLink() {
		navigator.clipboard.writeText(window.location.origin + obsUrl);
		copiedLink = true;
		setTimeout(() => (copiedLink = false), 2000);
	}
</script>

<div class="mt-4 flex justify-center">
	<button
		type="button"
		onclick={() => (showObsModal = true)}
		class="group relative rounded-lg p-[2px]"
	>
		<div
			class="animate-gradient-rotate absolute inset-0 rounded-lg bg-gradient-to-r from-emerald-500 via-cyan-500 to-emerald-500"
		></div>
		<div
			class="relative flex items-center gap-2 rounded-lg border-0 bg-black px-4 py-2 text-sm font-medium text-white transition-all hover:bg-black/60 hover:text-white"
		>
			<Monitor class="h-4 w-4" />
			Add to OBS
		</div>
	</button>
</div>

<AlertDialog.Root
	open={showObsModal}
	onOpenChange={(open) => {
		showObsModal = open;
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
				<AlertDialog.Title class="flex items-center gap-2">
					<Monitor class="h-5 w-5" />
					Add to OBS
				</AlertDialog.Title>
				<AlertDialog.Description class="space-y-4">
					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium text-white">1. Copy this link:</span>
						<p class="text-xs text-white/50">
							All your current settings are baked into this URL — the OBS browser source will
							display exactly as you've configured it.
						</p>
						<div style="display:grid; grid-template-columns: 1fr auto; gap: 0.5rem; min-width:0;">
							<code
								style="overflow-x:auto; white-space:nowrap; min-width:0;"
								class="rounded border border-white/10 bg-black px-3 py-2 text-sm text-emerald-400"
							>
								{window.location.origin}{obsUrl}
							</code>
							<button
								type="button"
								onclick={copyLink}
								class="self-center rounded bg-emerald-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700"
							>
								{copiedLink ? 'Copied!' : 'Copy'}
							</button>
						</div>
					</div>

					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium text-white">2. Add browser source in OBS</span>
						<span class="text-sm text-white/60">Paste the above link into the URL field</span>
					</div>

					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium text-white">3. Set dimensions:</span>
						<span class="text-sm text-white/60">Width: 800, Height: 900</span>
						<span class="text-sm text-white/60"
							>(Note: You can increase this if you have more runners and it cuts off)</span
						>
					</div>

					<div class="flex flex-col gap-2">
						<img
							src="/images/obs_embed_settings.png"
							alt="OBS Embed Settings"
							class="rounded-lg border border-white/10"
						/>
					</div>

					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium text-white">4. Click OK</span>
					</div>
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Action onclick={() => (showObsModal = false)}>Done</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Portal>
</AlertDialog.Root>

<style>
	@keyframes gradient-shift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradient-rotate {
		animation: gradient-shift 3s ease infinite;
		background-size: 200% 200%;
	}
</style>
