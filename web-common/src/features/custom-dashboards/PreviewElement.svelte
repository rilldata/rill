<script context="module" lang="ts">
  import { goto } from "$app/navigation";
  import * as ContextMenu from "@rilldata/web-common/components/context-menu";
  import { V1DashboardItem } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher, onMount } from "svelte";
  import { writable } from "svelte/store";
  import Component from "./Component.svelte";

  const zIndex = writable(0);
</script>

<script lang="ts">
  const dispatch = createEventDispatcher();

  export let i: number;
  export let gapSize: number;
  export let component: V1DashboardItem;
  export let selected: boolean;
  export let interacting: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let radius: number;
  export let scale: number;
  export let instanceId: string;

  let localZIndex = 0;

  $: componentName = component?.component;

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
    if (e.button !== 0) return;
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

{#if componentName}
  <ContextMenu.Root>
    <ContextMenu.Trigger asChild let:builder>
      <Component
        {instanceId}
        {i}
        {localZIndex}
        {interacting}
        {componentName}
        {padding}
        {radius}
        {scale}
        {selected}
        fontSize={component.fontSize}
        builders={[builder]}
        height={finalHeight}
        left={finalLeft}
        top={finalTop}
        width={finalWidth}
        on:change
        on:contextmenu
        on:mousedown={handleMouseDown}
      />
    </ContextMenu.Trigger>

    <ContextMenu.Content class="z-[100]">
      <ContextMenu.Item>Copy</ContextMenu.Item>
      <ContextMenu.Item>Delete</ContextMenu.Item>
      <ContextMenu.Item
        on:click={async () => {
          await goto(`/files/charts/${componentName}`);
        }}
      >
        Go to {componentName}.yaml
      </ContextMenu.Item>
      <ContextMenu.Item>Show details</ContextMenu.Item>
    </ContextMenu.Content>
  </ContextMenu.Root>
{/if}
