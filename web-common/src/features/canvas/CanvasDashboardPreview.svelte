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
  import { vector } from "./util";
  import GhostLine from "./GhostLine.svelte";

  const dispatch = createEventDispatcher();
  const zeroVector = [0, 0] as [0, 0];

  export let items: V1CanvasItem[];
  export let selectedIndex: number | null = null;
  export let showFilterBar = true;

  const { canvasEntity } = getCanvasStateManagers();

  let snap = true;
  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;
  let changing = false;
  let startMouse: Vector = [0, 0];
  let mousePosition: Vector = [0, 0];
  let initialElementDimensions: Vector = [0, 0];
  let initialElementPosition: Vector = [0, 0];
  let dimensionChange: [0 | 1 | -1, 0 | 1 | -1] = [0, 0];
  let positionChange: [0 | 1, 0 | 1] = [0, 0];
  let draggedComponent: {
    index: number;
    width: number;
    height: number;
  } | null = null;
  let dropTarget: {
    index: number;
    position: "left" | "right";
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;

  $: gapSize = defaults.DASHBOARD_WIDTH * (defaults.GAP_SIZE / 1000);
  $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;
  $: gridVector = [gridCell, gridCell] as Vector;

  $: mouseDelta = vector.divide(vector.subtract(mousePosition, startMouse), [
    scale,
    scale,
  ]);

  $: dragPosition = vector.add(
    vector.multiply(mouseDelta, positionChange),
    initialElementPosition,
  );

  $: resizeDimenions = vector.add(
    vector.multiply(mouseDelta, dimensionChange),
    initialElementDimensions,
  );

  $: console.log("[CanvasDashboardPreview] items updated:", items);

  // function handleMouseUp() {
  //   console.log("[CanvasDashboardPreview] handleMouseUp ", selectedIndex);

  //   if (selectedIndex === null || !changing) return;

  //   const cellPosition = getCell(dragPosition, true);
  //   const dimensions = getCell(resizeDimenions, true);

  //   items[selectedIndex].x = Math.max(
  //     0,
  //     dimensions[0] < 0 ? cellPosition[0] + dimensions[0] : cellPosition[0],
  //   );
  //   items[selectedIndex].y = Math.max(
  //     0,
  //     dimensions[1] < 0 ? cellPosition[1] + dimensions[1] : cellPosition[1],
  //   );

  //   items[selectedIndex].width = Math.max(1, Math.abs(dimensions[0]));
  //   items[selectedIndex].height = Math.max(1, Math.abs(dimensions[1]));

  //   dispatch("update", {
  //     index: selectedIndex,
  //     position: [items[selectedIndex].x, items[selectedIndex].y],
  //     dimensions: [items[selectedIndex].width, items[selectedIndex].height],
  //   });

  //   reset();
  // }

  function reset() {
    changing = false;
    mousePosition =
      startMouse =
      startMouse =
      initialElementPosition =
      initialElementDimensions =
      dimensionChange =
      positionChange =
      resizeDimenions =
        zeroVector;
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

  // function handleMouseMove(e: MouseEvent) {
  //   // No-op - we'll use drag events instead
  // }

  function handleScroll(
    e: UIEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    scrollOffset = e.currentTarget.scrollTop;
  }

  function getCell(rawVector: Vector, snap: boolean): Vector {
    const raw = vector.divide(rawVector, gridVector);

    if (!snap) return raw;

    return [Math.round(raw[0]), Math.round(raw[1])];
  }

  function deselect() {
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
  ): "left" | "right" {
    const targetElement = document.querySelector(
      `[data-index="${targetIndex}"]`,
    );
    if (!targetElement) return "left";

    const rect = targetElement.getBoundingClientRect();
    const mouseX = e.clientX;

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
    console.log("[DragOver]", {
      targetIndex,
      position,
      mouseX: e.clientX,
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

    if (targetItem.x === undefined || targetItem.width === undefined) return;

    // Create new array and remove dragged item
    const newItems = [...items];
    const [draggedItem] = newItems.splice(dragIndex, 1);

    // Calculate insert position based on drop position
    const insertIndex = position === "right" ? dropIndex : dropIndex;

    // Insert dragged item at new position
    newItems.splice(insertIndex, 0, draggedItem);

    // Recalculate x positions for all items
    newItems.forEach((item, index) => {
      if (index === 0) {
        item.x = 0;
      } else {
        const prevItem = newItems[index - 1];
        if (prevItem.x === undefined || prevItem.width === undefined) return;
        // Round the x position to ensure integer values
        item.x = Math.round(
          prevItem.x + prevItem.width + defaults.GAP_SIZE / 1000,
        );
      }
      // Keep same y position
      item.y = targetItem.y;
    });

    items = newItems;

    dispatch("update", {
      index: insertIndex,
      position: [newItems[insertIndex].x, newItems[insertIndex].y],
      dimensions: [draggedComponent.width, draggedComponent.height],
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
  {changing}
  {gapSize}
  {gridCell}
  {scrollOffset}
  {radius}
  {scale}
  {showGrid}
  height={maxBottom * gridCell * scale}
  width={defaults.DEFAULT_DASHBOARD_WIDTH}
  on:click={deselect}
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
      selected={draggedComponent?.index === i}
      interacting={false}
      {gapSize}
      width={Number(component.width ?? defaults.COMPONENT_WIDTH) * gridCell}
      height={Number(component.height ?? defaults.COMPONENT_HEIGHT) * gridCell}
      top={Number(component.y) * gridCell}
      left={Number(component.x) * gridCell}
      onDragOver={(e) => handleDragOver(e, i)}
      onDrop={(e) => handleDrop(e)}
      on:dragstart={handleDragStart}
      on:dragend={handleDragEnd}
    />
  {/each}

  {#if dropTarget && draggedComponent}
    {@const targetItem = items[dropTarget.index]}
    {#if targetItem && targetItem.x !== undefined && targetItem.y !== undefined && targetItem.width !== undefined && targetItem.height !== undefined}
      <GhostLine
        height={targetItem.height * gridCell}
        top={targetItem.y * gridCell}
        left={dropTarget.position === "right"
          ? (targetItem.x + targetItem.width) * gridCell
          : targetItem.x * gridCell}
        orientation="vertical"
      />
    {/if}
  {/if}
</DashboardWrapper>
