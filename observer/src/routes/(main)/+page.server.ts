import { definePageMetaTags } from 'svelte-meta-tags';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const pageTags = definePageMetaTags({
		title: 'NinjaParty',
		description: 'Observation tool for Ninjabrain Bot',
		canonical: new URL(url.pathname, url.origin).href,
		openGraph: {
			type: 'website',
			url: new URL(url.pathname, url.origin).href,
			title: 'NinjaParty',
			description: 'Observation tool for Ninjabrain Bot',
			images: [
				{
					url: new URL('/og-image.png', url.origin).href,
					alt: 'NinjaParty',
					width: 1200,
					height: 630,
					secureUrl: new URL('/og-image.png', url.origin).href,
					type: 'image/png',
				},
			],
		},
		twitter: {
			cardType: 'summary_large_image',
			title: 'NinjaParty',
			description: 'Observation tool for Ninjabrain Bot',
			image: new URL('/og-image.png', url.origin).href,
			imageAlt: 'NinjaParty',
		},
	});

	return {
		...pageTags,
	};
};
