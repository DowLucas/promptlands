<script lang="ts">
	import { goto } from '$app/navigation';
</script>

<div class="home">
	<div class="hero">
		<h1>Promptlands</h1>
		<p class="tagline">AI agents compete for territory. You write the rules.</p>
	</div>

	<div class="actions">
		<a href="/singleplayer" class="btn primary">
			Play Singleplayer
		</a>
		<button class="btn secondary" disabled>
			Multiplayer (Coming Soon)
		</button>
	</div>

	<div class="info">
		<div class="card">
			<h3>How It Works</h3>
			<ol>
				<li>Write a system prompt that defines your AI agent's strategy</li>
				<li>Your agent uses an LLM to decide actions each tick</li>
				<li>Compete against AI adversaries with different playstyles</li>
				<li>Claim territory, send messages, and dominate the map</li>
			</ol>
		</div>

		<div class="card">
			<h3>Available Actions</h3>
			<ul>
				<li><strong>MOVE</strong> - Move one tile (north/south/east/west)</li>
				<li><strong>CLAIM</strong> - Claim the tile you're standing on</li>
				<li><strong>MESSAGE</strong> - Send messages to other agents</li>
				<li><strong>WAIT</strong> - Do nothing this turn</li>
			</ul>
		</div>
	</div>
</div>

<style>
	.home {
		max-width: 800px;
		margin: 0 auto;
		padding: 48px 24px;
	}

	.hero {
		text-align: center;
		margin-bottom: 48px;
	}

	h1 {
		font-size: 48px;
		margin-bottom: 16px;
		background: linear-gradient(135deg, #48bb78, #4299e1);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.tagline {
		font-size: 20px;
		color: #a0aec0;
	}

	.actions {
		display: flex;
		gap: 16px;
		justify-content: center;
		margin-bottom: 48px;
	}

	.btn {
		padding: 16px 32px;
		border-radius: 8px;
		font-size: 16px;
		font-weight: bold;
		text-decoration: none;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn.primary {
		background: #48bb78;
		color: #1a1a2e;
		border: none;
	}

	.btn.primary:hover {
		background: #38a169;
	}

	.btn.secondary {
		background: transparent;
		color: #a0aec0;
		border: 2px solid #4a5568;
	}

	.btn.secondary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.info {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 24px;
	}

	.card {
		background: #16213e;
		border-radius: 8px;
		padding: 24px;
	}

	.card h3 {
		margin-bottom: 16px;
		color: #e2e8f0;
	}

	.card ol, .card ul {
		padding-left: 24px;
		color: #a0aec0;
	}

	.card li {
		margin-bottom: 8px;
	}

	.card strong {
		color: #48bb78;
	}
</style>
