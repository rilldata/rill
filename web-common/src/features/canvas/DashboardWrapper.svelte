<script lang="ts">
  import { onMount } from "svelte";
  import { GridStack, type GridStackOptions } from "gridstack";
  import "gridstack/dist/gridstack.min.css";

  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-slate-200";

  let grid;

  const options: GridStackOptions = {
    column: 12,
    // TODO: staticGrid for previewMode
    // if url is /custom/{canvasName} then staticGrid
  };

  onMount(() => {
    grid = GridStack.init(options);
  });
</script>

<div
  class="dashboard-theme-boundary size-full bg-gray-100 flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="canvas {color} max-w-[1440px] min-h-full"
    style:height="{height}px"
  >
    <div class="grid-stack">
      <slot />
    </div>
  </div>
</div>

<style lang="postcss">
  .canvas {
    width: 100%;
    height: 100%;
    position: relative;
    user-select: none;
    margin: 0;
    pointer-events: auto;
  }
</style>
