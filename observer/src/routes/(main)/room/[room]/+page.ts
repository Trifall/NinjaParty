import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) => {
	const roomCode = params.room.replace(/[^a-zA-Z0-9]/g, '').toLowerCase();

	if (!roomCode || roomCode.length !== 6) {
		error(400, 'Invalid room code');
	}

	return {
		room: roomCode,
	};
};
