<script lang="ts">
	import { page } from '$app/state';
	import Dashboard from '$lib/components/dashboard/Dashboard.svelte';
	import { RoomConnection } from '$lib/stores/room.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let isEmbed = page.url.searchParams.get('embed') === '1';
	// svelte-ignore state_referenced_locally
	let roomStore = $state(new RoomConnection(data.room, isEmbed));

	$effect(() => {
		roomStore.connect();
		return () => {
			roomStore.disconnect();
		};
	});
</script>

<svelte:head>
	<title>Room {data.room.toUpperCase()} - NinjaParty</title>
</svelte:head>

<div class="h-full min-h-[calc(100vh-64px)] w-full p-4 lg:p-8">
	<Dashboard {roomStore} roomCode={data.room} {isEmbed} />
</div>
