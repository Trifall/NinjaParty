import { definePageMetaTags } from 'svelte-meta-tags';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const pageTags = definePageMetaTags({
		title: 'Ninjabrain Party Relay',
		description: 'Real-time stronghold triangulation for Minecraft speedrunners',
		canonical: new URL(url.pathname, url.origin).href,
		openGraph: {
			type: 'website',
			url: new URL(url.pathname, url.origin).href,
			title: 'Ninjabrain Party Relay',
			description: 'Real-time stronghold triangulation for Minecraft speedrunners',
			images: [
				{
					url: new URL('/og-image.png', url.origin).href,
					alt: 'Ninjabrain Party Relay',
					width: 1200,
					height: 630,
					secureUrl: new URL('/og-image.png', url.origin).href,
					type: 'image/png',
				},
			],
		},
		twitter: {
			cardType: 'summary_large_image',
			title: 'Ninjabrain Party Relay',
			description: 'Real-time stronghold triangulation for Minecraft speedrunners',
			image: new URL('/og-image.png', url.origin).href,
			imageAlt: 'Ninjabrain Party Relay',
		},
	});

	return {
		...pageTags,
	};
};
