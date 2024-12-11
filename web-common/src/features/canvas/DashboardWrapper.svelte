<script lang="ts">
  import { type ComponentType, onMount } from "svelte";
  import type GridLines from "./GridLines.svelte";
  import * as defaults from "./constants";

  export let width = defaults.DASHBOARD_WIDTH;
  export let height: number;
  export let scale: number;
  export let embed = false;
  export let showGrid = false;
  export let changing = false;
  export let contentRect: DOMRectReadOnly;

  let GridLinesComponent: ComponentType<GridLines>;

  onMount(async () => {
    if (!embed) {
      GridLinesComponent = (await import("./GridLines.svelte")).default;
    }
  });
</script>

<div class="size-full bg-gray-100 flex justify-center overflow-y-auto">
  <div
    bind:contentRect
    class="wrapper bg-slate-200 max-w-[1200px] min-h-full relative"
    style:height="{height}px"
  >
    {#if GridLinesComponent && (showGrid || changing)}
      <svelte:component this={GridLinesComponent} {height} {scale} />
    {/if}
    <div role="presentation" class="grid-container size-full" on:click|self>
      <div
        class="dash"
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

  .grid-container {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    /* FIXME */
    grid-template-rows: repeat(auto-fill, var(--row-height, 120px));
    gap: var(--grid-gap, 24px);
    position: relative;
  }

  .dash {
    transform-origin: top left;
    position: absolute;
    touch-action: none;
    width: 100%;
    height: 100%;
  }
</style>
