<script lang="ts">
  import {
    type V1CanvasItem,
    type V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasComponent from "./CanvasComponent.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import type { GridStack } from "gridstack";
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";

  export let items: V1CanvasItem[];
  export let showFilterBar = true;
  export let spec: V1CanvasSpec;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let grid: GridStack;

  $: instanceId = $runtime.instanceId;

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

<CanvasDashboardWrapper bind:contentRect height={maxBottom}>
  {#each items as component, i (i)}
    <SvelteGridStack bind:grid {items} {spec} let:index let:item embed>
      {@const componentName = item.component}
      {#if componentName}
        <CanvasComponent embed i={index} {instanceId} {componentName} />
      {/if}
    </SvelteGridStack>
  {/each}
</CanvasDashboardWrapper>
