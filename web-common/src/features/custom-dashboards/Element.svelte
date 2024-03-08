<script lang="ts" context="module">
  import type { Vector } from "./types";
  import Chart from "@rilldata/web-common/features/custom-dashboards/Chart.svelte";
  import { Move, Scaling } from "lucide-svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import { quintOut } from "svelte/easing";
  import { get, writable } from "svelte/store";
  import { fade } from "svelte/transition";
  const zIndex = writable(0);
</script>

<script lang="ts">
  const dispatch = createEventDispatcher();

  export let gapSize: number;
  export let dimensions: Vector;
  export let chartName: string;
  export let i: number;
  export let position: Vector;
  export let isMoving: boolean;

  let hover = false;
  let timeout: ReturnType<typeof setTimeout> | null = null;

  let localZIndex = 0;

  onMount(() => {
    localZIndex = get(zIndex);
    zIndex.update((prev) => prev + 1);
  });
</script>

<div
  role="presentation"
  data-index={i}
  class="item absolute flex items-center justify-center"
  style:z-index={localZIndex}
  style:padding="{gapSize}px"
  style:left="{position[0]}px"
  style:top="{position[1]}px"
  style:width="{dimensions[0]}px"
  style:height="{dimensions[1]}px"
  on:mouseenter={() => {
    if (timeout) clearTimeout(timeout);
    hover = true;
  }}
  on:mouseleave={() => {
    timeout = setTimeout(() => {
      hover = false;
      timeout = null;
    }, 800);
  }}
>
  <div class="size-full" class:shadow-lg={isMoving}>
    <Chart {chartName} />
  </div>
</div>

{#if hover}
  <button
    transition:fade={{ duration: 400, easing: quintOut }}
    style:left="{position[0]}px"
    style:top="{position[1]}px"
    class="tool"
    style:z-index={localZIndex + 1}
    data-index={i}
    on:mousedown={(e) => {
      localZIndex = get(zIndex);
      zIndex.update((prev) => prev + 1);
      dispatch("mousedown", { e, position });
    }}
  >
    <Move class="w-3 h-3" />
  </button>
  <button
    transition:fade={{ duration: 400, easing: quintOut }}
    style:left="{position[0] + dimensions[0]}px"
    style:top="{position[1] + dimensions[1]}px"
    style:z-index={localZIndex + 1}
    class="tool -translate-x-full -translate-y-full"
    data-index={i}
    on:mousedown={(e) => {
      dispatch("resizestart", {
        e,
        dimensions,
      });
    }}
  >
    <Scaling class="w-3 h-3 -scale-x-100" />
  </button>
{/if}

<style lang="postcss">
  .tool {
    @apply absolute p-1 aspect-square;
    @apply opacity-50 bg-gray-100 border shadow-md;
    @apply rounded-full;
  }
</style>
