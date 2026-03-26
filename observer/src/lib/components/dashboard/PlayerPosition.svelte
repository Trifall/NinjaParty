<script lang="ts">
	import { settings } from '$lib/stores/settings.svelte';
	import type { PlayerPosition } from '$lib/types/stronghold';
	import { normalizeMinecraftAngle } from '$lib/utils/stronghold';

	interface Props {
		pos: PlayerPosition | undefined;
		inline?: boolean;
		colorNegatives?: boolean;
	}

	let { pos, inline = false, colorNegatives: colorNegativesProp }: Props = $props();
	let colorNegatives = $derived(colorNegativesProp ?? settings.colorNegatives.current);
</script>

{#if pos}
	{@const displayX =
		pos.isInNether && pos.xInOverworld !== undefined ? pos.xInOverworld / 8 : pos.xInOverworld}
	{@const displayZ =
		pos.isInNether && pos.zInOverworld !== undefined ? pos.zInOverworld / 8 : pos.zInOverworld}

	{#if inline}
		<!-- Ultra-compact inline display: position + dimension only, no angle -->
		<div class="flex items-center gap-2 text-xs text-gray-400">
			<span>Position:</span>
			<span>
				(<span
					class={colorNegatives && displayX !== undefined && displayX < 0
						? 'text-red-400'
						: 'text-gray-200'}>{displayX?.toFixed(0) ?? '?'}</span
				>,
				<span
					class={colorNegatives && displayZ !== undefined && displayZ < 0
						? 'text-red-400'
						: 'text-gray-200'}>{displayZ?.toFixed(0) ?? '?'}</span
				>)
			</span>
			<span class={pos.isInNether ? 'text-red-400' : 'text-emerald-400'}>
				{pos.isInNether ? 'Nether' : pos.isInOverworld ? 'Overworld' : '?'}
			</span>
		</div>
	{:else}
		<!-- Standard display -->
		<div class="flex items-center justify-between overflow-hidden rounded-lg border border-white/10 bg-white/[0.02] px-4 py-2.5 text-base">
			<div class="flex items-center gap-5">
				<span class="flex items-center gap-1.5 text-gray-400">
					<svg
						class="h-4 w-4"
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z" />
						<circle cx="12" cy="10" r="3" />
					</svg>Position:
				</span>
				<span class="text-gray-200">
					(<span
						class={colorNegatives && displayX !== undefined && displayX < 0 ? 'text-red-400' : ''}
						>{displayX?.toFixed(2) ?? '?'}</span
					>,
					<span
						class={colorNegatives && displayZ !== undefined && displayZ < 0 ? 'text-red-400' : ''}
						>{displayZ?.toFixed(2) ?? '?'}</span
					>)
				</span>
			</div>
			<div class="flex items-center gap-5">
				<span class="text-gray-400">Angle:</span>
				<span class="text-gray-200">
					{pos.horizontalAngle !== undefined
						? normalizeMinecraftAngle(pos.horizontalAngle).toFixed(2) + '°'
						: '-'}
				</span>
			</div>
			<div class="flex items-center gap-5">
				<span class="text-gray-400">Dimension:</span>
				<span class={pos.isInNether ? 'text-red-400' : 'text-emerald-400'}>
					{pos.isInNether ? 'Nether' : pos.isInOverworld ? 'Overworld' : '?'}
				</span>
			</div>
		</div>
	{/if}
{/if}
