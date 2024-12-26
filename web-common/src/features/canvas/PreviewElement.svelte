<script context="module" lang="ts">
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import Component from "./Component.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let gapSize: number;
  export let component: V1CanvasItem;
  export let selected: boolean;
  export let interacting: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let radius: number;
  export let scale: number;
  export let instanceId: string;
  export let onDragOver: (e: CustomEvent<DragEvent> | DragEvent) => void;
  export let onDrop: (e: CustomEvent<DragEvent> | DragEvent) => void;
  export let rowIndex: number;
  export let columnIndex: number;

  $: componentName = component?.component;
  $: inlineComponent = component?.definedInCanvas;

  $: finalLeft = width < 0 ? left + width : left;
  $: finalTop = height < 0 ? top + height : top;
  $: finalWidth = Math.abs(width);
  $: finalHeight = Math.abs(height);
  $: padding = gapSize;

  const dispatch = createEventDispatcher();

  function handleMouseDown(e: MouseEvent) {
    // if (e.button !== 0) return;
    dispatch("change", {
      e,
      dimensions: [width, height],
      position: [finalLeft, finalTop],
      changeDimensions: [0, 0],
      changePosition: [1, 1],
    });
  }

  function handleDragStart(e: DragEvent) {
    console.log("[PreviewElement] handleDragStart", { i, width, height });

    dispatch("dragstart", {
      componentIndex: i,
      width,
      height,
    });
  }

  function handleDragEnd() {
    console.log("[PreviewElement] handleDragEnd");
    dispatch("dragend");
  }
</script>

{#if componentName && !inlineComponent}
  <Component
    {instanceId}
    {i}
    {interacting}
    {componentName}
    {padding}
    {radius}
    {selected}
    {rowIndex}
    {columnIndex}
    builders={undefined}
    height={finalHeight}
    left={finalLeft}
    top={finalTop}
    width={finalWidth}
    on:dragstart={handleDragStart}
    on:dragend={handleDragEnd}
    on:dragover={onDragOver}
    on:drop={onDrop}
    on:mousedown={handleMouseDown}
  />
{:else if componentName}
  <Component
    {instanceId}
    {i}
    {interacting}
    {componentName}
    {padding}
    {radius}
    {selected}
    {rowIndex}
    {columnIndex}
    builders={undefined}
    height={finalHeight}
    left={finalLeft}
    top={finalTop}
    width={finalWidth}
    on:dragstart={handleDragStart}
    on:dragend={handleDragEnd}
    on:dragover={onDragOver}
    on:drop={onDrop}
    on:mousedown={handleMouseDown}
  />
{/if}
