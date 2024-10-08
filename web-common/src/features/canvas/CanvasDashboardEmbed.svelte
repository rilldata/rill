<script lang="ts">
  import {
    type V1CanvasItem,
    type V1ComponentVariable,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { setContext } from "svelte";
  import Component from "./Component.svelte";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";
  import { canvasVariablesStore } from "./variables-store";

  export let canvasName: string;
  export let columns = 20;
  export let items: V1CanvasItem[];
  export let gap = 4;
  export let chartView = false;
  export let variables: V1ComponentVariable[] = [];

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  setContext("rill::canvas:name", canvasName);

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

  $: if (variables.length && canvasName) {
    canvasVariablesStore.init(canvasName, variables);
  }
</script>

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
</DashboardWrapper>
