<script lang="ts">
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createEventDispatcher } from "svelte";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";
  import PreviewElement from "./PreviewElement.svelte";
  import type { Vector } from "./types";
  import GhostLine from "./GhostLine.svelte";
  import {
    moveItemToNewRow,
    recalculateRowPositions,
    validateItemPositions,
    isValidItem,
  } from "./util";

  export let items: V1CanvasItem[];
  export let selectedIndex: number | null = null;
  export let showFilterBar = true;

  const { canvasEntity } = getCanvasStateManagers();

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;
  let draggedComponent: {
    index: number;
    width: number;
    height: number;
  } | null = null;
  let dropTarget: {
    index: number;
    position: "left" | "right" | "bottom";
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;

  $: gapSize = defaults.DASHBOARD_WIDTH * (defaults.GAP_SIZE / 1000);
  $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: console.log("[CanvasDashboardPreview] items updated:", items);

  const dispatch = createEventDispatcher();
  const { canvasName } = getCanvasStateManagers();

  function handleChange(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
      dimensions: Vector;
      position: Vector;
      changeDimensions: [0 | 1 | -1, 0 | 1 | -1];
      changePosition: [0 | 1, 0 | 1];
    }>,
  ) {
    e.preventDefault();
    const index = Number(e.detail.e.currentTarget.dataset.index);
    selectedIndex = index;
    canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
  }

  function handleDragStart(e: CustomEvent) {
    const { componentIndex, width, height } = e.detail;
    console.log("[CanvasDashboardPreview] DragStart: ", {
      componentIndex,
      width,
      height,
    });
    draggedComponent = {
      index: componentIndex,
      width,
      height,
    };
  }

  function handleDragEnd() {
    draggedComponent = null;
  }

  function handleScroll(
    e: UIEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    scrollOffset = e.currentTarget.scrollTop;
  }

  function handleDeselect() {
    selectedIndex = null;
    canvasEntity.setSelectedComponentIndex(selectedIndex);
  }

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);

  function getDropPosition(
    e: DragEvent,
    targetIndex: number,
  ): "left" | "right" | "bottom" {
    const targetElement = document.querySelector(
      `[data-index="${targetIndex}"]`,
    );
    if (!targetElement) return "left";

    const rect = targetElement.getBoundingClientRect();
    const mouseX = e.clientX;
    const mouseY = e.clientY;

    // Define bottom zone as the lower 25% of the element
    const bottomZone = rect.bottom - rect.height * 0.25;

    // Check if mouse is in bottom zone first
    if (mouseY > bottomZone) {
      return "bottom";
    }

    // If not in bottom zone, determine left/right as before
    return mouseX > rect.left + rect.width / 2 ? "right" : "left";
  }

  function handleDragOver(e: DragEvent, targetIndex: number) {
    e.preventDefault();
    e.stopPropagation();

    // Don't show ghost line if dragging over self
    if (draggedComponent?.index === targetIndex) {
      dropTarget = null;
      return;
    }

    const position = getDropPosition(e, targetIndex);
    console.log("[CanvasDashboardPreview] handleDragOver", {
      targetIndex,
      position,
      mouseX: e.clientX,
      mouseY: e.clientY,
    });

    dropTarget = { index: targetIndex, position };
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();

    if (!draggedComponent || !dropTarget) return;

    const { index: dragIndex } = draggedComponent;
    const { index: dropIndex, position } = dropTarget;
    const targetItem = items[dropIndex];

    // Create new array and remove dragged item
    const newItems = [...items];
    const [draggedItem] = newItems.splice(dragIndex, 1);
    let insertIndex: number;

    if (position === "bottom") {
      if (!isValidItem(targetItem)) return;
      // Insert into new row below target
      moveItemToNewRow(draggedItem, targetItem?.y, targetItem.height);
      draggedItem.width = Math.min(
        draggedComponent.width / gridCell,
        defaults.COLUMN_COUNT,
      );
      draggedItem.height = draggedComponent.height / gridCell;
      insertIndex = dropIndex + 1;
      console.log(
        "[CanvasDashboardPreview] handleDrop: moving item to new row",
      );
    } else {
      // Insert into same row
      draggedItem.y = targetItem.y;
      insertIndex = dropIndex;
      newItems.splice(insertIndex, 0, draggedItem);
      recalculateRowPositions(newItems, targetItem.y);
    }

    newItems.splice(insertIndex, 0, draggedItem);
    validateItemPositions(newItems);
    items = newItems;

    dispatch("update", {
      index: insertIndex,
      position: [newItems[insertIndex]?.x, newItems[insertIndex]?.y],
      dimensions:
        position === "bottom"
          ? [defaults.COLUMN_COUNT, draggedItem.height]
          : [draggedComponent.width, draggedComponent.height],
      items: newItems,
    });

    dropTarget = null;
    draggedComponent = null;
  }
</script>

<!-- <svelte:window on:mousemove={handleMouseMove} on:mouseup={handleMouseUp} /> -->

{#if showFilterBar}
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
  >
    <CanvasFilters />
  </div>
{/if}

<DashboardWrapper
  bind:contentRect
  {scale}
  {showGrid}
  height={maxBottom * gridCell * scale}
  width={defaults.DEFAULT_DASHBOARD_WIDTH}
  on:click={handleDeselect}
  on:scroll={handleScroll}
  on:dragover={(e) => {
    e.preventDefault();
  }}
  on:drop={handleDrop}
>
  <section
    class="flex relative justify-between gap-x-4 py-4 pb-6 px-4"
  ></section>
  {#each items as component, i (i)}
    <PreviewElement
      {instanceId}
      {i}
      {scale}
      {component}
      {radius}
      selected={selectedIndex === i}
      interacting={false}
      {gapSize}
      width={Math.min(
        Number(component.width ?? defaults.COMPONENT_WIDTH),
        defaults.COLUMN_COUNT,
      ) * gridCell}
      height={Number(component.height ?? defaults.COMPONENT_HEIGHT) * gridCell}
      top={Number(component.y) * gridCell}
      left={Math.min(
        Number(component.x ?? 0),
        defaults.COLUMN_COUNT - (component.width ?? defaults.COMPONENT_WIDTH),
      ) * gridCell}
      onDragOver={(e) => handleDragOver(e, i)}
      onDrop={(e) => handleDrop(e)}
      on:dragstart={handleDragStart}
      on:dragend={handleDragEnd}
      on:change={handleChange}
    />
  {/each}

  {#if dropTarget && draggedComponent}
    {@const targetItem = items[dropTarget.index]}
    {#if targetItem && targetItem.x !== undefined && targetItem.y !== undefined && targetItem.width !== undefined && targetItem.height !== undefined}
      <GhostLine
        height={dropTarget.position === "bottom"
          ? 2
          : targetItem.height * gridCell}
        top={dropTarget.position === "bottom"
          ? (targetItem.y + targetItem.height) * gridCell
          : targetItem.y * gridCell}
        left={dropTarget.position === "right"
          ? (targetItem.x + targetItem.width) * gridCell
          : dropTarget.position === "bottom"
            ? targetItem.x * gridCell
            : targetItem.x * gridCell}
        orientation={dropTarget.position === "bottom"
          ? "horizontal"
          : "vertical"}
      />
    {/if}
  {/if}
</DashboardWrapper>
