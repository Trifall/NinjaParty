export const ROUTES = {
	HOME: '/',
};

export type Route = (typeof ROUTES)[keyof typeof ROUTES];
