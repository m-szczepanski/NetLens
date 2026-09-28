<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="min-h-screen bg-zinc-950 text-zinc-100 antialiased">
	<header class="border-b border-zinc-800">
		<div class="mx-auto flex max-w-4xl items-center gap-3 px-6 py-4">
			<span class="font-semibold tracking-tight">NetLens</span>
			<span class="rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-400">early development</span>
		</div>
	</header>

	<main class="mx-auto max-w-4xl px-6 py-8">
		{@render children()}
	</main>
</div>
