import * as z from 'zod';
import { VIEW_MODE } from '$lib/stores/settings.svelte';

const viewModeValues = Object.values(VIEW_MODE) as [string, ...string[]];

// mirrors every field in AppSettings so the embed looks identical to the
// regular browser view that generated the URL
export const embedParamsSchema = z.object({
	viewMode: z.enum(viewModeValues).default(VIEW_MODE.MAX),
	displayType: z.enum(['FOUR_FOUR', 'EIGHT_EIGHT', 'CHUNK']).default('FOUR_FOUR'),
	colorNegatives: z.coerce.boolean().default(true),
	autoHide: z.coerce.number().default(0),
});

export type EmbedParams = z.infer<typeof embedParamsSchema>;
