import { defineBaseMetaTags } from 'svelte-meta-tags';
import { siteConfig } from '$lib/data/site';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ url }) => {
	const siteName = siteConfig.name;
	const siteDescription = siteConfig.description;
	const baseTags = defineBaseMetaTags({
		title: siteName,
		titleTemplate: '%s | ' + siteName,
		description: siteDescription,
		canonical: new URL(url.pathname, url.origin).href,
		openGraph: {
			type: 'website',
			url: new URL(url.pathname, url.origin).href,
			title: siteName,
			description: siteDescription,
			images: [
				{
					url: new URL('/og-image.png', url.origin).href,
					alt: 'Open Graph image for ' + siteName,
					width: 1200,
					height: 630,
					secureUrl: new URL('/og-image.png', url.origin).href,
					type: 'image/png',
				},
			],
		},
		twitter: {
			cardType: 'summary_large_image',
			title: siteName,
			description: siteDescription,
			image: new URL('/og-image.png', url.origin).href,
			imageAlt: 'Twitter image for ' + siteName,
		},
	});

	return baseTags;
};
