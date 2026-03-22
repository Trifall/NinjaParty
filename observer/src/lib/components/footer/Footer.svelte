<script lang="ts">
	import AnimatedGradientText from '$lib/components/ui/AnimatedGradientText.svelte';
	import { Icons } from '$lib/components/ui/Icons.svelte';
	import {
		Tooltip,
		TooltipContent,
		TooltipProvider,
		TooltipTrigger,
	} from '$lib/components/ui/tooltip';

	interface Props {
		authorName?: string;
	}

	let { authorName = 'Jerren' }: Props = $props();
</script>

<footer class="w-full border-t border-gray-200 py-6 dark:border-gray-800">
	<div class="mx-auto max-w-4xl px-4">
		<div class="flex flex-col items-center justify-center space-y-2">
			<TooltipProvider>
				<div class="flex items-center space-x-4 px-4">
					<Tooltip delayDuration={0}>
						<TooltipTrigger class="p-1">
							<a
								href="https://github.com/Trifall/NinjaParty"
								class="text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
								target="_blank"
								rel="noopener noreferrer"
							>
								<Icons.GitHub fill="#ffffff" class="h-8 w-8 hover:fill-gray-300" />
							</a>
						</TooltipTrigger>
						<TooltipContent side="top">GitHub</TooltipContent>
					</Tooltip>
				</div>
				<span class="flex items-center gap-1 text-sm text-gray-800 dark:text-gray-200 md:text-lg">
					Created by
					<AnimatedGradientText
						noBorder
						className="p-0"
						textClassName="text-sm font-bold mb-0 md:text-lg pb-0"
						href="https://github.com/Trifall"
					>
						{authorName}
					</AnimatedGradientText>
				</span>
				<a
					href="https://github.com/Trifall/NinjaParty/releases"
					target="_blank"
					rel="noopener noreferrer"
					class="text-lg font-bold text-emerald-400 transition-colors hover:text-emerald-300 hover:underline"
				>
					v{__APP_VERSION__}
				</a>
			</TooltipProvider>
		</div>
	</div>
</footer>
