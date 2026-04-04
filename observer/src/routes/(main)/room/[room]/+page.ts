import { error } from '@sveltejs/kit';
import { API_URL } from '$lib/constants';
import type { PageLoad } from './$types';

export const ssr = false;

export const load: PageLoad = async ({ params, fetch }) => {
	const roomCode = params.room.replace(/[^a-zA-Z0-9]/g, '').toLowerCase();

	if (!roomCode || roomCode.length !== 6) {
		error(400, 'Invalid room code');
	}

	const res = await fetch(`${API_URL}/rooms/${roomCode}/exists`);
	if (!res.ok) {
		error(404, 'Room not found');
	}

	return {
		room: roomCode,
	};
};
