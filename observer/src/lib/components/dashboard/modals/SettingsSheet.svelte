<script lang="ts">
	import { Settings } from '@lucide/svelte';
	import * as Sheet from '$lib/components/ui/sheet';
	import { VIEW_MODE, settings } from '$lib/stores/settings.svelte';

	interface Props {
		compact?: boolean;
	}

	let { compact = false }: Props = $props();

	let currentDisplayType = $derived(settings.displayType.current);
	let currentViewMode = $derived(settings.viewMode.current);
	let currentEmbedAutoHide = $derived(settings.embedAutoHide.current);

	const AUTO_HIDE_OPTIONS = [
		{ value: 0, label: 'Disabled' },
		{ value: 5, label: '5 seconds' },
		{ value: 15, label: '15 seconds' },
		{ value: 30, label: '30 seconds' },
		{ value: 60, label: '60 seconds' },
	];
</script>

<Sheet.Root>
	<Sheet.Trigger
		class="inline-flex items-center justify-center whitespace-nowrap rounded-md border border-white/10 bg-black/40 font-medium text-white ring-offset-background backdrop-blur-xl transition-colors hover:bg-white/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 {compact
			? 'h-6 w-6 text-xs'
			: 'h-7 w-7 text-sm'}"
	>
		<Settings class={compact ? 'h-3 w-3' : 'h-4 w-4'} />
	</Sheet.Trigger>
	<Sheet.Content class="border-l-white/10 bg-black/95 text-white backdrop-blur-xl">
		<Sheet.Header>
			<Sheet.Title class="text-white">Settings</Sheet.Title>
			<Sheet.Description class="text-muted-foreground">
				Configure the observer dashboard display. Note: Since these are purely presentation
				preferences, they are saved locally in your browser.
			</Sheet.Description>
		</Sheet.Header>

		<div class="grid gap-6 py-6">
			<!-- View Mode -->
			<div class="space-y-4">
				<h4 class="text-sm font-medium leading-none text-white">View Mode</h4>
				<div class="flex flex-col space-y-3">
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="viewMode"
							value={VIEW_MODE.MAX}
							checked={currentViewMode === VIEW_MODE.MAX}
							onchange={() => (settings.viewMode.current = VIEW_MODE.MAX)}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Max View <span class="ml-1 text-muted-foreground"
								>- Full dashboard with all data</span
							></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="viewMode"
							value={VIEW_MODE.COMPACT}
							checked={currentViewMode === VIEW_MODE.COMPACT}
							onchange={() => (settings.viewMode.current = VIEW_MODE.COMPACT)}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Compact View <span class="ml-1 text-muted-foreground"
								>- Location + Position only</span
							></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="viewMode"
							value={VIEW_MODE.ULTRA_COMPACT}
							checked={currentViewMode === VIEW_MODE.ULTRA_COMPACT}
							onchange={() => (settings.viewMode.current = VIEW_MODE.ULTRA_COMPACT)}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Ultra Compact <span class="ml-1 text-muted-foreground">- Minimal display</span></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="viewMode"
							value={VIEW_MODE.GRID}
							checked={currentViewMode === VIEW_MODE.GRID}
							onchange={() => (settings.viewMode.current = VIEW_MODE.GRID)}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Grid View <span class="ml-1 text-muted-foreground">- 3 Column Layout</span></span
						>
					</label>
				</div>
			</div>

			<div class="space-y-4 border-t border-white/10 pt-4">
				<h4 class="text-sm font-medium leading-none text-white">Coordinate Display Type</h4>
				<div class="flex flex-col space-y-3">
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="displayType"
							value="FOUR_FOUR"
							checked={currentDisplayType === 'FOUR_FOUR'}
							onchange={() => (settings.displayType.current = 'FOUR_FOUR')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Location (4, 4) <span class="ml-1 text-muted-foreground"
								>- Default for starting staircase</span
							></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="displayType"
							value="EIGHT_EIGHT"
							checked={currentDisplayType === 'EIGHT_EIGHT'}
							onchange={() => (settings.displayType.current = 'EIGHT_EIGHT')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Location (8, 8) <span class="ml-1 text-muted-foreground"
								>- Center of stronghold chunk</span
							></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="displayType"
							value="CHUNK"
							checked={currentDisplayType === 'CHUNK'}
							onchange={() => (settings.displayType.current = 'CHUNK')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Chunk Coordinates <span class="ml-1 text-muted-foreground">- Raw chunk output</span
							></span
						>
					</label>
				</div>
			</div>

			<div class="space-y-4 border-t border-white/10 pt-4">
				<h4 class="text-sm font-medium leading-none text-white">Display Options</h4>
				<div class="flex flex-col space-y-3">
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="checkbox"
							checked={settings.colorNegatives.current}
							onchange={(e) => (settings.colorNegatives.current = e.currentTarget.checked)}
							class="h-4 w-4 rounded accent-emerald-500"
						/>
						<span class="text-sm"
							>Color negative coordinates <span class="ml-1 text-muted-foreground"
								>- Show negative values in red</span
							></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="checkbox"
							checked={settings.reorderOfflineRunners.current}
							onchange={(e) =>
								(settings.reorderOfflineRunners.current = e.currentTarget.checked)}
							class="h-4 w-4 rounded accent-emerald-500"
						/>
						<span class="text-sm"
							>Reorder offline runners <span class="ml-1 text-muted-foreground"
								>- Move offline runners to bottom</span
							></span
						>
					</label>
				</div>
			</div>

			<!-- Embed Auto-Hide -->
			<div class="space-y-4 border-t border-white/10 pt-4">
				<div>
					<h4 class="text-sm font-medium leading-none text-white">Embed Auto-Hide</h4>
					<p class="mt-1.5 text-xs text-muted-foreground">
						In OBS embed mode, fade out the overlay after this long with no runner data changes.
					</p>
				</div>
				<div class="flex flex-col space-y-3">
					{#each AUTO_HIDE_OPTIONS as option (option.value)}
						<label class="flex cursor-pointer items-center space-x-3">
							<input
								type="radio"
								name="embedAutoHide"
								value={option.value}
								checked={currentEmbedAutoHide === option.value}
								onchange={() => (settings.embedAutoHide.current = option.value)}
								class="h-4 w-4 accent-emerald-500"
							/>
							<span class="text-sm">{option.label}</span>
						</label>
					{/each}
				</div>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
