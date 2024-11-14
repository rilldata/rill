<script lang="ts">
  import { type ComponentType, onMount } from "svelte";
  import type GridLines from "./GridLines.svelte";

  export let width: number;
  export let height: number;
  export let scale: number;
  export let embed = false;
  export let showGrid = false;
  export let changing = false;
  export let scrollOffset = 0;
  export let gapSize = 0;
  export let gridCell = 0;
  export let radius = 0;
  export let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  export let color = "bg-slate-200";

  let GridLinesComponent: ComponentType<GridLines>;

  onMount(async () => {
    if (!embed) {
      GridLinesComponent = (await import("./GridLines.svelte")).default;
    }
  });
</script>

<div
  class="size-full bg-gray-100 flex justify-center overflow-y-auto"
  on:scroll
>
  <div
    bind:contentRect
    class="wrapper {color} max-w-[1440px] min-h-full"
    style:height="{height}px"
  >
    {#if GridLinesComponent && (showGrid || changing)}
      <svelte:component
        this={GridLinesComponent}
        {gridCell}
        {scrollOffset}
        {gapSize}
        {radius}
        {scale}
      />
    {/if}
    <div role="presentation" class="size-full relative" on:click|self>
      <div
        class="dash pointer-events-none"
        role="presentation"
        style:width="{width}px"
        style:height="{height}px"
        style:transform="scale({scale})"
      >
        <slot />
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

  .dash {
    transform-origin: top left;
    position: absolute;
    touch-action: none;
  }
</style>
