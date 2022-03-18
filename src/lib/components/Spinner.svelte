<script lang="ts">
import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
export let size = '1rem';
export let status:EntityStatus = EntityStatus.Idle;
export let bg = 'linear-gradient(to left, hsla(300, 100%, 50%, .5), hsla(1, 100%, 50%, .5))';
</script>

<div 
	class="status"
	class:running={status === EntityStatus.Running}
	class:idle={status === EntityStatus.Idle}
	style="
		--status-transition: 300ms;
		--background: {bg};
		--size: {size};
		width: {size}; height: {size};" />

<style>
	div {
		border-radius: 0px;
		transition: border-radius var(--status-transition), background var(--status-transition), border-color var(--status-transition);
		border: .125rem solid transparent;
		animation: spin 1s infinite;
		position: relative;
	}

	div::before {
		content: ' ';
		display: block;
		width: 100%;
		height: 100%;
		transition: opacity calc(var(--status-transition)),  border-radius var(--status-transition), transform var(--status-transition);
		background: var(--background);
		z-index: 100000;

	}
	.running {
		border-radius: 0px;
		position:relative;
	}

	.running::before {
		opacity: 1;
	}
	.idle {
		border-radius: 50%;
		border: .125rem solid currentColor;
	}

	.idle::before {
		opacity: 0;
		border-radius: 50%;
		transform: rotate(-180deg);
		transition: opacity var(--status-transition),  border-radius var(--status-transition), transform var(--status-transition);
	}
		
	@keyframes spin {
		0% {
			transform: rotate(360deg);
		} 100% {
			transform: rotate(0deg);
		}
	}
</style>