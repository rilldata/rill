<script lang="ts">
  export let yExtents: [number, number];
  export let xTicks: number[];
  export let yTicks: number[];
  export let xExtents: [number, number];
</script>

<g shape-rendering="crispEdges">
  {#each xTicks as x1, index (index)}
    <line {x1} x2={x1} y1={yExtents[0]} y2="100%" />
  {/each}

  {#each yTicks as y1, i (i)}
    {#if !(i === 0 && yExtents[0] === 0)}
      <line {y1} y2={y1} x1={xExtents[0]} x2="100%" />
    {/if}
  {/each}
</g>

<style lang="postcss">
  line {
    stroke-width: 1px;
    stroke-dasharray: 1px 1px;
    vector-effect: non-scaling-stroke;
    @apply stroke-gray-300;
  }
</style>
