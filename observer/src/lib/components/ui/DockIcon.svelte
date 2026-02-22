<script lang="ts">
	import { getContext } from 'svelte';
	import type { Snippet } from 'svelte';
	import { spring } from 'svelte/motion';
	import { useWindow } from '$lib/hooks/use-window.svelte';
	import { cn } from '$lib/utils';

	interface DockIconProps {
		className?: string;
		children?: Snippet;
	}

	let { className, children }: DockIconProps = $props();

	let iconRef: HTMLDivElement | undefined = $state(undefined);

	const { isDesktop } = useWindow();

	// Get dock context
	const dockContext = getContext<{
		mouseX: number;
		magnification: number;
		distance: number;
	}>('dock');

	// Responsive base sizes
	const baseSize = $derived(isDesktop() ? 72 : 64);

	// Detect Firefox for performance optimizations
	const isFirefox =
		typeof navigator !== 'undefined' && navigator.userAgent.toLowerCase().includes('firefox');

	// Use scale transform instead of width - much more performant
	// Scale of 1 = baseSize, scale > 1 = magnified
	// Firefox gets less stiff springs for smoother animation
	const scaleSpring = spring(1, {
		stiffness: isFirefox ? 0.08 : 0.12,
		damping: isFirefox ? 0.35 : 0.28,
		precision: 0.005,
	});

	// Cache bounds calculation to avoid layout thrashing
	let iconBounds: DOMRect | null = null;
	let boundsTimeout: ReturnType<typeof setTimeout> | null = null;

	function updateBounds() {
		if (iconRef) {
			iconBounds = iconRef.getBoundingClientRect();
		}
	}

	// Update scale based on mouse position
	$effect(() => {
		if (!iconRef || !dockContext) return;

		const mouseX = dockContext.mouseX;
		const magnification = dockContext.magnification;
		const distance = dockContext.distance;

		if (mouseX === Infinity) {
			// Mouse left the dock, reset to base scale
			scaleSpring.set(1);
			iconBounds = null;
			return;
		}

		// Update bounds if not cached (debounced)
		if (!iconBounds) {
			updateBounds();
			// Clear bounds after mouse stops moving for a bit
			if (boundsTimeout) clearTimeout(boundsTimeout);
			boundsTimeout = setTimeout(() => {
				iconBounds = null;
			}, 100);
		}

		if (!iconBounds) return;

		const iconCenterX = iconBounds.x + iconBounds.width / 2;
		const distanceFromMouse = mouseX - iconCenterX;

		// Calculate target scale
		const absDistance = Math.abs(distanceFromMouse);
		const normalizedDistance = Math.min(absDistance / distance, 1);
		const scaleFactor = Math.cos((normalizedDistance * Math.PI) / 2);

		// Cap the max scale to prevent overlap with neighboring icons
		// Max scale limited to 1.15 (15% larger) to stay within gap boundaries
		const rawScale = 1 + (magnification / baseSize - 1) * scaleFactor;
		const targetScale = Math.min(rawScale, 1.15);

		scaleSpring.set(targetScale);
	});

	// Reset scale when baseSize changes and mouse is not over dock
	$effect(() => {
		if (dockContext?.mouseX === Infinity) {
			scaleSpring.set(1);
		}
	});
</script>

<div
	bind:this={iconRef}
	class={cn(
		'flex aspect-square cursor-pointer items-center justify-center rounded-full bg-neutral-400/40 transition-colors hover:bg-neutral-400/60',
		className
	)}
	style="
		width: {baseSize}px;
		height: {baseSize}px;
		transform: scale({$scaleSpring});
		transform-origin: bottom center;
		will-change: transform;
	"
>
	{@render children?.()}
</div>
