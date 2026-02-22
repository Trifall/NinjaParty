/**
 * Redirect configuration - similar to Next.js redirects() in next.config.mjs
 * Define source paths and their destinations here
 */

export interface RedirectConfig {
	source: string;
	destination: string;
	permanent: boolean;
}

/**
 * Array of redirect configurations
 * - source: the incoming request path
 * - destination: where to redirect (can be external URL or internal path)
 * - permanent: true = 308 (permanent), false = 307 (temporary)
 */
export const redirects: RedirectConfig[] = [
	{
		source: '/utils',
		destination: 'https://utils.trifall.com',
		permanent: true,
	},
	{
		source: '/utilities',
		destination: 'https://utils.trifall.com',
		permanent: true,
	},
	{
		source: '/cosmic',
		destination: 'https://github.com/Trifall/cosmic',
		permanent: true,
	},
];

/**
 * Check if a pathname matches any redirect and return the redirect config
 */
export function getRedirect(pathname: string): RedirectConfig | undefined {
	return redirects.find((redirect) => redirect.source === pathname);
}
