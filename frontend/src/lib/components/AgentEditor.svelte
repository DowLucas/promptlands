<script lang="ts">
	import type { AdversaryType } from '$lib/types';

	export let onStart: (prompt: string, adversaries: string[]) => void;
	export let adversaryTypes: AdversaryType[] = [];

	let prompt = `You are a strategic agent competing for territory.
Your goal is to claim as many tiles as possible.
Be efficient with your movements and consider when to claim vs when to explore.
Pay attention to other agents' positions and messages.`;

	let selectedAdversaries: string[] = ['aggressive', 'defensive'];
	let playerName = 'Player';

	function toggleAdversary(type: string) {
		if (selectedAdversaries.includes(type)) {
			selectedAdversaries = selectedAdversaries.filter(t => t !== type);
		} else if (selectedAdversaries.length < 3) {
			selectedAdversaries = [...selectedAdversaries, type];
		}
	}

	function handleStart() {
		if (prompt.trim() && selectedAdversaries.length > 0) {
			onStart(prompt, selectedAdversaries);
		}
	}
</script>

<div class="editor">
	<div class="section">
		<h3>Your Agent's Instructions</h3>
		<p class="hint">Write the system prompt that will guide your AI agent's decisions.</p>
		<textarea
			bind:value={prompt}
			placeholder="Enter your agent's instructions..."
			rows="8"
		></textarea>
	</div>

	<div class="section">
		<h3>Select Opponents</h3>
		<p class="hint">Choose 1-3 AI adversaries to compete against.</p>
		<div class="adversary-grid">
			{#each adversaryTypes as adv}
				<button
					class="adversary-btn"
					class:selected={selectedAdversaries.includes(adv.type)}
					on:click={() => toggleAdversary(adv.type)}
				>
					<span class="name">{adv.name}</span>
					<span class="type">{adv.type}</span>
				</button>
			{/each}
		</div>
	</div>

	<button
		class="start-btn"
		on:click={handleStart}
		disabled={!prompt.trim() || selectedAdversaries.length === 0}
	>
		Start Game
	</button>
</div>

<style>
	.editor {
		background: #16213e;
		border-radius: 8px;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 24px;
		max-width: 600px;
		margin: 0 auto;
	}

	.section h3 {
		margin-bottom: 8px;
		color: #e2e8f0;
	}

	.hint {
		color: #a0aec0;
		font-size: 14px;
		margin-bottom: 12px;
	}

	textarea {
		width: 100%;
		background: #1a1a2e;
		border: 1px solid #4a5568;
		border-radius: 4px;
		padding: 12px;
		color: #e2e8f0;
		font-family: inherit;
		font-size: 14px;
		resize: vertical;
	}

	textarea:focus {
		outline: none;
		border-color: #48bb78;
	}

	.adversary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
		gap: 8px;
	}

	.adversary-btn {
		background: #1a1a2e;
		border: 2px solid #4a5568;
		border-radius: 8px;
		padding: 12px;
		cursor: pointer;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		transition: all 0.2s;
	}

	.adversary-btn:hover {
		border-color: #718096;
	}

	.adversary-btn.selected {
		border-color: #48bb78;
		background: #1a3630;
	}

	.adversary-btn .name {
		color: #e2e8f0;
		font-weight: bold;
	}

	.adversary-btn .type {
		color: #a0aec0;
		font-size: 12px;
	}

	.start-btn {
		background: #48bb78;
		color: #1a1a2e;
		border: none;
		border-radius: 8px;
		padding: 16px 32px;
		font-size: 16px;
		font-weight: bold;
		cursor: pointer;
		transition: background 0.2s;
	}

	.start-btn:hover:not(:disabled) {
		background: #38a169;
	}

	.start-btn:disabled {
		background: #4a5568;
		cursor: not-allowed;
	}
</style>
