<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		className?: string;
	}

	let { className }: Props = $props();

	let container: HTMLDivElement;
	let scaleX = $state(0);

	onMount(() => {
		const updateProgress = () => {
			const scrollY = window.scrollY;
			const maxScroll = document.documentElement.scrollHeight - window.innerHeight;
			scaleX = scrollY / maxScroll;
		};

		window.addEventListener('scroll', updateProgress, { passive: true });
		updateProgress();

		return () => window.removeEventListener('scroll', updateProgress);
	});
</script>

<div
	bind:this={container}
	class={cn(
		'fixed inset-x-0 top-0 z-[1000] h-1 origin-left bg-gradient-to-r from-[#A97CF8] via-[#F38CB8] to-[#FDCC92]',
		className
	)}
	style="transform: scaleX({scaleX})"
></div>
