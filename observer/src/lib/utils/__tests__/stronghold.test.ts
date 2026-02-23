import { describe, expect, it } from 'vitest';
import {
	chunkToEightEight,
	chunkToFourFour,
	chunkToNether,
	getAngleDifference,
	getAngleToStructure,
	getDisplayCoordinates,
	getOverworldDistance,
	getStrongholdOffset,
	normalizeMinecraftAngle,
} from '../stronghold';

describe('stronghold utils', () => {
	describe('normalizeMinecraftAngle', () => {
		it('should return angle unchanged if already in range', () => {
			expect(normalizeMinecraftAngle(0)).toBe(0);
			expect(normalizeMinecraftAngle(90)).toBe(90);
			expect(normalizeMinecraftAngle(-90)).toBe(-90);
			expect(normalizeMinecraftAngle(180)).toBe(180);
			expect(normalizeMinecraftAngle(-180)).toBe(-180);
		});

		it('should wrap angles greater than 180', () => {
			expect(normalizeMinecraftAngle(181)).toBe(-179);
			expect(normalizeMinecraftAngle(270)).toBe(-90);
			expect(normalizeMinecraftAngle(360)).toBe(0);
			expect(normalizeMinecraftAngle(450)).toBe(90);
		});

		it('should wrap angles less than -180', () => {
			expect(normalizeMinecraftAngle(-181)).toBe(179);
			expect(normalizeMinecraftAngle(-270)).toBe(90);
			expect(normalizeMinecraftAngle(-360)).toBeCloseTo(0);
			expect(normalizeMinecraftAngle(-450)).toBe(-90);
		});

		it('should handle 360 degree wrap correctly', () => {
			expect(normalizeMinecraftAngle(720)).toBeCloseTo(0);
			expect(normalizeMinecraftAngle(-720)).toBeCloseTo(0);
		});
	});

	describe('getAngleToStructure', () => {
		it('should return 0° when player is directly north of structure', () => {
			const angle = getAngleToStructure(0, -100, 0, 0);
			expect(angle).toBeCloseTo(0, 5);
		});

		it('should return 90° when player is directly west of structure', () => {
			const angle = getAngleToStructure(100, 0, 0, 0);
			expect(angle).toBeCloseTo(90, 5);
		});

		it('should return -180° when player is directly south of structure', () => {
			const angle = getAngleToStructure(0, 100, 0, 0);
			expect(angle).toBeCloseTo(-180, 5);
		});

		it('should return -90° when player is directly east of structure', () => {
			const angle = getAngleToStructure(-100, 0, 0, 0);
			expect(angle).toBeCloseTo(-90, 5);
		});

		it('should handle diagonal angles', () => {
			const angle = getAngleToStructure(-100, -100, 0, 0);
			expect(angle).toBeCloseTo(-45, 5);
		});
	});

	describe('getAngleDifference', () => {
		it('should return 0 when angles match', () => {
			expect(getAngleDifference(90, 90)).toBe(0);
		});

		it('should calculate positive turn for right direction', () => {
			const diff = getAngleDifference(0, -90);
			expect(diff).toBe(90);
		});

		it('should calculate negative turn for left direction', () => {
			const diff = getAngleDifference(0, 90);
			expect(diff).toBe(-90);
		});

		it('should wrap around 180°', () => {
			expect(getAngleDifference(170, -170)).toBe(-20);
			expect(getAngleDifference(-170, 170)).toBe(20);
		});

		it('should handle wrap from 180 to -180', () => {
			expect(getAngleDifference(180, -180)).toBe(0);
		});
	});

	describe('chunkToFourFour', () => {
		it('should convert chunk coordinates to (4,4) offset', () => {
			expect(chunkToFourFour(0, 0)).toEqual([4, 4]);
			expect(chunkToFourFour(1, 1)).toEqual([20, 20]);
			expect(chunkToFourFour(-1, 2)).toEqual([-12, 36]);
		});

		it('should multiply chunk by 16 and add 4', () => {
			expect(chunkToFourFour(10, 5)).toEqual([164, 84]);
		});
	});

	describe('chunkToEightEight', () => {
		it('should convert chunk coordinates to (8,8) offset', () => {
			expect(chunkToEightEight(0, 0)).toEqual([8, 8]);
			expect(chunkToEightEight(1, 1)).toEqual([24, 24]);
			expect(chunkToEightEight(-1, 2)).toEqual([-8, 40]);
		});

		it('should multiply chunk by 16 and add 8', () => {
			expect(chunkToEightEight(10, 5)).toEqual([168, 88]);
		});
	});

	describe('getDisplayCoordinates', () => {
		it('should return FOUR_FOUR coordinates when display type is FOUR_FOUR', () => {
			const result = getDisplayCoordinates(5, 10, 'FOUR_FOUR');
			expect(result).toEqual([84, 164]);
		});

		it('should return EIGHT_EIGHT coordinates when display type is EIGHT_EIGHT', () => {
			const result = getDisplayCoordinates(5, 10, 'EIGHT_EIGHT');
			expect(result).toEqual([88, 168]);
		});

		it('should return raw chunk coordinates when display type is CHUNK', () => {
			const result = getDisplayCoordinates(5, 10, 'CHUNK');
			expect(result).toEqual([5, 10]);
		});
	});

	describe('chunkToNether', () => {
		it('should multiply chunk coordinates by 2', () => {
			expect(chunkToNether(0, 0)).toEqual([0, 0]);
			expect(chunkToNether(1, 1)).toEqual([2, 2]);
			expect(chunkToNether(10, 20)).toEqual([20, 40]);
			expect(chunkToNether(-5, 15)).toEqual([-10, 30]);
		});
	});

	describe('getStrongholdOffset', () => {
		it('should return 8 for PRE_119 version', () => {
			expect(getStrongholdOffset('PRE_119')).toBe(8);
		});

		it('should return 0 for POST_119 version', () => {
			expect(getStrongholdOffset('POST_119')).toBe(0);
		});
	});

	describe('getOverworldDistance', () => {
		it('should calculate distance correctly for PRE_119', () => {
			const distance = getOverworldDistance(0, 0, 8, 8, 'PRE_119');
			expect(distance).toBeCloseTo(0, 5);
		});

		it('should calculate distance correctly for POST_119', () => {
			const distance = getOverworldDistance(0, 0, 0, 0, 'POST_119');
			expect(distance).toBeCloseTo(0, 5);
		});

		it('should calculate distance between player and stronghold', () => {
			const distance = getOverworldDistance(0, 0, 100, 100, 'PRE_119');
			const expected = Math.sqrt((8 - 100) ** 2 + (8 - 100) ** 2);
			expect(distance).toBeCloseTo(expected, 5);
		});

		it('should account for version offset in distance', () => {
			const pre119Dist = getOverworldDistance(0, 0, 8, 8, 'PRE_119');
			const post119Dist = getOverworldDistance(0, 0, 0, 0, 'POST_119');
			expect(pre119Dist).toBeCloseTo(0, 5);
			expect(post119Dist).toBeCloseTo(0, 5);
		});
	});
});
