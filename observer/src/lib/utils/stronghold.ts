import type { CoordinateDisplayType } from '$lib/stores/settings.svelte';

export type McVersion = 'PRE_119' | 'POST_119';

// ==================== COLOR UTILITIES (matching Ninjabrain Bot) ====================

/*
	- Color Mapping: https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/gui/style/theme/ColorMap.java
	- Themes: https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/gui/style/theme/CustomTheme.java
*/

// Correction increments color
export function getCorrectionColor(increments: number): string {
	return increments >= 0 ? '#75CC6C' : '#CC6E72';
}

// Certainty color with gradient interpolation (matches Ninjabrain Bot)
export function getCertaintyColor(certainty: number): string {
	const t = Math.max(0, Math.min(1, certainty));

	// Color stops matching Ninjabrain Bot gradient
	const stops = [
		{ pos: 0, r: 255, g: 0, b: 0 }, // #FF0000 - Red (0-10%)
		{ pos: 0.25, r: 255, g: 165, b: 0 }, // #FFA500 - Orange (20-30%)
		{ pos: 0.5, r: 255, g: 255, b: 0 }, // #FFFF00 - Yellow (40-60%)
		{ pos: 0.75, r: 144, g: 238, b: 144 }, // #90EE90 - Light green (70-80%)
		{ pos: 1, r: 0, g: 206, b: 41 }, // #00CE29 - Green (90-100%)
	];

	return interpolateColorStops(t, stops);
}

/*
	Travel angle color: https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/model/datastate/common/StructureInformation.java:70-72 - getTravelAngleDiffColor() returns 1 - |angleDiff| / 180 (same formula as your TypeScript)
*/

// Angle difference color (for travel direction)
// color = 1 - |angleDiff| / 180
// 0° diff = green, 90° diff = yellow, 180° diff = red
export function getAngleDiffColor(angleDiff: number): string {
	const absDiff = Math.abs(angleDiff);
	const t = 1 - Math.min(absDiff / 180, 1);

	const stops = [
		{ pos: 0, r: 255, g: 0, b: 0 }, // #FF0000 - Red (180° off)
		{ pos: 0.5, r: 255, g: 255, b: 0 }, // #FFFF00 - Yellow (90° off)
		{ pos: 1, r: 0, g: 206, b: 41 }, // #00CE29 - Green (0° off)
	];

	return interpolateColorStops(t, stops);
}

function interpolateColorStops(
	t: number,
	stops: { pos: number; r: number; g: number; b: number }[]
): string {
	// Find the two stops to interpolate between
	let lower = stops[0];
	let upper = stops[1];

	for (let i = 1; i < stops.length; i++) {
		if (t <= stops[i].pos) {
			lower = stops[i - 1];
			upper = stops[i];
			break;
		}
	}

	// Calculate local t within the segment
	const segmentSize = upper.pos - lower.pos;
	const localT = segmentSize > 0 ? (t - lower.pos) / segmentSize : 0;

	const r = Math.round(lower.r + (upper.r - lower.r) * localT);
	const g = Math.round(lower.g + (upper.g - lower.g) * localT);
	const b = Math.round(lower.b + (upper.b - lower.b) * localT);

	return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`;
}

// ==================== ANGLE UTILITIES ====================

/*
Angle calculations:
- https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/util/Coords.java - getPhi(x, z) = -Math.atan2(x, z) - returns angle in radians
- https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/model/datastate/common/StructureInformation.java:42-47 - angle to structure and angle difference with wrapping (lines 42-47)
- https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/model/datastate/stronghold/Chunk.java:105-108 - getAngleError() uses -180 / Math.PI * Math.atan2(deltaX, deltaZ)
*/

// Minecraft angles: 0 -> 180 -> wraps to -180 -> 0
// Normalizes to -180 to 180 range
export function normalizeMinecraftAngle(angle: number): number {
	let normalized = angle % 360;
	if (normalized > 180) normalized -= 360;
	if (normalized < -180) normalized += 360;
	return normalized;
}

// Calculate angle to structure from player position
export function getAngleToStructure(
	playerX: number,
	playerZ: number,
	structureX: number,
	structureZ: number
): number {
	const xDiff = structureX - playerX;
	const zDiff = structureZ - playerZ;
	return (-Math.atan2(xDiff, zDiff) * 180) / Math.PI;
}

// Calculate angle difference with wrapping (for turn direction)
export function getAngleDifference(angleToStructure: number, playerAngle: number): number {
	let diff = (angleToStructure - playerAngle) % 360;
	if (diff > 180) diff -= 360;
	if (diff < -180) diff += 360;
	return diff;
}

/*
Coordinate conversions:
- https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/model/datastate/stronghold/Chunk.java:34-56:
  - fourFourX()/fourFourZ() = 16 * x/z + 4
  - eightEightX()/eightEightZ() = 16 * x/z + 8
  - netherX()/netherZ() = 2 * x/z
Version-dependent offset:
- https://github.com/Ninjabrain1/Ninjabrain-Bot/blob/a8d5d29c2f39cc4d1f8f1fd88cddbca8ea05cce1/src/main/java/ninjabrainbot/model/datastate/stronghold/StrongholdConstants.java:30-33 - getStrongholdChunkCoord() returns 8 for PRE119, 0 for POST119

*/

// ==================== OVERWORLD COORDINATES ====================

// (4,4) - Starter staircase location (most common display)
export function chunkToFourFour(chunkX: number, chunkZ: number): [number, number] {
	return [16 * chunkX + 4, 16 * chunkZ + 4];
}

// (8,8) - Center of stronghold chunk
export function chunkToEightEight(chunkX: number, chunkZ: number): [number, number] {
	return [16 * chunkX + 8, 16 * chunkZ + 8];
}

export function getDisplayCoordinates(
	chunkX: number,
	chunkZ: number,
	displayType: CoordinateDisplayType
): [number, number] {
	if (displayType === 'FOUR_FOUR') return chunkToFourFour(chunkX, chunkZ);
	if (displayType === 'EIGHT_EIGHT') return chunkToEightEight(chunkX, chunkZ);
	return [chunkX, chunkZ];
}

// ==================== NETHER COORDINATES ====================
// Nether coords (displayed in UI)
export function chunkToNether(chunkX: number, chunkZ: number): [number, number] {
	return [2 * chunkX, 2 * chunkZ];
}

// ==================== INTERNAL CALCULATIONS ====================
// Version-dependent stronghold chunk offset
export function getStrongholdOffset(version: McVersion): number {
	return version === 'PRE_119' ? 8 : 0;
}

// Overworld distance from player to stronghold (blocks)
export function getOverworldDistance(
	chunkX: number,
	chunkZ: number,
	playerX: number,
	playerZ: number,
	version: McVersion
): number {
	const offset = getStrongholdOffset(version);
	const strongX = 16 * chunkX + offset;
	const strongZ = 16 * chunkZ + offset;
	return Math.sqrt((strongX - playerX) ** 2 + (strongZ - playerZ) ** 2);
}
