<script lang="ts">
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { type V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Component from "./Component.svelte";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";

  export let items: V1CanvasItem[];
  export let chartView = false;
  export let showFilterBar = true;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  $: ({ instanceId } = $runtime);

  const dashboardWidth = chartView
    ? defaults.DEFAULT_DASHBOARD_WIDTH / 2
    : defaults.DEFAULT_DASHBOARD_WIDTH;

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / dashboardWidth;
  $: gapSize = dashboardWidth * (defaults.GAP_SIZE / 1000);
  $: gridCell = dashboardWidth / defaults.COLUMN_COUNT;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

{#if showFilterBar}
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
  >
    <CanvasFilters />
  </div>
{/if}

<DashboardWrapper
  bind:contentRect
  {scale}
  height={maxBottom * gridCell * scale}
  width={dashboardWidth}
>
  {#each items as component, i (i)}
    {@const componentName = component.component}
    {#if componentName}
      <Component
        {i}
        {instanceId}
        {componentName}
        {chartView}
        {radius}
        embed={true}
        padding={gapSize}
        width={Number(component.width ?? defaults.COMPONENT_WIDTH) * gridCell}
        height={Number(component.height ?? defaults.COMPONENT_HEIGHT) *
          gridCell}
        left={Number(component.x) * gridCell}
        top={Number(component.y) * gridCell}
        rowIndex={Math.floor(Number(component.y))}
        columnIndex={Math.floor(Number(component.x))}
      />
    {/if}
  {/each}
</DashboardWrapper>
