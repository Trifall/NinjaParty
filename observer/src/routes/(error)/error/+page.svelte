<script lang="ts">
	import { House } from '@lucide/svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { Pathname } from '$app/types';
	import { ROUTES } from '$src/lib/routes.js';
	import { getPublicSiteName } from '$src/lib/utils/format';
	import { Button } from '$components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$components/ui/card';

	// get error message from the page error or URL query parameters
	let error =
		page.url.searchParams.get('message') ||
		page.url.searchParams.get('error') ||
		page.error?.message ||
		'An unknown error occurred';
</script>

<svelte:head>
	<title>Error | {getPublicSiteName()}</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Error</CardTitle>
			{#if !error}
				<CardDescription>Something went wrong</CardDescription>
			{/if}
		</CardHeader>
		<CardContent>
			<Button
				href={resolve(ROUTES.HOME as Pathname)}
				variant="outline"
				class="flex items-center gap-2"
			>
				<House class="h-4 w-4" />
				Back to Home
			</Button>
		</CardContent>
	</Card>
</div>
