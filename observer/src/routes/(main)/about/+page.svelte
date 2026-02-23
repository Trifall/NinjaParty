<script lang="ts">
	import { ArrowLeft, CodeXml, Github, Radio } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
</script>

<svelte:head>
	<title>About - Ninjabrain Party</title>
</svelte:head>

<div class="flex min-h-[90vh] w-full flex-col items-center justify-center px-4 py-16">
	<div class="relative w-full max-w-[640px] duration-700 animate-in fade-in slide-in-from-bottom-6">
		<!-- Glow backdrop -->
		<div
			class="absolute -inset-2 rounded-3xl bg-gradient-to-br from-emerald-500/15 via-cyan-500/10 to-transparent blur-2xl"
		></div>

		<div
			class="relative rounded-2xl border border-white/10 bg-black/50 shadow-2xl backdrop-blur-xl"
		>
			<!-- Header -->
			<div class="border-white/8 border-b px-8 py-6">
				<a
					href="/"
					class="mb-5 inline-flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-2 text-xs font-medium text-white/80 transition-colors hover:bg-white/10 hover:text-white"
				>
					<ArrowLeft class="h-3.5 w-3.5" />
					Back to home
				</a>

				<h1
					class="bg-gradient-to-br from-white via-white/90 to-white/50 bg-clip-text text-4xl font-black leading-tight tracking-tight text-transparent"
				>
					Ninjabrain<br />Party
				</h1>
				<p class="mt-2 text-sm font-medium uppercase tracking-wide text-emerald-400/80">
					Real-time Ninjabrain Bot observation
				</p>
			</div>

			<!-- Body -->
			<div class="space-y-7 px-8 py-7">
				<p class="text-[15px] font-semibold leading-relaxed text-white/80">
					A tool to connect a player's Ninjabrain Bot data to observers.
				</p>

				<!-- Feature cards -->
				<div class="space-y-3">
					<div class="border-white/6 bg-white/3 flex gap-4 rounded-xl border p-4 transition-colors">
						<div
							class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-emerald-500/15 text-emerald-400"
						>
							<Radio class="h-4 w-4" />
						</div>
						<div class="w-full">
							<p class="text-sm font-semibold text-white">How it works</p>
							<p class="mt-1 text-sm leading-relaxed text-white/80">
								A local relay runs alongside Ninjabrain Bot, reading its data and forwarding it to a
								central server. Observers connect via a simple room code and see live stronghold
								positions, eye throw data, and player stats through WebSockets.
							</p>
							<img
								src="/images/system_architecture.png"
								alt="System Architecture"
								class="mt-4 max-w-full rounded-lg border border-white/10"
							/>
						</div>
					</div>

					<div class="border-white/6 bg-white/3 flex gap-4 rounded-xl border p-4 transition-colors">
						<div
							class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-500/15 text-violet-400"
						>
							<CodeXml class="h-4 w-4" />
						</div>
						<div>
							<p class="text-sm font-semibold text-white">Open source</p>
							<p class="mt-1 text-sm leading-relaxed text-white/80">
								Fully open source. The websocket server is self-hostable. No personal data is stored
								nor collected. View the code on GitHub below.
							</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="border-white/8 flex items-center justify-center border-t px-8 py-5">
				<a
					href="https://github.com/Trifall/Ninjabrain-Party"
					target="_blank"
					rel="noopener noreferrer"
					class="animate-gradient-rotate relative rounded-lg p-[2px]"
				>
					<div
						class="absolute inset-0 rounded-lg bg-[conic-gradient(from_0deg,theme(colors.emerald.500),theme(colors.cyan.500),theme(colors.emerald.500))]"
					></div>
					<Button
						variant="outline"
						class="relative gap-2 border-0 bg-black/90 text-white/80 transition-colors hover:bg-black hover:text-white"
					>
						<Github class="h-4 w-4" />
						View Source on GitHub
					</Button>
				</a>
			</div>
		</div>
	</div>
</div>
