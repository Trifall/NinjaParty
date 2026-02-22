import { createChildLogger } from '@/src/lib/server/logger';
import type { Handle, HandleValidationError } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { building } from '$app/environment';
import { getRedirect } from '$lib/data/redirects';

const logger = createChildLogger('Hooks');

// server startup tasks (only when not building)
if (!building) {
	const isBun = typeof Bun !== 'undefined' || !!process.versions.bun;
	if (isBun) {
		logger.debug('Bun detected!');
	}
}

export const handle: Handle = sequence(
	// protocol fix and debug logging for reverse proxy
	async ({ event, resolve }) => {
		const proto = event.request.headers.get('x-forwarded-proto');
		if (proto === 'https') {
			event.url.protocol = 'https:';
		}

		return resolve(event);
	},
	// redirect middleware - check for configured redirects
	async ({ event, resolve }) => {
		// skip during build process
		if (building) {
			return resolve(event);
		}

		const url = new URL(event.request.url);

		// check if there's a redirect for this pathname
		const redirect = getRedirect(url.pathname);
		if (redirect) {
			const status = redirect.permanent ? 308 : 307;
			const destination = redirect.destination + url.search;
			logger.debug(`Redirecting ${url.pathname} to ${destination} (${status})`);
			return new Response(null, {
				status,
				headers: {
					Location: destination,
				},
			});
		}

		return resolve(event);
	}
);

export const handleValidationError: HandleValidationError = () => {
	return {
		message: 'Bad Request',
	};
};
