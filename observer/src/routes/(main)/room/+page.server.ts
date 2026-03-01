import { redirect } from '@sveltejs/kit';

export const load = async () => {
	// redirect to root
	throw redirect(302, '/');
};
