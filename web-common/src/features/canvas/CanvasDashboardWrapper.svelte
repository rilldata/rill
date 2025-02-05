<script lang="ts">
  import * as defaults from "./constants";
  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let bgColor = "bg-white";
</script>

<div
  class="dashboard-theme-boundary size-full bg-gray-100 flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="canvas {bgColor} min-h-full"
    style:height="{height}px"
    style:max-width="{defaults.DEFAULT_DASHBOARD_WIDTH}px"
  >
    <slot />
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
