<script lang="ts">
  import { type V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasComponent from "./CanvasComponent.svelte";
  import * as defaults from "./constants";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  // import GridStackItem from "./GridStackItem.svelte";
  // import { canvasVariablesStore } from "./variables-store";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import { EMBED_GRIDSTACK_OPTIONS } from "./constants";

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

  // $: if (variables.length && canvasName) {
  //   canvasVariablesStore.init(canvasName, variables);
  // }

  $: items = items.map((item) => ({
    ...item,
    component: item.component ?? "",
    w: Number(item.width),
    h: Number(item.height),
    x: Number(item.x),
    y: Number(item.y),
  }));
</script>

<CanvasDashboardWrapper bind:contentRect height={maxBottom * gridCell * scale}>
  <SvelteGridStack options={EMBED_GRIDSTACK_OPTIONS} {items} let:index let:item>
    {@const componentName = item.component}
    {#if componentName}
      <CanvasComponent
        embed
        i={index}
        {instanceId}
        {componentName}
        width={Number(item.w ?? defaults.COMPONENT_WIDTH) * gridCell}
        height={Number(item.h ?? defaults.COMPONENT_HEIGHT) * gridCell}
        left={Number(item.x) * gridCell}
        top={Number(item.y) * gridCell}
      />
    {/if}
  </SvelteGridStack>
</CanvasDashboardWrapper>
