<script lang="ts">
  import { V1DashboardItem } from "@rilldata/web-common/runtime-client";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";
  import Component from "./Component.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let columns = 20;
  export let items: V1DashboardItem[];
  export let gap = 4;
  export let chartView = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: instanceId = $runtime.instanceId;

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;
  $: gapSize = defaults.DASHBOARD_WIDTH * (gap / 1000);
  $: gridCell = defaults.DASHBOARD_WIDTH / columns;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<DashboardWrapper
  bind:contentRect
  color="bg-slate-50"
  {scale}
  height={maxBottom * gridCell}
  width={defaults.DASHBOARD_WIDTH}
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
