import { PUBLIC_SERVER_URL, PUBLIC_USE_SECURE_URLS } from '$env/static/public';

const isSecure = PUBLIC_USE_SECURE_URLS !== 'false';
const wsProtocol = isSecure ? 'wss' : 'ws';
const httpProtocol = isSecure ? 'https' : 'http';

export const WS_URL = `${wsProtocol}://${PUBLIC_SERVER_URL}/ws`;
export const API_URL = `${httpProtocol}://${PUBLIC_SERVER_URL}/api/v1`;
