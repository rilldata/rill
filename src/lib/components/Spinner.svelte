<script>
export let size = '1rem';
export let bg = 'linear-gradient(to left, hsl(300, 100%, 50%), hsl(1, 100%, 50%))';
</script>

<div class='spinner spinning' style="width: {size}; height: {size}; background: {bg};" />

<style>
	.spinner {
        opacity: .5;
	}
	
	.spinning {
		animation: spin 1s infinite;
	}
	
	@keyframes spin {
		0% {
			transform: rotate(360deg);
		} 100% {
			transform: rotate(0deg);
		}
	}
</style>