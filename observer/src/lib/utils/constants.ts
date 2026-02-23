/* this data is not available in the API for Ninjabrain v1.5.1 release, so we are estimating it

	formula uses Math.cos(beta * toRad)
		- verticalAngle = roughly -32.11166deg from my limited testing
		- cos(-32.11166deg) = cos(32.11166deg) = ~0.8466
			-> so correction per increment = 0.001874 / 0.8466 = ~0.002213 deg
			-> correctionIncrements = round(correction / 0.002213) = round(correction * 451.9)

	this is a rough approximation, we don't have access to the exact data
	*/
const DERIVED_INCREMENTS_MULTIPLIER = 451.9;
export function getDerivedIncrements(correction: number | undefined): number {
	if (correction === undefined) return 0;
	return Math.round(correction * DERIVED_INCREMENTS_MULTIPLIER);
}

export const BOAT_ALLOWABLE_ERROR = 0.03;
export const BOAT_STD_DEV = 0.0007;
export const NON_BOAT_WARNING_THRESHOLD = 0.05; // guessing this
