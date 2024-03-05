<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 1;

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

<section class="w-full h-fit max-h-full overflow-y-scroll flex-none">
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
        class="flex items-center justify-center col-start-1 flex-grow-0 overflow-hidden rounded"
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
  div {
    @apply border bg-red-300 border-black;
  }
</style>
