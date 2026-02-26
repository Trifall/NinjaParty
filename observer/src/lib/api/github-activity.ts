/// <reference types="@cloudflare/workers-types" />

// Types from Culpa API
export interface ProcessedCommit {
	repo: string;
	message: string;
	href: string;
	sha: string;
	date: string;
	additions: number;
	deletions: number;
}

export interface WorkerResponse {
	commits: ProcessedCommit[];
	stats: {
		totalAdditions: number;
		totalDeletions: number;
		totalCommits: number;
	};
}

interface CacheEntry {
	data: WorkerResponse;
	timestamp: number;
}

// Configuration
const CULPA_API_URL = 'https://culpa.trifall.com/v1/latest?username=Trifall&limit=3';
const KV_KEY = 'github:commits:latest';
const CACHE_TTL = 3600 * 1000; // 1 hour in milliseconds
const STALE_THRESHOLD = 900 * 1000; // 15 minutes in milliseconds

/**
 * Fetches the latest GitHub commits from Culpa API with KV caching.
 * Implements stale-while-revalidate pattern for optimal UX.
 * Falls back to direct API call if KV is not available (e.g., in dev).
 */
export async function fetchLatestCommits(
	kv: KVNamespace | undefined
): Promise<WorkerResponse | null> {
	try {
		// If no KV (e.g., local dev), fetch directly without caching
		if (!kv) {
			return await fetchFromAPI();
		}

		// Check KV cache first
		const cached = await kv.get<CacheEntry>(KV_KEY, 'json');
		const now = Date.now();

		if (cached) {
			const age = now - cached.timestamp;

			// Fresh cache (< 15 min): return immediately
			if (age < STALE_THRESHOLD) {
				return cached.data;
			}

			// Stale cache (15 min - 1 hour): return stale data, trigger background refresh
			if (age < CACHE_TTL) {
				// Trigger background refresh without awaiting
				refreshCache(kv).catch(console.error);
				return cached.data;
			}
		}

		// No cache or expired (> 1 hour): fetch fresh data
		return await refreshCache(kv);
	} catch (error) {
		console.error('Error fetching commits:', error);
		return null;
	}
}

/**
 * Fetches fresh data from Culpa API without caching.
 */
async function fetchFromAPI(): Promise<WorkerResponse | null> {
	try {
		const response = await fetch(CULPA_API_URL, {
			headers: {
				Accept: 'application/json',
			},
		});

		if (!response.ok) {
			throw new Error(`Culpa API error: ${response.status} ${response.statusText}`);
		}

		return await response.json();
	} catch (error) {
		console.error('Error fetching from API:', error);
		return null;
	}
}

/**
 * Fetches fresh data from Culpa API and updates the cache.
 */
async function refreshCache(kv: KVNamespace): Promise<WorkerResponse | null> {
	try {
		const data = await fetchFromAPI();
		if (!data) return null;

		// Store in KV cache
		const cacheEntry: CacheEntry = {
			data,
			timestamp: Date.now(),
		};

		await kv.put(KV_KEY, JSON.stringify(cacheEntry));

		return data;
	} catch (error) {
		console.error('Error refreshing cache:', error);
		return null;
	}
}
