<script lang="ts">
	import type { AdversaryType, MapSize, MapPreset, MapConfig } from '$lib/types';

	export let onStart: (prompt: string, adversaries: string[], mapConfig: MapConfig) => void;
	export let adversaryTypes: AdversaryType[] = [];

	let prompt = `You are a strategic agent competing for territory.
Your goal is to claim as many tiles as possible.
Be efficient with your movements and consider when to claim vs when to explore.
Pay attention to other agents' positions and messages.`;

	let selectedAdversaries: string[] = ['aggressive', 'defensive'];
	let playerName = 'Player';

	// Map configuration
	let selectedPreset = 'default';
	let selectedSize: MapSize = 'huge';

	// Available map presets
	const mapPresets: MapPreset[] = [
		{ id: 'default', name: 'Standard World', description: 'Balanced world with all biomes', theme: 'mixed', size: 'huge' },
		{ id: 'infernal_realms', name: 'Infernal Realms', description: 'Volcanic world with lava and badlands', theme: 'volcanic', size: 'huge' },
		{ id: 'frozen_wastes', name: 'Frozen Wastes', description: 'Icy world with crystal formations', theme: 'ice', size: 'huge' },
		{ id: 'ancient_world', name: 'Ancient World', description: 'Forests and ruins to explore', theme: 'ancient', size: 'massive' },
		{ id: 'neon_wilderness', name: 'Neon Wilderness', description: 'Alien bioluminescent landscape', theme: 'scifi', size: 'huge' },
		{ id: 'void_incursion', name: 'Void Incursion', description: 'Reality collapsing - high difficulty', theme: 'horror', size: 'large' },
		{ id: 'crystalline_expanse', name: 'Crystalline Expanse', description: 'Crystal-dominated mystical world', theme: 'crystal', size: 'huge' }
	];

	// Available map sizes
	const mapSizes: { value: MapSize; label: string; tiles: string }[] = [
		{ value: 'tiny', label: 'Tiny', tiles: '128x128' },
		{ value: 'small', label: 'Small', tiles: '256x256' },
		{ value: 'medium', label: 'Medium', tiles: '512x512' },
		{ value: 'large', label: 'Large', tiles: '1024x1024' },
		{ value: 'huge', label: 'Huge', tiles: '2048x2048' },
		{ value: 'massive', label: 'Massive', tiles: '4096x4096' }
	];

	function toggleAdversary(type: string) {
		if (selectedAdversaries.includes(type)) {
			selectedAdversaries = selectedAdversaries.filter(t => t !== type);
		} else if (selectedAdversaries.length < 3) {
			selectedAdversaries = [...selectedAdversaries, type];
		}
	}

	function selectPreset(presetId: string) {
		selectedPreset = presetId;
		// Update size to match preset default
		const preset = mapPresets.find(p => p.id === presetId);
		if (preset) {
			selectedSize = preset.size;
		}
	}

	function handleStart() {
		if (prompt.trim() && selectedAdversaries.length > 0) {
			const mapConfig: MapConfig = {
				preset: selectedPreset,
				size: selectedSize
			};
			onStart(prompt, selectedAdversaries, mapConfig);
		}
	}

	$: selectedPresetInfo = mapPresets.find(p => p.id === selectedPreset);
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

	<div class="section">
		<h3>World Theme</h3>
		<p class="hint">Choose a themed world to play in.</p>
		<div class="preset-grid">
			{#each mapPresets as preset}
				<button
					class="preset-btn"
					class:selected={selectedPreset === preset.id}
					on:click={() => selectPreset(preset.id)}
				>
					<span class="preset-name">{preset.name}</span>
					<span class="preset-desc">{preset.description}</span>
				</button>
			{/each}
		</div>
	</div>

	<div class="section">
		<h3>Map Size</h3>
		<p class="hint">Larger maps create a pixelated aesthetic when zoomed out.</p>
		<div class="size-grid">
			{#each mapSizes as size}
				<button
					class="size-btn"
					class:selected={selectedSize === size.value}
					on:click={() => selectedSize = size.value}
				>
					<span class="size-label">{size.label}</span>
					<span class="size-tiles">{size.tiles}</span>
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
		max-width: 700px;
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

	.preset-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 8px;
	}

	.preset-btn {
		background: #1a1a2e;
		border: 2px solid #4a5568;
		border-radius: 8px;
		padding: 12px;
		cursor: pointer;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 4px;
		text-align: left;
		transition: all 0.2s;
	}

	.preset-btn:hover {
		border-color: #718096;
	}

	.preset-btn.selected {
		border-color: #6366f1;
		background: #1a1a3e;
	}

	.preset-name {
		color: #e2e8f0;
		font-weight: bold;
		font-size: 14px;
	}

	.preset-desc {
		color: #a0aec0;
		font-size: 12px;
	}

	.size-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
		gap: 8px;
	}

	.size-btn {
		background: #1a1a2e;
		border: 2px solid #4a5568;
		border-radius: 8px;
		padding: 10px;
		cursor: pointer;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		transition: all 0.2s;
	}

	.size-btn:hover {
		border-color: #718096;
	}

	.size-btn.selected {
		border-color: #f6ad55;
		background: #2d2a1a;
	}

	.size-label {
		color: #e2e8f0;
		font-weight: bold;
		font-size: 14px;
	}

	.size-tiles {
		color: #a0aec0;
		font-size: 11px;
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
