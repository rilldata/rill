<script lang="ts">
  import { V1DashboardItem } from "@rilldata/web-common/runtime-client";
  import Chart from "./Chart.svelte";
  import * as defaults from "./constants";
  import Wrapper from "./Wrapper.svelte";
  import Component from "./Component.svelte";

  export let columns = 20;
  export let components: V1DashboardItem[];
  export let gap = 4;
  export let chartView = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;
  $: gapSize = defaults.DASHBOARD_WIDTH * (gap / 1000);
  $: gridCell = defaults.DASHBOARD_WIDTH / columns;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: maxBottom = components.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<Wrapper
  bind:contentRect
  color="bg-slate-50"
  height={maxBottom * gridCell}
  {scale}
  width={defaults.DASHBOARD_WIDTH}
>
  {#each components as component, i (i)}
    {#if component.component && typeof component.component === "string"}
      <Component
        {chartView}
        embed
        {i}
        {scale}
        {radius}
        padding={gapSize}
        width={Number(component.width ?? defaults.COMPONENT_WIDTH) * gridCell}
        height={Number(component.height ?? defaults.COMPONENT_HEIGHT) *
          gridCell}
        left={Number(component.x) * gridCell}
        top={Number(component.y) * gridCell}
      >
        <!-- TODO -->
        <!--{#if component.markdown}-->
        <!--  <Markdown-->
        <!--    markdown={component.markdown}-->
        <!--    fontSize={component.fontSize ?? defaults.FONT_SIZE}-->
        <!--  />-->
        <!--{:else if component.chart}-->
        <Chart {chartView} chartName={component.component} />
        <!--{/if}-->
      </Component>
    {/if}
  {/each}
</Wrapper>
