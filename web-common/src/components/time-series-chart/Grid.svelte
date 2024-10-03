<script lang="ts">
	import type { ScaleLinear, ScaleTime } from 'd3-scale';

	export let xScale: ScaleTime<number, number>;
	export let yScale: ScaleLinear<number, number>;

	let showX = true;
	let showY = true;
</script>

<g shape-rendering="crispEdges">
	{#if showX}
		{#each xScale.ticks(3) as tick, i (i)}
			<line
				x1={xScale(tick)}
				x2={xScale(tick)}
				y1={0}
				y2={100}
				class="stroke-gray-300"
				stroke-width={1}
				stroke-dasharray="1,1"
				vector-effect="non-scaling-stroke"
			/>
		{/each}
	{/if}
	{#if showY}
		{#each yScale.ticks(2) as tick, i (i)}
			<line
				y1={yScale(tick)}
				y2={yScale(tick)}
				x1={0}
				x2={100}
				class="stroke-gray-300"
				stroke-width={1}
				stroke-dasharray="1,1"
				vector-effect="non-scaling-stroke"
			/>
		{/each}
	{/if}
</g>
