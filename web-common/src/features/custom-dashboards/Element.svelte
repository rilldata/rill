<script lang="ts" context="module">
  import { goto } from "$app/navigation";
  import * as ContextMenu from "@rilldata/web-common/components/context-menu";
  import Chart from "@rilldata/web-common/features/custom-dashboards/Chart.svelte";
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher, onMount } from "svelte";
  import { writable } from "svelte/store";
  import Component from "./Component.svelte";

  const zIndex = writable(0);
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
  export let chartView = false;

  let localZIndex = 0;

  $: chartName = chart.chart ?? "No chart name";

  $: finalLeft = width < 0 ? left + width : left;
  $: finalTop = height < 0 ? top + height : top;
  $: finalWidth = Math.abs(width);
  $: finalHeight = Math.abs(height);
  $: padding = gapSize;

  onMount(() => {
    localZIndex = $zIndex;
    zIndex.set(++localZIndex);
  });

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0 || chartView) return;
    localZIndex = $zIndex;
    zIndex.set(++localZIndex);
    dispatch("change", {
      e,
      dimensions: [width, height],
      position: [finalLeft, finalTop],
      changeDimensions: [0, 0],
      changePosition: [1, 1],
    });
  }
</script>

<ContextMenu.Root>
  <ContextMenu.Trigger asChild let:builder>
    <Component
      {chartView}
      builders={[builder]}
      left={finalLeft}
      top={finalTop}
      {padding}
      {scale}
      {radius}
      {selected}
      {interacting}
      width={finalWidth}
      height={finalHeight}
      {i}
      on:mousedown={handleMouseDown}
      on:contextmenu
      on:change
    >
      <Chart {chartName} />
    </Component>
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
