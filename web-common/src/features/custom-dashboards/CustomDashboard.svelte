<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 1;
  export let showGrid = false;

  let canvasWidth = 0;

  $: maxRow = Math.max(
    ...charts.map(
      (chart) => (Number(chart?.y) ?? 0) + (Number(chart.height) ?? 0) - 1,
    ),
  );

  $: gapSize = canvasWidth * (gap / 100);

  $: gapCount = columns - 1;

  $: gridCellSize = (canvasWidth - gapCount * gapSize) / columns;
</script>

<section
  class="w-full h-fit max-h-full min-h-screen overflow-y-scroll flex-none relative"
>
  {#if showGrid}
    <svg
      width="100%"
      height="100%"
      xmlns="http://www.w3.org/2000/svg"
      class="absolute"
    >
      <defs>
        <pattern
          id="grid"
          width={gridCellSize + gapSize}
          height={gridCellSize + gapSize}
          patternUnits="userSpaceOnUse"
        >
          <rect width={gridCellSize} height={gridCellSize} fill="white" />
          <path
            d="M {gridCellSize + gapSize} 0 L 0 0 0 {gridCellSize + gapSize}"
            fill="none"
            stroke="none"
            stroke-width="1"
          />
        </pattern>
      </defs>

      <rect width="100%" height="100%" fill="url(#grid)" />
    </svg>
  {/if}

  <section
    class="grid w-full h-fit"
    style:gap="{gapSize}px"
    style:grid-template-columns="repeat({columns}, 1fr)"
    style:grid-template-rows="repeat({maxRow + 1}, {gridCellSize}px)"
    bind:clientWidth={canvasWidth}
  >
    {#each charts as chart, i (i)}
      <div
        data-index={i}
        class="item flex items-center justify-center col-start-1 flex-grow-0 overflow-hidden rounded"
        style:grid-column-start={chart.x}
        style:grid-row-start={chart.y}
        style:grid-column-end="span {chart.width}"
        style:grid-row-end="span {chart.height}"
      >
        {chart.chart}
      </div>
    {/each}
  </section>
</section>

<style lang="postcss">
  .item {
    @apply border bg-red-300 border-black;
  }
</style>
