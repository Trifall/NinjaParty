<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		children?: Snippet;
		className?: string;
		textClassName?: string;
		noBorder?: boolean;
		inline?: boolean;
		href?: string;
	}

	let {
		children,
		className,
		textClassName,
		noBorder = false,
		inline = false,
		href,
	}: Props = $props();

	const textClassnames = $derived(
		cn(
			'inline-block animate-gradient bg-gradient-to-r from-[#00FFFF] via-[#FF1493] to-[#00FFFF] bg-[length:var(--bg-size)_100%] bg-clip-text pb-2 text-3xl font-semibold leading-normal text-transparent',
			textClassName
		)
	);
</script>

<div
	class={cn(
		'group relative mx-auto flex max-w-fit flex-row items-center justify-center rounded-2xl px-4 py-1.5 text-sm font-medium transition-shadow duration-500 ease-out [--bg-size:300%]',
		{
			'bg-white/40 shadow-[inset_0_-8px_10px_#8fdfff1f] backdrop-blur-sm hover:shadow-[inset_0_-5px_10px_#8fdfff3f] dark:bg-black/40':
				!noBorder,
		},
		{
			'mx-0 inline-flex px-0': inline,
		},
		className
	)}
>
	{#if !noBorder}
		<div
			class="animate-gradient bg-gradient-to-r from-[#ffaa40]/50 via-[#9c40ff]/50 to-[#ffaa40]/50 bg-[length:var(--bg-size)_100%] [border-radius:inherit] [mask-composite:subtract] [mask:linear-gradient(#fff_0_0)_content-box,linear-gradient(#fff_0_0)]"
		></div>
	{/if}

	{#if href}
		<a {href} target="_blank" rel="noopener noreferrer" class={textClassnames}>
			{@render children?.()}
		</a>
	{:else}
		<span class={textClassnames}>
			{@render children?.()}
		</span>
	{/if}
</div>
