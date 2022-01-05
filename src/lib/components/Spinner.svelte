<div class='spinner spinning' />

<style>
	.spinner {
		width: 1rem;
		height: 1rem;
        opacity: .5;
		background: linear-gradient(to left, hsl(300, 100%, 50%), hsl(1, 100%, 50%));
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