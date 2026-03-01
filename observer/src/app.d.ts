// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Locals {
		// }
		/* 		interface Platform {
			env: {
			};
		}
 */
	}
}

// Type declarations for @sveltejs/enhanced-img
declare module '*?enhanced' {
	const value: string;
	export default value;
}

export {};
