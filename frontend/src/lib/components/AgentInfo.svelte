<script lang="ts">
	import type { Agent } from '$lib/types';

	export let agent: Agent | null = null;
	export let isPlayer: boolean = false;

	function getHPColor(hp: number, maxHP: number): string {
		const ratio = hp / maxHP;
		if (ratio > 0.6) return '#22c55e';
		if (ratio > 0.3) return '#eab308';
		return '#ef4444';
	}

	function getEnergyColor(energy: number, maxEnergy: number): string {
		const ratio = energy / maxEnergy;
		if (ratio > 0.6) return '#3b82f6';
		if (ratio > 0.3) return '#8b5cf6';
		return '#6366f1';
	}
</script>

{#if agent}
	<div class="agent-info" class:is-player={isPlayer} class:is-dead={agent.is_dead}>
		<div class="header">
			<span class="name">{agent.name}</span>
			{#if agent.is_dead}
				<span class="death-badge">DEAD</span>
			{/if}
		</div>

		<div class="stats">
			<!-- HP Bar -->
			<div class="stat-row">
				<span class="stat-label">HP</span>
				<div class="bar-container">
					<div
						class="bar hp-bar"
						style="width: {(agent.hp / agent.max_hp) * 100}%; background-color: {getHPColor(agent.hp, agent.max_hp)}"
					></div>
				</div>
				<span class="stat-value">{agent.hp}/{agent.max_hp}</span>
			</div>

			<!-- Energy Bar -->
			<div class="stat-row">
				<span class="stat-label">Energy</span>
				<div class="bar-container">
					<div
						class="bar energy-bar"
						style="width: {(agent.energy / agent.max_energy) * 100}%; background-color: {getEnergyColor(agent.energy, agent.max_energy)}"
					></div>
				</div>
				<span class="stat-value">{agent.energy}/{agent.max_energy}</span>
			</div>
		</div>

		<div class="upgrades">
			<div class="upgrade" title="Vision - Increases sight range">
				<span class="upgrade-icon">👁</span>
				<span class="upgrade-level">{agent.vision_level}</span>
			</div>
			<div class="upgrade" title="Memory - More memory slots">
				<span class="upgrade-icon">🧠</span>
				<span class="upgrade-level">{agent.memory_level}</span>
			</div>
			<div class="upgrade" title="Strength - More attack damage">
				<span class="upgrade-icon">💪</span>
				<span class="upgrade-level">{agent.strength_level}</span>
			</div>
			<div class="upgrade" title="Storage - More inventory slots">
				<span class="upgrade-icon">🎒</span>
				<span class="upgrade-level">{agent.storage_level}</span>
			</div>
		</div>

		<div class="position">
			Position: ({agent.position.x}, {agent.position.y})
		</div>
	</div>
{/if}

<style>
	.agent-info {
		background: #1f2937;
		border: 1px solid #374151;
		border-radius: 8px;
		padding: 12px;
		font-size: 14px;
	}

	.agent-info.is-player {
		border-color: #3b82f6;
	}

	.agent-info.is-dead {
		opacity: 0.6;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.name {
		font-weight: 600;
		color: #f3f4f6;
	}

	.death-badge {
		background: #ef4444;
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 10px;
		font-weight: 600;
	}

	.stats {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin-bottom: 12px;
	}

	.stat-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.stat-label {
		width: 50px;
		color: #9ca3af;
		font-size: 12px;
	}

	.bar-container {
		flex: 1;
		height: 8px;
		background: #374151;
		border-radius: 4px;
		overflow: hidden;
	}

	.bar {
		height: 100%;
		transition: width 0.3s ease;
	}

	.stat-value {
		width: 60px;
		text-align: right;
		color: #d1d5db;
		font-size: 12px;
		font-family: monospace;
	}

	.upgrades {
		display: flex;
		justify-content: space-around;
		margin-bottom: 12px;
		padding: 8px;
		background: #111827;
		border-radius: 6px;
	}

	.upgrade {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		cursor: help;
	}

	.upgrade-icon {
		font-size: 16px;
	}

	.upgrade-level {
		font-size: 12px;
		font-weight: 600;
		color: #fbbf24;
	}

	.position {
		color: #6b7280;
		font-size: 12px;
		text-align: center;
	}
</style>
