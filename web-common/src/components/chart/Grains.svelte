<script lang="ts">
  import { Point } from "./Chart.svelte";

  export let referenceLine: Point[];
  export let yMin: number;
  export let height: number;
  export let hoveredIndex: number | null;
</script>

<g class="absolute">
  {#each referenceLine as { x }, i (i)}
    {#if i !== referenceLine.length - 1}
      {@const nextStart = referenceLine[i + 1].x}
      <rect
        {x}
        y={yMin}
        width={Math.abs(nextStart - x)}
        {height}
        class="fill-gray-300 opacity-0"
        class:opacity-40={hoveredIndex === i}
      />
    {/if}
  {/each}
</g>
