<script lang="ts">
  import { type V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Component from "./Component.svelte";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { canvasVariablesStore } from "./variables-store";
  import GridStackItem from "./GridStackItem.svelte";

  export let columns = 20;
  export let items: V1CanvasItem[];
  export let gap = 1;
  export let chartView = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  $: instanceId = $runtime.instanceId;

  const dashboardWidth = chartView
    ? defaults.DASHBOARD_WIDTH / 2
    : defaults.DASHBOARD_WIDTH;

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / dashboardWidth;
  $: gapSize = dashboardWidth * (gap / 1000);
  $: gridCell = dashboardWidth / columns;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<!-- <CanvasDashboardWrapper
  bind:contentRect
  height={maxBottom * gridCell * scale}
  readonly={true}
>
  {#each items as component, i (i)}
    {@const componentName = component.component}
    {#if componentName}
      <Component
        embed
        {i}
        {instanceId}
        {componentName}
        {chartView}
        {scale}
        {radius}
        padding={gapSize}
        width={Number(component.width ?? defaults.COMPONENT_WIDTH) * gridCell}
        height={Number(component.height ?? defaults.COMPONENT_HEIGHT) *
          gridCell}
        left={Number(component.x) * gridCell}
        top={Number(component.y) * gridCell}
      />
    {/if}
  {/each}
</CanvasDashboardWrapper> -->

<CanvasDashboardWrapper
  bind:contentRect
  height={maxBottom * gridCell * scale}
  readonly={true}
>
  {#each items as component, i (i)}
    <GridStackItem
      index={i}
      width={component.width}
      height={component.height}
      x={component.x}
      y={component.y}
      content={component.component}
    />
  {/each}
</CanvasDashboardWrapper>
