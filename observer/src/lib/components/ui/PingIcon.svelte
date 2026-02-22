<script lang="ts">
	interface Props {
		class?: string;
	}

	let { class: className = '' }: Props = $props();
</script>

<div class="relative {className}">
	<div class="h-3 w-3 animate-ping rounded-full bg-green-500"></div>
	<div class="absolute left-0 top-0 h-3 w-3 animate-ping rounded-full bg-green-500"></div>
	<div class="absolute left-0 top-0 h-3 w-3 animate-pulse rounded-full bg-green-500"></div>
</div>
