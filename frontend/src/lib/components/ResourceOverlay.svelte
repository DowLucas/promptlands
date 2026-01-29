<script lang="ts">
	import { playerAgent } from '$lib/stores/game';
</script>

<div class="resource-overlay">
	<div class="resource-item" style="--res-color: #fbbf24">
		<span class="resource-dot"></span>
		<span class="resource-label">Coins</span>
		<span class="resource-count">{$playerAgent?.coins ?? 0}</span>
	</div>
</div>

<style>
	.resource-overlay {
		position: absolute;
		top: 12px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		gap: 12px;
		padding: 6px 14px;
		background: rgba(15, 15, 30, 0.75);
		backdrop-filter: blur(6px);
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.08);
		pointer-events: none;
		z-index: 15;
	}

	.resource-item {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 0.75rem;
		color: #cbd5e1;
	}

	.resource-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--res-color);
		flex-shrink: 0;
	}

	.resource-label {
		color: #94a3b8;
	}

	.resource-count {
		font-weight: 600;
		color: #e2e8f0;
		min-width: 1ch;
	}
</style>
