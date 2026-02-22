<script lang="ts">
	import { setContext } from 'svelte';
	import type { Snippet } from 'svelte';
	import { useWindow } from '$lib/hooks/use-window.svelte';
	import { cn } from '$lib/utils';

	interface DockProps {
		className?: string;
		children?: Snippet;
	}

	let { className, children }: DockProps = $props();

	let dockRef: HTMLDivElement;
	let mouseX = $state(Infinity);

	const { isMobile, isDesktop } = useWindow();

	// Default values with responsive sizes
	const DEFAULT_MAGNIFICATION = {
		small: 72,
		regular: 84,
		large: 96,
	};

	const DEFAULT_DISTANCE = {
		small: 48,
		regular: 64,
		large: 80,
	};

	// Calculate responsive values based on window size
	const magnification = $derived(
		isMobile()
			? DEFAULT_MAGNIFICATION.small
			: isDesktop()
				? DEFAULT_MAGNIFICATION.large
				: DEFAULT_MAGNIFICATION.regular
	);

	const distance = $derived(
		isMobile()
			? DEFAULT_DISTANCE.small
			: isDesktop()
				? DEFAULT_DISTANCE.large
				: DEFAULT_DISTANCE.regular
	);

	// Provide context to child DockIcon components
	setContext('dock', {
		get mouseX() {
			return mouseX;
		},
		get magnification() {
			return magnification;
		},
		get distance() {
			return distance;
		},
	});

	function handleMouseMove(e: MouseEvent) {
		mouseX = e.pageX;
	}

	function handleMouseLeave() {
		mouseX = Infinity;
	}
</script>

<div
	bind:this={dockRef}
	role="toolbar"
	tabindex="0"
	onmousemove={handleMouseMove}
	onmouseleave={handleMouseLeave}
	class={cn(
		'mx-auto mt-8 flex h-[55px] w-max items-end gap-2 rounded-2xl border px-0 py-2 dark:border-[#707070] xs:h-[78px]',
		className
	)}
>
	{@render children?.()}
</div>
