<script lang="ts">
	import { Settings } from '@lucide/svelte';
	import * as Sheet from '$lib/components/ui/sheet';
	import { settings } from '$lib/stores/settings.svelte';

	let currentDisplayType = $derived(settings.displayType.current);
	let currentMcVersion = $derived(settings.mcVersion.current);
	let currentViewMode = $derived(settings.viewMode.current);
</script>

<Sheet.Root>
	<Sheet.Trigger
		class="inline-flex h-7 w-7 items-center justify-center whitespace-nowrap rounded-md border border-white/10 bg-black/40 text-sm font-medium text-white ring-offset-background backdrop-blur-xl transition-colors hover:bg-white/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
	>
		<Settings class="h-4 w-4" />
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
							value="max"
							checked={currentViewMode === 'max'}
							onchange={() => (settings.viewMode.current = 'max')}
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
							value="compact"
							checked={currentViewMode === 'compact'}
							onchange={() => (settings.viewMode.current = 'compact')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>Compact View <span class="ml-1 text-muted-foreground"
								>- Location + Position only</span
							></span
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
				<h4 class="text-sm font-medium leading-none text-white">Minecraft Version</h4>
				<div class="flex flex-col space-y-3">
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="mcVersion"
							value="PRE_119"
							checked={currentMcVersion === 'PRE_119'}
							onchange={() => (settings.mcVersion.current = 'PRE_119')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm"
							>1.18 and below <span class="ml-1 text-muted-foreground">- Most common</span></span
						>
					</label>
					<label class="flex cursor-pointer items-center space-x-3">
						<input
							type="radio"
							name="mcVersion"
							value="POST_119"
							checked={currentMcVersion === 'POST_119'}
							onchange={() => (settings.mcVersion.current = 'POST_119')}
							class="h-4 w-4 accent-emerald-500"
						/>
						<span class="text-sm">1.19 and above</span>
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
				</div>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
