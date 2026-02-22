<script lang="ts">
	import { scroll } from 'motion';
	import { onMount } from 'svelte';
	import ProjectLink from '$lib/components/projects/ProjectLink.svelte';
	import type { TProjectData } from '$lib/data/projects';
	import { useWindow } from '$lib/hooks/use-window.svelte';
	import { cn } from '$lib/utils';

	interface Props {
		content: TProjectData[];
		contentClassName?: string;
	}

	let { content, contentClassName }: Props = $props();

	const { isMobile, isTablet } = useWindow();

	let activeCard = $state(0);
	let container: HTMLElement | null = $state(null);

	let isDark = $state(true);

	onMount(() => {
		isDark = document.documentElement.classList.contains('dark');
		const observer = new MutationObserver(() => {
			isDark = document.documentElement.classList.contains('dark');
		});
		observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

		if (!container || !content || content.length === 0) return;

		const updateActiveCard = (progress: number) => {
			let activeCardIndex = 0;

			if (isMobile()) {
				if (progress < 0.22) activeCardIndex = 0;
				else if (progress < 0.415) activeCardIndex = 1;
				else if (progress < 0.593) activeCardIndex = 2;
				else if (progress < 0.78) activeCardIndex = 3;
				else activeCardIndex = 4;
			} else if (isTablet()) {
				if (progress < 0.25) activeCardIndex = 0;
				else if (progress < 0.43) activeCardIndex = 1;
				else if (progress < 0.593) activeCardIndex = 2;
				else if (progress < 0.74) activeCardIndex = 3;
				else activeCardIndex = 4;
			} else {
				if (progress < 0.332) activeCardIndex = 0;
				else if (progress < 0.443) activeCardIndex = 1;
				else if (progress < 0.57) activeCardIndex = 2;
				else if (progress < 0.68) activeCardIndex = 3;
				else activeCardIndex = 4;
			}

			activeCard = activeCardIndex;
		};

		const cleanup = scroll(
			(progress: number) => {
				updateActiveCard(progress);
			},
			{ target: container, offset: ['0.15 1', '0.85 0'] }
		);

		return () => {
			cleanup?.();
			observer.disconnect();
		};
	});
</script>

<div bind:this={container} class="relative flex justify-center space-x-10 rounded-md py-0 lg:p-10">
	<div class="div relative flex items-start px-4 sm:pl-12">
		<div class="max-w-2xl py-0 lg:py-10">
			{#each content as item, index (item.title + index)}
				{#if index < 5}
					<div
						class="{index === 0
							? 'mb-28 mt-0 lg:mb-28 lg:mt-20'
							: 'mb-28 mt-20'} min-h-48 rounded-xl border-4 {activeCard === index
							? 'border-black'
							: 'border-transparent'} p-4 transition-colors duration-500 dark:border-none"
						style="background-color: {activeCard === index
							? isDark
								? 'black'
								: 'white'
							: 'transparent'};"
					>
						<h2
							class="text-2xl font-bold text-slate-600 transition-opacity duration-500 dark:text-slate-100"
							style="opacity: {activeCard === index ? 1 : 0.3};"
						>
							{item.title}
						</h2>
						<p
							class="text-kg mt-10 max-w-sm text-slate-800 transition-opacity duration-500 dark:text-slate-300"
							style="opacity: {activeCard === index ? 1 : 0.3};"
						>
							{item.description}
						</p>
						<div class="flex items-center justify-center lg:hidden">
							<div
								class="mt-8 flex h-fit w-full items-center justify-center overflow-hidden rounded-md border-2 bg-slate-200 text-slate-500 transition-opacity duration-500 dark:border-white dark:bg-black"
								style="opacity: {activeCard === index ? 1 : 0.3};"
							>
								{#if item.content}
									<!-- eslint-disable-next-line svelte/no-at-html-tags -->
									{@html item.content}
								{/if}
							</div>
						</div>
						<div class="flex items-center justify-center">
							<div
								class="card-link-container mt-4 flex h-fit w-full items-center justify-center gap-4 overflow-hidden"
							>
								{#each item.link as link (link.link)}
									<ProjectLink linkData={link} isCardActive={activeCard === index} />
								{/each}
							</div>
						</div>
					</div>
				{/if}
			{/each}
		</div>
	</div>
	<div
		class={cn(
			'sticky top-1/3 hidden overflow-hidden rounded-xl',
			'sm:max-h-[16.5rem] sm:w-[500px] lg:mb-12 lg:block lg:min-h-[20rem] lg:w-[600px] lg:min-w-[560px] xl:min-h-[24rem] xl:w-[700px]',
			contentClassName
		)}
	>
		<div
			class="relative h-full w-full overflow-hidden rounded-xl bg-gradient-to-r from-[#FF1493] to-[#00FFFF]"
		>
			<div
				class="absolute inset-0 m-[3px] overflow-hidden overflow-x-hidden rounded-xl bg-white dark:bg-black"
			>
				{#if content[activeCard]?.content}
					<!-- eslint-disable-next-line svelte/no-at-html-tags -->
					{@html content[activeCard].content}
				{/if}
			</div>
		</div>
	</div>
</div>
