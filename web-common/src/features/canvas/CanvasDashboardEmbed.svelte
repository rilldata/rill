<script lang="ts">
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import {
    type V1CanvasItem,
    type V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";

  export let items: V1CanvasItem[];
  export let showFilterBar = true;
  export let spec: V1CanvasSpec;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

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
  {JSON.stringify(spec)}
</CanvasDashboardWrapper>
