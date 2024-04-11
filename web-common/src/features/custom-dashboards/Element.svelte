<script lang="ts" context="module">
  import { goto } from "$app/navigation";
  import * as ContextMenu from "@rilldata/web-common/components/context-menu";
  import Chart from "@rilldata/web-common/features/custom-dashboards/Chart.svelte";
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher, onMount } from "svelte";
  import { writable } from "svelte/store";
  import ResizeHandle from "./ResizeHandle.svelte";
  import type { Vector } from "./types";

  const zIndex = writable(0);

  const options = [0, 0.5, 1];
  const allSides = options
    .flatMap((y) => options.map((x) => [x, y] as [number, number]))
    .filter(([x, y]) => !(x === 0.5 && y === 0.5));
</script>

<script lang="ts">
  const dispatch = createEventDispatcher();

  export let i: number;
  export let gapSize: number;
  export let chart: V1DashboardComponent;
  export let selected: boolean;
  export let interacting: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let radius: number;
  export let scale: number;

  let localZIndex = 0;

  $: chartName = chart.chart ?? "No chart name";

  $: finalLeft = width < 0 ? left + width : left;
  $: finalTop = height < 0 ? top + height : top;
  $: finalWidth = Math.abs(width);
  $: finalHeight = Math.abs(height);
  $: padding = gapSize;

  $: position = [finalLeft, finalTop] as Vector;
  $: dimensions = [finalWidth, finalHeight] as Vector;

  onMount(() => {
    localZIndex = $zIndex;
    zIndex.set(++localZIndex);
  });

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    localZIndex = $zIndex;
    zIndex.set(++localZIndex);
    dispatch("change", {
      e,
      dimensions: [width, height],
      position,
      changeDimensions: [0, 0],
      changePosition: [1, 1],
    });
  }
</script>

<ContextMenu.Root>
  <ContextMenu.Trigger asChild let:builder>
    <div
      {...builder}
      use:builder.action
      role="presentation"
      data-index={i}
      class="wrapper hover:cursor-pointer active:cursor-grab pointer-events-auto"
      style:z-index={localZIndex}
      style:padding="{padding}px"
      style:left="{finalLeft}px"
      style:top="{finalTop}px"
      style:width="{finalWidth}px"
      style:height="{finalHeight}px"
      on:contextmenu
      on:mousedown|capture={handleMouseDown}
    >
      <div class="size-full relative">
        {#each allSides as side}
          <ResizeHandle
            {scale}
            {i}
            {side}
            {position}
            {dimensions}
            {selected}
            on:change
          />
        {/each}

        <div
          class="size-full overflow-hidden"
          class:shadow-lg={interacting}
          style:border-radius="{radius}px"
        >
          <Chart {chartName} />
        </div>
      </div>
    </div>
  </ContextMenu.Trigger>

  <ContextMenu.Content class="z-[100]">
    <ContextMenu.Item>Copy</ContextMenu.Item>
    <ContextMenu.Item>Delete</ContextMenu.Item>
    <ContextMenu.Item
      on:click={async () => {
        await goto(`/files/charts/${chartName}`);
      }}
    >
      Go to {chartName}.yaml
    </ContextMenu.Item>
    <ContextMenu.Item>Show details</ContextMenu.Item>
  </ContextMenu.Content>
</ContextMenu.Root>

<style lang="postcss">
  .wrapper {
    @apply absolute;
  }
</style>
