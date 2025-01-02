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
  import { getRowIndex, getColumnIndex } from "./util";
  import { Grid, groupItemsByRow, isValidItem } from "./grid";
  import type { DropPosition } from "./types";
  import type { RowGroup } from "./types";

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
    position: DropPosition;
  } | null = null;
  let hoveredIndex: number | null = null;
  let resizingRow: {
    index: number;
    startY: number;
    initialHeight: number;
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: extraLeftPadding = !$navigationOpen;

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;

  $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;
  $: radius = gridCell * defaults.COMPONENT_RADIUS;

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);

  $: console.log("[CanvasDashboardPreview] items updated:", items);

  const dispatch = createEventDispatcher();
  const { canvasName } = getCanvasStateManagers();

  $: itemsByRow = groupItemsByRow(items);

  const grid = new Grid(items);

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
    $canvasStore.setSelectedComponentIndex(selectedIndex);
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

    // Don't show drop target line if dragging over self
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
    draggedComponent = null;
    dropTarget = null;
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

  function getDropPosition(e: DragEvent, targetIndex: number): DropPosition {
    const targetElement = document.querySelector(
      `[data-component-index="${targetIndex}"]`,
    );
    if (!targetElement) return "left";

    const rect = targetElement.getBoundingClientRect();
    return grid.getDropPosition(e.clientX, e.clientY, rect);
  }

  function handleDrop(e: DragEvent | CustomEvent<DragEvent>) {
    e.preventDefault();
    e.stopPropagation();

    if (!draggedComponent || !dropTarget) return;

    const { index: dragIndex } = draggedComponent;
    const { index: dropIndex, position } = dropTarget;
    const targetItem = items[dropIndex];
    const draggedItem = items[dragIndex];

    console.log("[CanvasDashboardPreview] handleDrop", {
      position,
      targetItem,
      draggedItem,
    });

    if (!isValidItem(targetItem) || !isValidItem(draggedItem)) return;

    const { items: newItems, insertIndex } = grid.moveItem(
      draggedItem,
      targetItem,
      position,
      dragIndex,
    );

    items = newItems;

    dispatch("update", {
      index: dragIndex,
      position: [newItems[insertIndex]?.x, newItems[insertIndex]?.y],
      dimensions: [newItems[insertIndex]?.width, newItems[insertIndex]?.height],
      items: newItems,
    });

    if (selectedIndex === dragIndex) {
      selectedIndex = insertIndex;
      $canvasStore.setSelectedComponentIndex(insertIndex);
    }

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

  function handleRowResizeStart(
    e: MouseEvent,
    rowIndex: number,
    currentHeight: number,
  ) {
    e.preventDefault();
    document.body.classList.add("resizing-row");
    resizingRow = {
      index: rowIndex,
      startY: e.clientY,
      initialHeight: currentHeight,
    };
    console.log("[CanvasDashboardPreview] Starting resize of row:", {
      rowIndex,
      currentHeight,
      totalRows: itemsByRow.length,
      affectedRows: itemsByRow.slice(rowIndex + 1), // Rows that will move
    });
  }

  function handleRowResize(e: MouseEvent) {
    if (!resizingRow) return;

    const deltaY = e.clientY - resizingRow.startY;
    const newHeight = Math.round(
      Math.max(defaults.MIN_ROW_HEIGHT, resizingRow.initialHeight + deltaY),
    );

    const row = itemsByRow[resizingRow.index];
    if (!row) return;

    // Create a copy of items for DOM updates
    const updatedItems = [...items];
    const rowItems = updatedItems.filter((item) => item.y === row.y);

    // Update height for all items in the row
    rowItems.forEach((item) => {
      item.height = Math.round(newHeight / defaults.GRID_CELL_SIZE);
    });

    // Update positions of all rows below
    let currentY = 0;
    const updatedRows = groupItemsByRow(updatedItems);
    updatedRows.forEach((currentRow, idx) => {
      if (!resizingRow) return;
      currentRow.items.forEach((item) => {
        item.y = Math.round(currentY / defaults.GRID_CELL_SIZE);
      });
      // Use the height from the current row being processed
      const rowHeight =
        idx === resizingRow.index
          ? newHeight
          : currentRow.height * defaults.GRID_CELL_SIZE;
      currentY += rowHeight;
    });

    // Update the DOM immediately
    items = updatedItems;
  }

  function handleRowResizeEnd() {
    if (resizingRow) {
      // Only dispatch update to save YAML when resize ends
      dispatch("update", {
        index: -1,
        items,
        position: [0, 0],
        dimensions: [0, 0],
      });
    }
    document.body.classList.remove("resizing-row");
    resizingRow = null;
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
  <div class="grid auto-rows-min w-full gap-0">
    {#each itemsByRow as row, index (index)}
      <div
        class="row w-full"
        data-row-index={index}
        style="height: {row.height * gridCell}px;"
      >
        {#each row.items as component}
          {@const i = items.indexOf(component)}
          <!-- FIXME: padding 16 -->
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
            height={Number(component.height ?? defaults.COMPONENT_HEIGHT) *
              gridCell}
            top={0}
            left={Math.min(
              Number(component.x ?? 0),
              defaults.COLUMN_COUNT -
                (component.width ?? defaults.COMPONENT_WIDTH),
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
      </div>

      {#if index < itemsByRow.length - 1}
        <button
          type="button"
          aria-label="Resize row"
          class="row-resize-handle w-full h-[3px] cursor-row-resize bg-transparent hover:bg-primary-300 z-[50] opacity-0 hover:opacity-100 pointer-events-auto"
          on:mousedown|stopPropagation={(e) =>
            handleRowResizeStart(e, index, row.height * gridCell)}
        />
      {/if}
    {/each}
  </div>

  {#if dropTarget && draggedComponent}
    {@const targetItem = items[dropTarget.index]}
    {#if targetItem && targetItem.x !== undefined && targetItem.y !== undefined && targetItem.width !== undefined && targetItem.height !== undefined}
      <DropTargetLine
        height={dropTarget.position === "bottom" ||
        dropTarget.position === "row" ||
        dropTarget.position === "top"
          ? 2
          : targetItem.height * gridCell}
        top={dropTarget.position === "bottom"
          ? (targetItem.y + targetItem.height) * gridCell
          : dropTarget.position === "top"
            ? targetItem.y * gridCell
            : dropTarget.position === "row"
              ? targetItem.y * gridCell
              : targetItem.y * gridCell}
        left={dropTarget.position === "right"
          ? (targetItem.x + targetItem.width) * gridCell
          : dropTarget.position === "bottom"
            ? 0
            : dropTarget.position === "top" || dropTarget.position === "row"
              ? targetItem.x * gridCell
              : targetItem.x * gridCell}
        width={dropTarget.position === "bottom"
          ? defaults.DASHBOARD_WIDTH
          : undefined}
        orientation={dropTarget.position === "bottom" ||
        dropTarget.position === "top" ||
        dropTarget.position === "row"
          ? "horizontal"
          : "vertical"}
      />
    {/if}
  {/if}
</DashboardWrapper>

<svelte:window on:mousemove={handleRowResize} on:mouseup={handleRowResizeEnd} />

<style>
  :global(body.resizing-row) {
    cursor: row-resize !important;
  }

  .row {
    position: relative;
    min-height: 0;
  }

  .row-resize-handle {
    position: relative;
    transition: opacity 0.2s;
  }
</style>
