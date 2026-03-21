// oxlint-disable triple-slash-reference
/// <reference types="vitest" />
/// <reference types="vite/client" />
import { enhancedImages } from '@sveltejs/enhanced-img';
import { sveltekit } from '@sveltejs/kit/vite';
import { readFileSync } from 'fs';
import { defineConfig } from 'vite';

const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf8'));

export default defineConfig({
	plugins: [enhancedImages(), sveltekit()],
	define: {
		__APP_VERSION__: JSON.stringify(pkg.version),
	},
	build: {
		rollupOptions: {
			output: {
				manualChunks: (id) => {
					// dont chunk SSR-only code on client
					if (id.includes('node_modules')) {
						if (id.includes('highlight.js')) {
							return 'highlight-js';
						}
						if (id.includes('zod')) {
							return 'zod';
						}
						// server-only dependencies should not be in client bundle
						if (id.includes('drizzle-orm')) {
							return undefined; // exclude from client
						}
						// PixiJS and its filters (large graphics library)
						if (id.includes('pixi.js') || id.includes('@pixi/')) {
							return 'pixi';
						}
						// tanstack
						if (id.includes('@tanstack/table')) {
							return '@tanstack/table';
						}
						// lucide icons
						if (id.includes('@lucide/svelte')) {
							return 'lucide-icons';
						}
						// bits-ui
						if (id.includes('bits-ui')) {
							return 'bits-ui';
						}
						// date-fns
						if (id.includes('date-fns')) {
							return 'date-fns';
						}

						if (id.includes('better-auth')) {
							return 'better-auth';
						}
						// svelte-sonner (toast library)
						if (id.includes('svelte-sonner')) {
							return 'svelte-sonner';
						}
						// vaul-svelte (drawer component)
						if (id.includes('vaul-svelte')) {
							return 'vaul-svelte';
						}
						// mode-watcher (theme switching)
						if (id.includes('mode-watcher')) {
							return 'mode-watcher';
						}
						// marked (markdown parser)
						if (id.includes('marked') || id.includes('markedjs')) {
							return 'marked';
						}
						// remaining smaller vendor libs
						return 'vendor';
					}
				},
			},
		},
	},
	optimizeDeps: {
		// exclude server-only dependencies from optimization
		exclude: ['drizzle-orm'],
	},
	ssr: {
		// ensure these are not bundled in client
		noExternal: [],
		external: ['fs-extra', 'winston', 'better-auth/adapters'],
	},
});
