<script lang="ts">
	import { onMount } from 'svelte';

	interface HyperTextProps {
		text: string;
		duration?: number;
		className?: string;
		animateOnLoad?: boolean;
	}

	let { text, duration = 200, className, animateOnLoad = true }: HyperTextProps = $props();

	let displayText = $derived<string[]>(text.split(''));
	let interations = $state(0);
	let isFirstRender = $state(true);

	const alphabets = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('');
	const getRandomInt = (max: number) => Math.floor(Math.random() * max);

	onMount(() => {
		const interval = setInterval(
			() => {
				if (!animateOnLoad && isFirstRender) {
					clearInterval(interval);
					isFirstRender = false;
					return;
				}
				if (interations < text.length) {
					displayText = displayText.map((l, i) =>
						l === ' ' ? l : i <= interations ? text[i] : alphabets[getRandomInt(26)]
					);
					interations = interations + 0.05;
				} else {
					clearInterval(interval);
				}
			},
			duration / 55 / (text.length * 1000)
		);

		return () => clearInterval(interval);
	});
</script>

<div class="flex scale-100 cursor-default overflow-hidden">
	{#each displayText as letter, i (i)}
		<h1 class="font-mono {letter === ' ' ? 'w-3' : ''} {className}">
			{letter.toUpperCase()}
		</h1>
	{/each}
</div>
