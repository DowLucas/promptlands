<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import AgentEditor from '$lib/components/AgentEditor.svelte';
	import type { AdversaryType } from '$lib/types';

	let adversaryTypes: AdversaryType[] = [];
	let loading = false;
	let error = '';

	onMount(async () => {
		try {
			const res = await fetch('/api/adversaries');
			if (res.ok) {
				adversaryTypes = await res.json();
			}
		} catch (e) {
			console.error('Failed to fetch adversary types:', e);
		}
	});

	async function startGame(prompt: string, adversaries: string[]) {
		loading = true;
		error = '';

		try {
			const res = await fetch('/api/games/singleplayer', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					player_prompt: prompt,
					adversaries: adversaries
				})
			});

			if (!res.ok) {
				const data = await res.json();
				throw new Error(data.error || 'Failed to create game');
			}

			const data = await res.json();
			goto(`/game/${data.game_id}?agent=${data.player_agent_id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Unknown error';
			loading = false;
		}
	}
</script>

<div class="page">
	<header>
		<a href="/" class="back">Back</a>
		<h1>Singleplayer</h1>
	</header>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<div class="loading">Starting game...</div>
	{:else}
		<AgentEditor {adversaryTypes} onStart={startGame} />
	{/if}
</div>

<style>
	.page {
		max-width: 800px;
		margin: 0 auto;
		padding: 24px;
	}

	header {
		display: flex;
		align-items: center;
		gap: 16px;
		margin-bottom: 24px;
	}

	.back {
		color: #a0aec0;
		text-decoration: none;
	}

	.back:hover {
		color: #e2e8f0;
	}

	h1 {
		font-size: 24px;
	}

	.error {
		background: #fc8181;
		color: #1a1a2e;
		padding: 12px 16px;
		border-radius: 8px;
		margin-bottom: 16px;
	}

	.loading {
		text-align: center;
		padding: 48px;
		color: #a0aec0;
	}
</style>
