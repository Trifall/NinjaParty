import { onMount } from 'svelte';

type WindowSize = {
	width: number | undefined;
	height: number | undefined;
};

type WindowDimensions = {
	windowSize: () => WindowSize;
	isMobile: () => boolean;
	isTablet: () => boolean;
	isDesktop: () => boolean;
};

export function useWindow(): WindowDimensions {
	let windowSize = $state<WindowSize>({ width: undefined, height: undefined });

	onMount(() => {
		const handleResize = () => {
			windowSize = {
				width: window.innerWidth,
				height: window.innerHeight,
			};
		};

		window.addEventListener('resize', handleResize);
		handleResize();

		return () => window.removeEventListener('resize', handleResize);
	});

	return {
		windowSize: () => windowSize,
		isMobile: () => typeof windowSize.width === 'number' && windowSize.width <= 480,
		isTablet: () =>
			typeof windowSize.width === 'number' && windowSize.width < 1024 && windowSize.width > 480,
		isDesktop: () => typeof windowSize.width === 'number' && windowSize.width >= 1024,
	};
}
