import { env } from '$env/dynamic/public';

export const getPublicSiteName = () => {
	return env.PUBLIC_SITE_NAME && env.PUBLIC_SITE_NAME.length > 0
		? env.PUBLIC_SITE_NAME
		: 'svelte-starter';
};

/**
 * Formats the server uptime
 * @param ms - Uptime in milliseconds
 * @returns formatted uptime string
 */
export const formatUptime = (ms: number): string => {
	const seconds = Math.floor(ms / 1000);
	const days = Math.floor(seconds / (24 * 60 * 60));
	const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
	const minutes = Math.floor((seconds % (60 * 60)) / 60);
	const remainingSeconds = Math.floor(seconds % 60);

	let formatted = '';
	if (days > 0) {
		formatted += `${days}d `;
	}
	if (hours > 0 || days > 0) {
		formatted += `${hours}h `;
	}
	if (minutes > 0 || hours > 0 || days > 0) {
		formatted += `${minutes}m `;
	}
	formatted += `${remainingSeconds}s`;

	return formatted;
};

/**
 * Formats the JSON data for display
 * @param json - JSON data
 * @returns formatted string
 */
export const formatJSON = (json: unknown): string => {
	if (!json) return 'No data';
	try {
		// add proper indentation for JSON display and remove brackets
		const jsonString = JSON.stringify(json, null, 2);
		// remove first and last lines (opening and closing brackets)
		const lines = jsonString.split('\n');
		// if its just brackets, return the content without brackets
		if (lines.length <= 2) {
			return jsonString.slice(1, -1).trim();
		}
		// for multi-line JSON, remove first and last lines and adjust indentation
		const contentLines = lines.slice(1, -1);
		// remove 2 spaces of indentation from each line
		return contentLines.map((line) => line.substring(2)).join('\n');
	} catch (e: unknown) {
		console.error('Error formatting JSON:', e);
		return 'Invalid JSON data';
	}
};

/**
 * Formats the date for display
 * @param date - Date string or Date object
 * @returns formatted date string in long format (e.g. January 1, 2025)
 */
export const formatFormalDate = (date: Date | string): string => {
	return new Date(date).toLocaleDateString('en-US', {
		year: 'numeric',
		month: 'long',
		day: 'numeric',
	});
};

/*
 * Formats the date relative to now
 * @param date - Date string or Date object
 * @returns formatted date string relative to now (e.g. 2 days ago, yesterday, today, etc.)
 */
export const formatRelativeTime = (date: Date | string | null): string => {
	if (!date) return 'Never';
	const now = new Date();
	const target = new Date(date);

	// handle invalid dates
	if (isNaN(target.getTime())) return 'Unknown';

	const diffInMs = now.getTime() - target.getTime();
	const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

	// handle future dates (clock sync issues, timezone problems, etc.)
	if (diffInDays < 0) return 'Today';

	if (diffInDays === 0) return 'Today';
	if (diffInDays === 1) return 'Yesterday';
	if (diffInDays < 7) return `${diffInDays} days ago`;
	if (diffInDays < 30) return `${Math.floor(diffInDays / 7)} weeks ago`;
	if (diffInDays < 365) return `${Math.floor(diffInDays / 30)} months ago`;
	return `${Math.floor(diffInDays / 365)} years ago`;
};

/**
 * Formats the date for display
 * @param date - Date string or Date object
 * @returns formatted date string in MM/DD/YYYY format
 */
export const formatDate = (date: string | Date) => {
	if (!date) return 'N/A';
	// create a date object, ensuring proper handling of date strings
	const dateObj = new Date(date);
	// force UTC interpretation to avoid timezone issues
	dateObj.setTime(dateObj.getTime() + dateObj.getTimezoneOffset() * 60 * 1000);
	// format as M/D/YYYY to match test expectations
	const month = dateObj.getMonth() + 1; // getMonth() is zero-indexed
	const day = dateObj.getDate();
	const year = dateObj.getFullYear();
	return `${month}/${day}/${year}`;
};
