<script lang="ts">
  import { type V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasComponent from "./CanvasComponent.svelte";
  import * as defaults from "./constants";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import type { GridStack } from "gridstack";

  export let items: V1CanvasItem[];
  // export let chartView = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let grid: GridStack;

  $: instanceId = $runtime.instanceId;
  $: options = {
    column: 12,
    resizable: {
      handles: "e,se,s,sw,w",
    },
    animate: true,
    float: true,
    staticGrid: true,
  };

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
  <SvelteGridStack bind:grid {options} {items} let:index let:item>
    {@const componentName = item.component}
    {#if componentName}
      <CanvasComponent embed i={index} {instanceId} {componentName} />
    {/if}
  </SvelteGridStack>
</CanvasDashboardWrapper>
