<script lang="ts">
  import { scaleTime } from "d3-scale";

  export let start: Date;
  export let end: Date;
  export let right: number = 25;
  export let height: number = 26;

  // Compute width from parent (will be set via bind:clientWidth)
  let containerWidth = 300;

  $: plotRight = containerWidth - right;
  $: xScale = scaleTime().domain([start, end]).range([0, plotRight]);
  $: ticks = xScale.ticks(Math.max(2, Math.floor(containerWidth / 100)));

  function formatTick(date: Date): string {
    const month = date.toLocaleString(undefined, { month: "short" });
    const day = date.getDate();
    const year = date.getFullYear();
    // Show year if ticks span multiple years
    const startYear = start.getFullYear();
    const endYear = end.getFullYear();
    if (startYear !== endYear) {
      return `${month} ${day}, ${year}`;
    }
    return `${month} ${day}`;
  }
</script>

<div bind:clientWidth={containerWidth} style="height: {height}px;">
  <svg width={containerWidth} {height}>
    {#each ticks as tick}
      <text
        class="fill-fg-secondary text-[11px]"
        text-anchor="start"
        x={xScale(tick)}
        y={height - 4}
      >
        {formatTick(tick)}
      </text>
    {/each}
  </svg>
</div>
