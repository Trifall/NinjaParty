import { env } from '$env/dynamic/public';

const isSecure = env.PUBLIC_USE_SECURE_URLS !== 'false';
const wsProtocol = isSecure ? 'wss' : 'ws';
const httpProtocol = isSecure ? 'https' : 'http';

export const WS_URL = `${wsProtocol}://${env.PUBLIC_SERVER_URL}/ws`;
export const API_URL = `${httpProtocol}://${env.PUBLIC_SERVER_URL}/api/v1`;
