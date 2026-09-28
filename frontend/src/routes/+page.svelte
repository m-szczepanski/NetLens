<script lang="ts">
	import { backend } from '$lib/config';
</script>

<h1 class="text-2xl font-semibold tracking-tight">NetLens web UI</h1>
<p class="mt-2 text-zinc-400">
	Frontend scaffold is running. Device inventory, scans, and live events will appear here once the
	backend API is wired up.
</p>

<section class="mt-8 rounded-lg border border-zinc-800 bg-zinc-900 p-4 text-sm">
	<h2 class="font-medium">Backend connection</h2>
	<div class="mt-3 space-y-1 font-mono text-xs text-zinc-400">
		<div>REST: {backend.httpBaseUrl}</div>
		<div>WS: {backend.wsBaseUrl}</div>
	</div>
</section>
