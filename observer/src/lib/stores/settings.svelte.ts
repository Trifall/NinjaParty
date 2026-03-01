import { PersistedState } from 'runed';
import type { McVersion } from '$lib/utils/stronghold';

export type CoordinateDisplayType = 'FOUR_FOUR' | 'EIGHT_EIGHT' | 'CHUNK';
export type ViewMode = 'max' | 'compact';

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
	mcVersion = createPersistedSafe<McVersion>('ninjabrain_settings_mc_version', 'PRE_119');
	colorNegatives = createPersistedSafe<boolean>('ninjabrain_settings_color_negatives', false);
	viewMode = createPersistedSafe<ViewMode>('ninjabrain_settings_view_mode', 'max');
}

export const settings = new AppSettings();
