import adapterCloudflare from '@sveltejs/adapter-cloudflare';
import adapterNode from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://svelte.dev/docs/kit/adapters for more information about adapters.
		adapter: process.env.BUILD_TARGET === 'node' ? adapterNode() : adapterCloudflare(),
		alias: {
			$components: './src/lib/components',
			$types: './src/types',
			$src: './src',
			$database: './database',
			'@': './',
		},
		// CSRF protection for cross-origin submissions from third parties
		csrf: {
			trustedOrigins: process.env.CSRF_TRUSTED_ORIGINS
				? process.env.CSRF_TRUSTED_ORIGINS.split(',').map((s) => s.trim()).filter(Boolean)
				: [],
		},
		experimental: {
			remoteFunctions: true,
		},
	},
	compilerOptions: {
		experimental: {
			async: true,
		},
	},
};

export default config;
