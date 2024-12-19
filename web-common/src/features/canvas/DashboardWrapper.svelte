<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { GridStackManager } from "./grid-stack-manager";
  import "gridstack/dist/gridstack.min.css";

  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-transparent";
  export let readonly = false;

  let gridStackElement: HTMLElement;
  const gridStackManager = new GridStackManager(readonly);

  onMount(() => {
    if (gridStackElement) {
      gridStackManager.initialize();
    }
  });

  onDestroy(() => {
    gridStackManager.destroy();
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
    <div bind:this={gridStackElement} class="grid-stack">
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
