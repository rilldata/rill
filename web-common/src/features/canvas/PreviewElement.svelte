<script context="module" lang="ts">
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import Component from "./Component.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let padding: number;
  export let component: V1CanvasItem;
  export let selected: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let radius: number;
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

  $: transform = `translate(${finalLeft}px, ${finalTop}px)`;

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

  function handleMouseEnter() {
    dispatch("mouseenter", { index: i });
  }

  function handleMouseLeave() {
    dispatch("mouseleave", { index: i });
  }
</script>

{#if componentName && !inlineComponent}
  <div
    class="component absolute"
    role="presentation"
    data-component-index={i}
    style:width="{finalWidth}px"
    style:height="{finalHeight}px"
    style:padding="{padding}px"
    style:transform
    style:will-change="transform"
    on:dragstart={handleDragStart}
    on:dragend={handleDragEnd}
    on:dragover={onDragOver}
    on:drop={onDrop}
    on:mousedown={handleMouseDown}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
  >
    <Component
      {instanceId}
      {i}
      {componentName}
      {padding}
      {radius}
      {selected}
      {rowIndex}
      {columnIndex}
      builders={undefined}
      height={finalHeight}
      left={0}
      top={0}
      width={finalWidth}
    />
  </div>
{:else if componentName}
  <div
    class="component absolute"
    role="presentation"
    data-component-index={i}
    style:width="{finalWidth}px"
    style:height="{finalHeight}px"
    style:padding="{padding}px"
    style:transform
    style:will-change="transform"
    on:dragstart={handleDragStart}
    on:dragend={handleDragEnd}
    on:dragover={onDragOver}
    on:drop={onDrop}
    on:mousedown={handleMouseDown}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
  >
    <Component
      {instanceId}
      {i}
      {componentName}
      {padding}
      {radius}
      {selected}
      {rowIndex}
      {columnIndex}
      builders={undefined}
      height={finalHeight}
      left={0}
      top={0}
      width={finalWidth}
    />
  </div>
{/if}

<style lang="postcss">
  .component {
    touch-action: none;
    transform-origin: 0 0;
    left: 0;
    top: 0;
  }
</style>
