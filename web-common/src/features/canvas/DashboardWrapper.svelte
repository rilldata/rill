<script lang="ts">
  import { onMount } from "svelte";
  import { GridStack } from "gridstack";
  import "gridstack/dist/gridstack.min.css";

  export let height: number;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-slate-200";

  let grid;

  onMount(() => {
    grid = GridStack.init({
      column: 12,
    });
  });
</script>

<div
  class="dashboard-theme-boundary size-full bg-gray-100 flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="wrapper {color} max-w-[1440px] min-h-full"
    style:height="{height}px"
  >
    <div class="grid-stack">
      <div class="grid-stack-item" data-gs-w="4" data-gs-h="2">
        <div class="grid-stack-item-content">
          <slot />
        </div>
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    width: 100%;
    height: 100%;
    position: relative;
    user-select: none;
    margin: 0;
    pointer-events: auto;
  }

  .grid-stack-item-content {
    background-color: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }
</style>
