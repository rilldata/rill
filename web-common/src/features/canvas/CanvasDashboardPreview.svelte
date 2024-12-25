<script lang="ts">
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
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

  const { canvasStore } = getCanvasStateManagers();

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
    position: "left" | "right" | "bottom";
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: extraLeftPadding = !$navigationOpen;

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
    dimensionChange = e.detail.changeDimensions;
    positionChange = e.detail.changePosition;
    const index = Number(e.detail.e.currentTarget.dataset.index);
    initialElementDimensions = e.detail.dimensions;
    initialElementPosition = e.detail.position;
    startMouse = [
      e.detail.e.clientX - contentRect.left,
      e.detail.e.clientY - contentRect.top - scrollOffset,
    ];
    mousePosition = startMouse;
    selectedIndex = index;
    canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
    changing = true;
  }

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
    $canvasStore.setSelectedComponentIndex(selectedIndex);
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
    console.log("[DragOver]", {
      targetIndex,
      position,
      mouseX: e.clientX,
      mouseY: e.clientY,
    });

    dropTarget = { index: targetIndex, position };
  }

  function isValidItem(item: V1CanvasItem): boolean {
    return (
      item.x !== undefined &&
      item.y !== undefined &&
      item.width !== undefined &&
      item.height !== undefined
    );
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    e.stopPropagation();

    if (!draggedComponent || !dropTarget) return;

    const { index: dragIndex } = draggedComponent;
    const { index: dropIndex, position } = dropTarget;
    const targetItem = items[dropIndex];

    if (!isValidItem(targetItem)) return;

    // Create new array and remove dragged item
    const newItems = [...items];
    const [draggedItem] = newItems.splice(dragIndex, 1);
    let insertIndex: number;

    if (position === "bottom") {
      draggedItem.y = targetItem.y + targetItem.height;
      draggedItem.x = 0;
      draggedItem.width = defaults.COLUMN_COUNT;
      insertIndex = dropIndex + 1;
      newItems.splice(insertIndex, 0, draggedItem);
    } else {
      draggedItem.y = targetItem.y;
      insertIndex = position === "right" ? dropIndex : dropIndex;
      newItems.splice(insertIndex, 0, draggedItem);

      // Recalculate x positions for all items in the row
      let currentX = 0;
      newItems.forEach((item, index) => {
        if (item.y === targetItem.y) {
          // Only adjust items in same row
          if (index === 0 || newItems[index - 1].y !== item.y) {
            item.x = 0;
            currentX = item.width;
          } else {
            // Check if there's enough space in the row
            const newX = Math.round(currentX + defaults.GAP_SIZE / 1000);

            // Ensure x position never exceeds grid bounds
            if (newX + item.width > defaults.COLUMN_COUNT) {
              item.y = targetItem.y + targetItem.height;
              item.x = 0;
              currentX = item.width;
            } else {
              item.x = Math.min(newX, defaults.COLUMN_COUNT - item.width);
              currentX = item.x + item.width;
            }
          }
        }
      });
    }

    // Validate all x positions one final time
    newItems.forEach((item) => {
      if (item.x !== undefined && item.width !== undefined) {
        item.x = Math.min(
          Math.max(0, item.x),
          defaults.COLUMN_COUNT - item.width,
        );
      }
    });

    items = newItems;

    dispatch("update", {
      index: insertIndex,
      position: [newItems[insertIndex].x, newItems[insertIndex].y],
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

<div
  id="header"
  class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
  class:left-shift={extraLeftPadding}
>
  <CanvasFilters />
</div>

<DashboardWrapper
  bind:contentRect
  {changing}
  {gapSize}
  {gridCell}
  {scrollOffset}
  {radius}
  {scale}
  height={maxBottom * gridCell * scale}
  width={defaults.DASHBOARD_WIDTH}
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
        width={dropTarget.position === "bottom"
          ? targetItem.width * gridCell
          : 2}
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
