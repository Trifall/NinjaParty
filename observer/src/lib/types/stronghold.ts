// ==================== ENUMS ====================
export type ResultType =
	| 'NONE'
	| 'TRIANGULATION'
	| 'FAILED'
	| 'BLIND'
	| 'DIVINE'
	| 'ALL_ADVANCEMENTS';

export type EnderEyeThrowType = 'NORMAL' | 'NORMAL_WITH_ALT_STD' | 'MANUAL' | 'BOAT';

export type BoatState = 'NONE' | 'ERROR' | 'MEASURING' | 'VALID';

export type BlindEvaluation =
	| 'EXCELLENT'
	| 'HIGHROLL_GOOD'
	| 'HIGHROLL_OKAY'
	| 'BAD_BUT_IN_RING'
	| 'BAD'
	| 'NOT_IN_RING';

export type InformationMessageSeverity = 'INFO' | 'WARNING' | 'ERROR';

export type InformationMessageType =
	| 'PORTAL_LINKING'
	| 'NEXT_THROW_DIRECTION'
	| 'MISMEASURE'
	| 'MC_VERSION'
	| 'COMBINED_CERTAINTY';

// ==================== SHARED TYPES ====================
export interface PlayerPosition {
	xInOverworld?: number;
	zInOverworld?: number;
	isInOverworld?: boolean;
	isInNether?: boolean;
	horizontalAngle?: number;
}

export interface EyeThrow {
	xInOverworld: number;
	zInOverworld: number;
	angle: number;
	angleWithoutCorrection: number;
	correction: number;
	correctionIncrements: number;
	error: number;
	type: EnderEyeThrowType;
}

export interface ChunkPrediction {
	certainty: number;
	chunkX: number;
	chunkZ: number;
	overworldDistance: number;
}

export interface StructurePosition {
	xInOverworld?: number;
	zInOverworld?: number;
	overworldDistance?: number;
	travelAngle?: number;
}

export interface BlindResult {
	xInNether: number;
	zInNether: number;
	highrollProbability: number;
	highrollThreshold: number;
	averageDistance: number;
	improveDirection: number;
	improveDistance: number;
	evaluation: BlindEvaluation;
}

export interface DivineResult {
	fossilXCoordinate: number;
	formattedHighrollCoords: string;
	formattedSafeCoords: string;
}

export interface InformationMessage {
	severity: InformationMessageSeverity;
	type: InformationMessageType;
	message: string;
}

// ==================== APPS ====================
export type RoomMessageSource = 'ninjabrain-local' | 'server';

export type RoomMessageType =
	| 'stronghold_update'
	| 'presence'
	| 'boat_update'
	| 'blind_update'
	| 'divine_update'
	| 'all_advancements_update'
	| 'information_messages_update';

export interface RoomMessage {
	type: RoomMessageType;
	ts: number;
	source: RoomMessageSource;
	payload: any;
}

export interface PresencePayload {
	status: 'online' | 'offline';
}

export interface StrongholdPayload {
	resultType: ResultType;
	predictions?: ChunkPrediction[];
	eyeThrows?: EyeThrow[];
	playerPosition?: PlayerPosition;
}

export interface BoatPayload {
	boatAngle: number | null;
	boatState: BoatState;
}

export interface BlindPayload {
	isBlindModeEnabled: boolean;
	hasDivine: boolean;
	blindResult: BlindResult | {};
}

export interface DivinePayload {
	isDivineModeEnabled: boolean;
	divineResult: DivineResult | {};
}

export interface AllAdvancementsPayload {
	isAllAdvancementsModeEnabled: boolean;
	spawn: StructurePosition | {};
	outpost: StructurePosition | {};
	monument: StructurePosition | {};
	stronghold: StructurePosition | {};
	deepDark: StructurePosition | {};
	cityQuery: StructurePosition | {};
	shulkerTransport: StructurePosition | {};
	generalLocation: StructurePosition | {};
}

export interface InformationMessagesPayload {
	informationMessages?: InformationMessage[];
}
