<script lang="ts">
  import { TimeSeriesLine } from "./Chart.svelte";

  export let lines: TimeSeriesLine[];
</script>

{#each lines as { data, key, color } (key)}
  <g>
    {#each data as { x, y }}
      {#if y !== null && y !== undefined}
        <line
          x1={x}
          x2={x}
          y1={0}
          y2={y}
          class="stroke-{color ?? 'primary-500'}"
        />
        <path
          d="M{x} {y} l 0 0"
          stroke-width="3"
          stroke-linecap="round"
          stroke={color ? "black" : "var(--color-primary-500)"}
        />
      {/if}
    {/each}
  </g>
{/each}

<style lang="postcss">
  line {
    stroke-width: 1.5px;
  }

  line,
  path {
    vector-effect: non-scaling-stroke;
  }
</style>
