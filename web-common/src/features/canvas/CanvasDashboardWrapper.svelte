<script lang="ts">
  import GridStack from "./GridStack.svelte";

  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-transparent";
  export let readonly = false;
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
    <GridStack {readonly}>
      <slot />
    </GridStack>
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
