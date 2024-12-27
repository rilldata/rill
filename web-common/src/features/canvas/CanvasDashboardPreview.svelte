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
  import DropTargetLine from "./DropTargetLine.svelte";
  import {
    validateItemPositions,
    isValidItem,
    groupItemsByRow,
    leftAlignRow,
    vector,
    getRowIndex,
    getColumnIndex,
  } from "./util";

  export let items: V1CanvasItem[];
  export let selectedIndex: number | null = null;

  const { canvasStore } = getCanvasStateManagers();

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
  let hoveredIndex: number | null = null;

  $: ({ instanceId } = $runtime);

  $: extraLeftPadding = !$navigationOpen;

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
    const componentIndex = Number(
      e.detail.e.currentTarget.dataset.componentIndex,
    );
    selectedIndex = componentIndex;
    canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
  }

  function handleDragStart(e: CustomEvent) {
    const { componentIndex, width, height } = e.detail;
    console.log("[CanvasDashboardPreview] DragStart: ", {
      componentIndex,
      width,
      height,
    });
    hoveredIndex = null;
    draggedComponent = {
      index: componentIndex,
      width,
      height,
    };
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

  function handleDragEnd() {
    if (!dropTarget) {
      draggedComponent = null;
    }
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
      `[data-component-index="${targetIndex}"]`,
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

  function handleDrop(e: DragEvent | CustomEvent<DragEvent>) {
    e.preventDefault();
    e.stopPropagation();

    if (!draggedComponent || !dropTarget) return;

    const { index: dragIndex } = draggedComponent;
    const { index: dropIndex, position } = dropTarget;
    const targetItem = items[dropIndex];
    const draggedItem = items[dragIndex];

    console.log("[CanvasDashboardPreview] Drag and Drop:", {
      dragged: {
        index: dragIndex,
        item: draggedItem,
      },
      target: {
        index: dropIndex,
        item: targetItem,
        position,
      },
    });

    if (!isValidItem(targetItem) || !isValidItem(draggedItem)) return;

    // Group items by row before modification
    const rows = groupItemsByRow([...items]);
    const newItems = [...items];

    // Create a deep copy of the dragged item to preserve all properties
    const [draggedItemFull] = newItems.splice(dragIndex, 1);
    const removedItem = {
      ...draggedItemFull,
      x: draggedItemFull.x,
      y: draggedItemFull.y,
      width: draggedItemFull.width,
      height: draggedItemFull.height,
    };
    let insertIndex = dropIndex;

    switch (position) {
      case "bottom": {
        console.log("[CanvasDashboardPreview] Dropping bottom");
        // Create new row
        const newY = targetItem.y + targetItem.height;
        removedItem.y = newY;
        removedItem.x = 0;
        removedItem.width = defaults.COLUMN_COUNT;
        removedItem.height = draggedItem.height;
        insertIndex = dropIndex + 1;

        rows.push({
          y: newY,
          height: removedItem.height,
          items: [removedItem],
        });
        break;
      }

      // FIXME: when dropping right, the dragged item gets placed in the last index of the row
      case "right": {
        console.log("[CanvasDashboardPreview] Dropping right");
        const targetRow = rows.find((row) => row.y === targetItem.y);
        if (targetRow) {
          // Get all items in this row from the original items array
          const rowItems = items
            .filter((item) => item.y === targetItem.y)
            .filter((item) => item !== draggedItem)
            .sort((a, b) => (a.x ?? 0) - (b.x ?? 0));

          // Calculate new position using vector
          const newPosition: Vector = vector.add(
            [targetItem.x, targetItem.y],
            [targetItem.width, 0],
          );

          // Find insert position based on x coordinate
          const nextItemIndex = rowItems.findIndex(
            (item) => (item.x ?? 0) > newPosition[0],
          );

          // Calculate the actual index in the full items array
          insertIndex = dropIndex + 1;

          // Set position properties using vector
          removedItem.x = newPosition[0];
          removedItem.y = newPosition[1];
          removedItem.width = removedItem.width;
          removedItem.height = draggedItem.height;

          targetRow.items.push(removedItem);
          targetRow.height = Math.max(targetRow.height, removedItem.height);
        }
        break;
      }

      case "left": {
        console.log("[CanvasDashboardPreview] Dropping left");
        const targetRow = rows.find((row) => row.y === targetItem.y);
        if (targetRow) {
          // Insert before target
          removedItem.x = targetItem.x;
          removedItem.y = targetItem.y;
          removedItem.width = removedItem.width;
          removedItem.height = draggedItem.height;

          targetRow.items.push(removedItem);
          targetRow.height = Math.max(targetRow.height, removedItem.height);

          insertIndex = dropIndex;
        }
        break;
      }

      default: {
        console.warn(
          "[CanvasDashboardPreview] Unknown drop position:",
          position,
        );
        return;
      }
    }

    // Reinsert the item into the array
    newItems.splice(insertIndex, 0, removedItem);

    // Validate item positions
    validateItemPositions(newItems);

    // Update items
    items = newItems;

    // Dispatch update once after position is set
    dispatch("update", {
      index: insertIndex,
      position: [removedItem.x, removedItem.y],
      dimensions: [removedItem.width, removedItem.height],
      items: newItems,
    });

    // Update selected index to follow the dropped item
    if (selectedIndex === dragIndex) {
      selectedIndex = insertIndex;
      canvasStore.setSelectedComponentIndex($canvasName, insertIndex);
    }

    // Reset drop target and dragged component
    dropTarget = null;
    draggedComponent = null;
  }

  function handleMouseEnter(e: CustomEvent<{ index: number }>) {
    if (draggedComponent) {
      // Don't update hover state while dragging
      return;
    }
    hoveredIndex = e.detail.index;
    console.log("[CanvasDashboardPreview] Component hovered:", hoveredIndex);
  }

  function handleMouseLeave(e: CustomEvent<{ index: number }>) {
    if (draggedComponent) {
      // Don't update hover state while dragging
      return;
    }
    if (hoveredIndex === e.detail.index) {
      hoveredIndex = null;
      console.log("[CanvasDashboardPreview] Component unhovered");
    }
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
  {scale}
  height={maxBottom * gridCell * scale}
  width={defaults.DASHBOARD_WIDTH}
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
      padding={16}
      rowIndex={getRowIndex(component, items)}
      columnIndex={getColumnIndex(component, items)}
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
      onDragOver={(e) =>
        handleDragOver(e instanceof CustomEvent ? e.detail : e, i)}
      onDrop={(e) => handleDrop(e instanceof CustomEvent ? e.detail : e)}
      on:dragstart={handleDragStart}
      on:dragend={handleDragEnd}
      on:change={handleChange}
      on:mouseenter={handleMouseEnter}
      on:mouseleave={handleMouseLeave}
    />
  {/each}

  {#if dropTarget && draggedComponent}
    {@const targetItem = items[dropTarget.index]}
    {#if targetItem && targetItem.x !== undefined && targetItem.y !== undefined && targetItem.width !== undefined && targetItem.height !== undefined}
      <DropTargetLine
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
