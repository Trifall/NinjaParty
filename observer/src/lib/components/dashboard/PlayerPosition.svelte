<script lang="ts">
	import { settings } from '$lib/stores/settings.svelte';
	import type { PlayerPosition } from '$lib/types/stronghold';
	import { normalizeMinecraftAngle } from '$lib/utils/stronghold';

	interface Props {
		pos: PlayerPosition | undefined;
	}

	let { pos }: Props = $props();
</script>

{#if pos}
	{@const displayX =
		pos.isInNether && pos.xInOverworld !== undefined ? pos.xInOverworld / 8 : pos.xInOverworld}
	{@const displayZ =
		pos.isInNether && pos.zInOverworld !== undefined ? pos.zInOverworld / 8 : pos.zInOverworld}
	<div
		class="flex items-center justify-between border-t border-white/20 bg-white/[0.02] px-4 py-2.5 text-base"
	>
		<div class="flex items-center gap-5">
			<span class="flex items-center gap-1.5 text-gray-400"
				><svg
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
				</svg>Position:</span
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
