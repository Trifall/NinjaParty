import type { RunnerState } from './room.svelte';

const RUNNER_ORDER_KEY = (code: string) => `ninjabrain_runner_order_${code}`;
const roomPreferencesCache = new Map<string, RoomPreferences>();

export interface RunnerMove {
	name: string;
	position: number;
}

function sanitizeStoredOrder(order: readonly string[]): string[] {
	const seen = new Set<string>();
	const sanitized: string[] = [];

	for (const rawName of order) {
		const name = rawName.trim();
		if (!name || seen.has(name)) continue;
		seen.add(name);
		sanitized.push(name);
	}

	return sanitized;
}

export function parseRunnerMoves(raw: string): RunnerMove[] {
	if (!raw) return [];

	const seen = new Set<string>();
	const moves: RunnerMove[] = [];

	for (const token of raw.split(',')) {
		const separator = token.lastIndexOf(':');
		if (separator <= 0) continue;

		const encodedName = token.slice(0, separator);
		const positionValue = Number.parseInt(token.slice(separator + 1), 10);

		let name = '';
		try {
			name = decodeURIComponent(encodedName).trim();
		} catch {
			continue;
		}

		if (!name || !Number.isInteger(positionValue) || seen.has(name)) continue;

		seen.add(name);
		moves.push({ name, position: positionValue });
	}

	return moves;
}

export function applyRunnerMoves(
	defaultOrder: readonly string[],
	moves: readonly RunnerMove[]
): string[] {
	const orderedNames = sanitizeStoredOrder(defaultOrder);
	const knownNames = new Set(orderedNames);

	for (const move of moves) {
		if (!knownNames.has(move.name)) continue;

		const currentIndex = orderedNames.indexOf(move.name);
		if (currentIndex === -1) continue;

		const [name] = orderedNames.splice(currentIndex, 1);
		const targetIndex = Math.min(Math.max(move.position - 1, 0), orderedNames.length);
		orderedNames.splice(targetIndex, 0, name);
	}

	return orderedNames;
}

export function serializeRunnerMoves(
	defaultOrder: readonly string[],
	targetOrder: readonly string[]
): string {
	const baseline = sanitizeStoredOrder(defaultOrder);
	const baselineSet = new Set(baseline);
	const desired = mergeRunnerOrder(
		sanitizeStoredOrder(targetOrder).filter((name) => baselineSet.has(name)),
		baseline
	);
	const working = [...baseline];
	const moves: string[] = [];

	for (let targetIndex = 0; targetIndex < desired.length; targetIndex += 1) {
		const name = desired[targetIndex];
		if (working[targetIndex] === name) continue;

		const currentIndex = working.indexOf(name);
		if (currentIndex === -1) continue;

		working.splice(currentIndex, 1);
		working.splice(targetIndex, 0, name);
		moves.push(`${encodeURIComponent(name)}:${targetIndex + 1}`);
	}

	return moves.join(',');
}

export function getRoomPreferences(roomCode: string): RoomPreferences {
	let prefs = roomPreferencesCache.get(roomCode);
	if (!prefs) {
		prefs = new RoomPreferences(roomCode);
		roomPreferencesCache.set(roomCode, prefs);
	}
	return prefs;
}

export class RoomPreferences {
	private roomCode: string;

	// ordered list of runner names (top-to-bottom display order)
	private _order: string[] = $state([]);

	constructor(roomCode: string) {
		this.roomCode = roomCode;
		try {
			const raw = localStorage.getItem(RUNNER_ORDER_KEY(roomCode));
			if (raw) {
				const parsed = JSON.parse(raw);
				if (Array.isArray(parsed)) {
					this._order = sanitizeStoredOrder(parsed.filter((value) => typeof value === 'string'));
				}
			}
		} catch {
			// ignore corrupt data
		}
	}

	get order(): string[] {
		return this._order;
	}

	set order(newOrder: string[]) {
		this._order = sanitizeStoredOrder(newOrder);
		localStorage.setItem(RUNNER_ORDER_KEY(this.roomCode), JSON.stringify(this._order));
	}
}

export function mergeRunnerOrder(
	storedOrder: readonly string[],
	currentRunnerNames: readonly string[]
): string[] {
	// keep order as a full identity list instead of sparse "runner -> slot" assignments that avoids overlap rules and makes malformed URLs easy to normalize
	const normalizedStored = sanitizeStoredOrder(storedOrder);
	const merged = [...normalizedStored];
	const seen = new Set(normalizedStored);

	for (const name of currentRunnerNames) {
		if (!seen.has(name)) {
			merged.push(name);
			seen.add(name);
		}
	}

	return merged;
}

function defaultSort(runners: RunnerState[], joinOrder: string[]): RunnerState[] {
	return [...runners].sort((a, b) => {
		if (a.status !== b.status) return a.status === 'online' ? -1 : 1;
		const aIdx = joinOrder.indexOf(a.name);
		const bIdx = joinOrder.indexOf(b.name);
		if (aIdx !== bIdx) return aIdx - bIdx;
		return a.name.localeCompare(b.name);
	});
}

export function computeOrderedRunners(
	runners: RunnerState[],
	orderList: string[],
	reorderOffline: boolean,
	joinOrder: string[]
): RunnerState[] {
	if (runners.length === 0) return [];

	const defaultOrdered = defaultSort(runners, joinOrder);
	const visibleRunnerNames = new Set(defaultOrdered.map((runner) => runner.name));
	const normalizedVisibleOrder = mergeRunnerOrder(
		sanitizeStoredOrder(orderList).filter((name) => visibleRunnerNames.has(name)),
		defaultOrdered.map((runner) => runner.name)
	);
	const runnersByName = new Map(defaultOrdered.map((runner) => [runner.name, runner]));
	const ordered = normalizedVisibleOrder
		.map((name) => runnersByName.get(name))
		.filter((runner): runner is RunnerState => runner !== undefined);

	if (reorderOffline) {
		const online = ordered.filter((runner) => runner.status === 'online');
		const offline = ordered.filter((runner) => runner.status === 'offline');
		return [...online, ...offline];
	}

	return ordered;
}
