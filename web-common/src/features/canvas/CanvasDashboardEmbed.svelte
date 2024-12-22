<script lang="ts">
  import { type V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import type { GridStack } from "gridstack";

  export let items: V1CanvasItem[];
  // export let chartView = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let grid: GridStack;

  $: instanceId = $runtime.instanceId;

  // const dashboardWidth = chartView
  //   ? defaults.DASHBOARD_WIDTH / 2
  //   : defaults.DASHBOARD_WIDTH;

  // $: gridWidth = contentRect.width;
  // $: scale = gridWidth / dashboardWidth;
  // $: gridCell = dashboardWidth / columns;
  // $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);

  // $: if (variables.length && canvasName) {
  //   canvasVariablesStore.init(canvasName, variables);
  // }
</script>

<CanvasDashboardWrapper bind:contentRect height={maxBottom}>
  <SvelteGridStack bind:grid {items} let:index let:item embed>
    {@const componentName = item.component}
    {#if componentName}
      <CanvasComponent embed i={index} {instanceId} {componentName} />
    {/if}
  </SvelteGridStack>
</CanvasDashboardWrapper>
