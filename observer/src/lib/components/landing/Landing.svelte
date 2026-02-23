<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Search, Loader2 } from '@lucide/svelte';

	let roomCode = $state('');
	let isLoading = $state(false);
	let error = $state('');

	async function joinRoom() {
		if (!roomCode.trim()) {
			error = 'Please enter a room code';
			return;
		}

		if (roomCode.trim().length !== 6) {
			error = 'Room code must be exactly 6 characters';
			return;
		}

		isLoading = true;
		error = '';

		// Optional: you could make a quick REST API call here to verify the room exists
		// before routing, but for now we'll just route directly to it.

		await goto(`/room/${roomCode.trim().toLowerCase()}`);
		isLoading = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			joinRoom();
		}
	}
</script>

<div class="flex h-[80vh] w-full flex-col items-center justify-center px-4">
	<div class="relative w-full max-w-md animate-in fade-in slide-in-from-bottom-8 duration-700">
		<div
			class="absolute -inset-1 rounded-2xl bg-gradient-to-r from-emerald-500 to-cyan-500 opacity-20 blur-xl transition-all duration-500 group-hover:opacity-40"
		></div>

		<div class="relative flex flex-col items-center space-y-8 rounded-2xl border border-white/10 bg-black/40 p-10 text-center shadow-2xl backdrop-blur-xl">
			<div class="space-y-3">
				<h1 class="bg-gradient-to-br from-white to-white/60 bg-clip-text text-4xl font-extrabold tracking-tight text-transparent sm:text-5xl">
					Ninjabrain Party
				</h1>
				<p class="text-sm font-medium text-muted-foreground sm:text-base">
					Real-time stronghold triangulation for competitive Minecraft speedrunners and coaches.
				</p>
			</div>

			<div class="w-full space-y-4">
				<div class="relative">
					<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-muted-foreground">
						<Search class="h-5 w-5" />
					</div>
					<Input
						type="text"
						placeholder="Enter 6-character room code"
						class="h-14 w-full bg-white/5 pl-10 text-center text-lg uppercase tracking-widest text-white transition-all focus:border-emerald-500/50 focus:bg-white/10 focus:ring-emerald-500/50"
						bind:value={roomCode}
						onkeydown={handleKeydown}
						maxlength={6}
						disabled={isLoading}
					/>
				</div>

				{#if error}
					<p class="text-sm font-medium text-red-500 animate-in fade-in slide-in-from-top-1">{error}</p>
				{/if}

				<Button
					class="group relative h-14 w-full overflow-hidden rounded-lg bg-white font-bold text-black transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50"
					onclick={joinRoom}
					disabled={isLoading || roomCode.length !== 6}
				>
					<div class="absolute inset-0 flex h-full w-full justify-center [filter:blur(1px)]">
						<div class="absolute inset-0 h-full w-full bg-gradient-to-r from-emerald-500 to-cyan-500 opacity-0 transition-opacity duration-300 group-hover:opacity-20"></div>
					</div>

					<span class="relative flex items-center justify-center gap-2">
						{#if isLoading}
							<Loader2 class="h-5 w-5 animate-spin" />
							Connecting...
						{:else}
							Join Room
						{/if}
					</span>
				</Button>
			</div>
		</div>
	</div>
</div>
