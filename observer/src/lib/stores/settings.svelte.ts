import { PersistedState } from 'runed';

export type CoordinateDisplayType = 'FOUR_FOUR' | 'EIGHT_EIGHT' | 'CHUNK';

export const VIEW_MODE = {
	MAX: 'max',
	COMPACT: 'compact',
	ULTRA_COMPACT: 'ultra-compact',
	GRID: 'grid',
} as const;

export type ViewMode = (typeof VIEW_MODE)[keyof typeof VIEW_MODE];

function createPersistedSafe<T>(key: string, defaultValue: T): PersistedState<T> {
	try {
		return new PersistedState<T>(key, defaultValue);
	} catch {
		localStorage.removeItem(key);
		return new PersistedState<T>(key, defaultValue);
	}
}

class AppSettings {
	displayType = createPersistedSafe<CoordinateDisplayType>(
		'ninjabrain_settings_display_type',
		'FOUR_FOUR'
	);
	colorNegatives = createPersistedSafe<boolean>('ninjabrain_settings_color_negatives', true);
	viewMode = createPersistedSafe<ViewMode>('ninjabrain_settings_view_mode', VIEW_MODE.MAX);
	// 0 = disabled, otherwise seconds of inactivity before embed fades out
	embedAutoHide = createPersistedSafe<number>('ninjabrain_settings_embed_auto_hide', 0);
	// when true offline runners sink to bottom regardless of custom drag order
	reorderOfflineRunners = createPersistedSafe<boolean>(
		'ninjabrain_settings_reorder_offline_runners',
		false
	);
}

export const settings = new AppSettings();
