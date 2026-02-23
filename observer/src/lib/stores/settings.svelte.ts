import type { McVersion } from '$lib/utils/stronghold';

export type CoordinateDisplayType = 'FOUR_FOUR' | 'EIGHT_EIGHT' | 'CHUNK';

class AppSettings {
	#displayType = $state<CoordinateDisplayType>('FOUR_FOUR');
	#mcVersion = $state<McVersion>('PRE_119');
	#colorNegatives = $state<boolean>(false);

	constructor() {
		if (typeof window !== 'undefined') {
			const savedDisplay = localStorage.getItem('ninjabrain_settings_display_type');
			if (
				savedDisplay === 'FOUR_FOUR' ||
				savedDisplay === 'EIGHT_EIGHT' ||
				savedDisplay === 'CHUNK'
			) {
				this.#displayType = savedDisplay;
			}

			const savedVersion = localStorage.getItem('ninjabrain_settings_mc_version');
			if (savedVersion === 'PRE_119' || savedVersion === 'POST_119') {
				this.#mcVersion = savedVersion;
			}

			const savedColorNegatives = localStorage.getItem('ninjabrain_settings_color_negatives');
			if (savedColorNegatives === 'true') {
				this.#colorNegatives = true;
			}
		}
	}

	get displayType(): CoordinateDisplayType {
		return this.#displayType;
	}

	set displayType(value: CoordinateDisplayType) {
		this.#displayType = value;
		if (typeof window !== 'undefined') {
			localStorage.setItem('ninjabrain_settings_display_type', value);
		}
	}

	get mcVersion(): McVersion {
		return this.#mcVersion;
	}

	set mcVersion(value: McVersion) {
		this.#mcVersion = value;
		if (typeof window !== 'undefined') {
			localStorage.setItem('ninjabrain_settings_mc_version', value);
		}
	}

	get colorNegatives(): boolean {
		return this.#colorNegatives;
	}

	set colorNegatives(value: boolean) {
		this.#colorNegatives = value;
		if (typeof window !== 'undefined') {
			localStorage.setItem('ninjabrain_settings_color_negatives', String(value));
		}
	}
}

export const settings = new AppSettings();
