<script lang="ts">
	import type { PageData } from './$types';
	import { RoomConnection } from '$lib/stores/room.svelte';
	import Dashboard from '$lib/components/dashboard/Dashboard.svelte';

	let { data }: { data: PageData } = $props();

	// svelte-ignore state_referenced_locally
	let roomStore = $state(new RoomConnection(data.room));

	$effect(() => {
		roomStore.connect();
		return () => {
			roomStore.disconnect();
		};
	});
</script>

<svelte:head>
	<title>Room {data.room.toUpperCase()} - Ninjabrain Party</title>
</svelte:head>

<div class="h-full min-h-[calc(100vh-64px)] w-full p-4 lg:p-8">
	<Dashboard {roomStore} roomCode={data.room} />
</div>
