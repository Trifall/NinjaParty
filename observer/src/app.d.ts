// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			user: DBUser | null;
			isAdmin: boolean;
		}
		interface Platform {
			env: {
				PORTFOLIO_CACHE: KVNamespace;
			};
		}
	}
}

// Type declarations for @sveltejs/enhanced-img
declare module '*?enhanced' {
	const value: string;
	export default value;
}

export {};
