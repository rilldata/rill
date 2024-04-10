<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import Chart from "./Chart.svelte";
  import { DEFAULT_WIDTH, DEFAULT_RADIUS } from "./constants";
  import Wrapper from "./Wrapper.svelte";
  import Component from "./Component.svelte";

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 4;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / DEFAULT_WIDTH;
  $: gapSize = DEFAULT_WIDTH * (gap / 1000);
  $: gridCell = DEFAULT_WIDTH / columns;
  $: radius = gridCell * DEFAULT_RADIUS;

  $: maxBottom = charts.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<Wrapper
  {scale}
  width={DEFAULT_WIDTH}
  height={maxBottom * gridCell}
  bind:contentRect
>
  {#each charts as chart, i (i)}
    {#if chart.chart}
      <Component
        embed
        {i}
        {scale}
        {radius}
        padding={gapSize}
        width={Number(chart.width) * gridCell}
        height={Number(chart.height) * gridCell}
        left={Number(chart.x) * gridCell}
        top={Number(chart.y) * gridCell}
      >
        <Chart chartName={chart.chart} />
      </Component>
    {/if}
  {/each}
</Wrapper>
