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
  import DropIndicator from "./DropIndicator.svelte";
  import { getRowIndex, getColumnIndex, redistributeRowColumns } from "./util";
  import { Grid, groupItemsByRow, isValidItem } from "./grid";
  import type { DropPosition } from "./types";

  export let items: V1CanvasItem[];
  export let selectedIndex: number | null = null;

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
  let resizingCol: {
    index: number;
    startX: number;
    initialWidth: number;
    maxWidth: number;
  } | null = null;
  let hideTimeout: ReturnType<typeof setTimeout>;
  let activeResizeGroup: number | null = null;
  let clickedResizeHandle: number | null = null;

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

  $: itemsByRow = groupItemsByRow(items);

  const { canvasStore } = getCanvasStateManagers();

  const dispatch = createEventDispatcher();

  const grid = new Grid(items);

  function handleChange(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
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
    // console.log("[CanvasDashboardPreview] Component hovered:", hoveredIndex);
  }

  function handleMouseLeave(e: CustomEvent<{ index: number }>) {
    if (draggedComponent) {
      // Don't update hover state while dragging
      return;
    }
    if (hoveredIndex === e.detail.index) {
      hoveredIndex = null;
      // console.log("[CanvasDashboardPreview] Component unhovered");
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

  function handleColResize(e: MouseEvent) {
    if (!resizingCol) return;

    const deltaX = e.clientX - resizingCol.startX;
    console.log("[CanvasDashboardPreview] handleColResize", {
      deltaX,
      resizingCol,
    });
    const currentRow = itemsByRow.find((row) =>
      row.items.some((item) => items.indexOf(item) === resizingCol?.index),
    );
    if (!resizingCol || !currentRow) return;

    // Sort items by x position for consistent resizing
    const sortedRowItems = [...currentRow.items].sort(
      (a, b) => (a.x ?? 0) - (b.x ?? 0),
    );
    const resizingItemIndex = sortedRowItems.findIndex(
      (item) => items.indexOf(item) === resizingCol?.index,
    );

    // Get next item to determine maximum resize width
    const nextItem = sortedRowItems[resizingItemIndex + 1];
    if (!nextItem) return;

    const newWidth = Math.round(
      Math.max(defaults.GRID_CELL_SIZE, resizingCol.initialWidth + deltaX),
    );

    const updatedItems = [...items];
    const item = updatedItems[resizingCol.index];
    if (!item) return;

    // Calculate new widths ensuring they stay within bounds
    const currentItemWidth = item.width ?? defaults.COMPONENT_WIDTH;
    const currentX = item.x ?? 0;

    // Get next item's current width
    const nextItemWidth = nextItem.width ?? defaults.COMPONENT_WIDTH;
    const nextItemX = nextItem.x ?? 0;

    // Calculate maximum allowed width to prevent collision
    const maxAllowedWidth = Math.min(
      // Don't exceed grid width
      defaults.COLUMN_COUNT - currentX,
      // Allow resizing considering combined width of current and next item
      nextItemX - currentX + nextItemWidth,
    );

    // Ensure new width doesn't exceed available space
    const finalWidth = Math.min(
      Math.round(newWidth / defaults.GRID_CELL_SIZE),
      maxAllowedWidth,
    );

    const widthDiff = finalWidth - currentItemWidth;

    // Only proceed if we have space to resize
    if (widthDiff === 0) return;

    // Check if resize is possible while maintaining minimum widths
    const canResize =
      finalWidth >= defaults.COMPONENT_MIN_WIDTH && // Minimum 2 columns for current item
      nextItemWidth - widthDiff >= defaults.COMPONENT_MIN_WIDTH; // Minimum 2 columns for next item

    if (canResize) {
      item.width = finalWidth;

      // Only adjust the next item's width
      const nextUpdatedItem = updatedItems[items.indexOf(nextItem)];
      if (nextUpdatedItem) {
        // Maintain next item's minimum width
        nextUpdatedItem.width = nextItemWidth - widthDiff;
        // Update x position of next item to be right after current item
        nextUpdatedItem.x = currentX + finalWidth;
      }

      // Update the UI immediately
      items = updatedItems;
    }
  }

  function handleColResizeEnd() {
    if (resizingCol) {
      const item = items[resizingCol.index];
      if (item) {
        dispatch("update", {
          index: resizingCol.index,
          items,
          position: [item.x, item.y],
          dimensions: [item.width, item.height],
        });
      }
    }
    document.body.classList.remove("resizing-col");
    resizingCol = null;
  }

  function handleColumnResizeStart(
    e: MouseEvent,
    index: number,
    initialWidth: number,
    columnIndex: number,
  ) {
    e.preventDefault();
    resizingCol = {
      index,
      startX: e.clientX,
      initialWidth,
      maxWidth: (defaults.COLUMN_COUNT - columnIndex) * gridCell,
    };
    document.body.classList.add("resizing-col");
  }

  function handleSpreadEvenly(index: number) {
    console.log("[CanvasDashboardPreview] handleSpreadEvenly", {
      index,
    });
    // Get the item at the resize handle
    const selectedItem = items[index];
    if (!selectedItem) return;

    // Get all items in the same row
    const rowItems = items.filter((item) => item.y === selectedItem.y);
    if (!rowItems.length) return;

    // Create a row group for redistribution
    const row = {
      y: selectedItem.y,
      height: selectedItem.height,
      items: rowItems,
    };

    // Get redistributed items
    const redistributedItems = redistributeRowColumns(row);
    if (!redistributedItems) return;

    // Update items with new widths and positions
    items = items.map((item) => {
      const redistributedItem = redistributedItems.find(
        (ri) => ri.x === item.x && ri.y === item.y,
      );
      return redistributedItem || item;
    });

    // Notify parent of update
    dispatch("update", {
      index: -1,
      items,
      position: [0, 0],
      dimensions: [0, 0],
    });
  }

  function handleResizeGroupEnter(index: number) {
    clearTimeout(hideTimeout);
    activeResizeGroup = index;
  }

  function handleResizeGroupLeave(index: number) {
    hideTimeout = setTimeout(() => {
      if (activeResizeGroup === index) {
        activeResizeGroup = null;
      }
    }, 300);
  }

  function handleResizeHandleClick(index: number, e: MouseEvent) {
    e.stopPropagation();
    clickedResizeHandle = clickedResizeHandle === index ? null : index;
  }

  // Add click handler to document to close when clicking outside
  function handleDocumentClick(e: MouseEvent) {
    if (clickedResizeHandle !== null) {
      const target = e.target as HTMLElement;
      if (!target.closest(".floating-buttons")) {
        clickedResizeHandle = null;
      }
    }
  }
</script>

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
  <div class="grid-container w-full">
    {#each itemsByRow as row, index (index)}
      <div
        class="row w-full"
        data-row-index={index}
        style="height: {row.height * gridCell}px;"
      >
        {#each row.items as component, itemIndex}
          {@const i = items.indexOf(component)}
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

          {#if itemIndex < row.items.length - 1}
            <div
              class="col-resize-group"
              role="presentation"
              class:active={activeResizeGroup === i}
              on:mouseenter={() => handleResizeGroupEnter(i)}
              on:mouseleave={() => handleResizeGroupLeave(i)}
            >
              <button
                type="button"
                aria-label="Resize column"
                class="col-resize-handle absolute w-[3px] cursor-col-resize bg-transparent hover:bg-primary-300 z-[50] opacity-0 pointer-events-auto"
                class:bg-primary-300={clickedResizeHandle === i}
                style="left: {((component.x ?? 0) +
                  (component.width ?? defaults.COMPONENT_WIDTH)) *
                  gridCell -
                  1.5}px; height: {row.height * gridCell}px;"
                on:mousedown|stopPropagation={(e) =>
                  handleColumnResizeStart(
                    e,
                    i,
                    (component.width ?? defaults.COMPONENT_WIDTH) * gridCell,
                    component.x ?? 0,
                  )}
                on:click|stopPropagation={(e) => handleResizeHandleClick(i, e)}
              />

              <!-- Floating button group -->
              <div
                class="floating-buttons absolute flex flex-col gap-1 opacity-0 transition-opacity pointer-events-none"
                class:show={clickedResizeHandle === i}
                style="left: {((component.x ?? 0) +
                  (component.width ?? defaults.COMPONENT_WIDTH)) *
                  gridCell -
                  12}px; top: {row.height * gridCell}px;"
              >
                <button
                  type="button"
                  class="p-1 rounded bg-white shadow-md hover:bg-slate-50 border border-slate-200 pointer-events-auto"
                  title="Spread evenly"
                  on:mousedown|stopPropagation
                  on:click|stopPropagation={() => {
                    handleSpreadEvenly(i);
                    clickedResizeHandle = null;
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 16 16"
                    fill="none"
                  >
                    <path
                      fill-rule="evenodd"
                      clip-rule="evenodd"
                      d="M11.8637 1.46968C11.5708 1.17679 11.096 1.17679 10.8031 1.46968C10.5102 1.76257 10.5102 2.23745 10.8031 2.53034L12.1894 3.91668L6.00006 3.91667C5.58584 3.91667 5.25006 4.25246 5.25005 4.66667C5.25005 5.08088 5.58584 5.41667 6.00005 5.41667L12.1894 5.41668L10.8031 6.80301C10.5102 7.09591 10.5102 7.57078 10.8031 7.86368C11.0959 8.15657 11.5708 8.15657 11.8637 7.86368L14.5302 5.19721L14.5304 5.19701C14.671 5.05636 14.7501 4.8656 14.7501 4.66668C14.7501 4.46777 14.671 4.27701 14.5304 4.13635L14.5297 4.13566C14.394 4.00034 14.2068 3.91668 14.0001 3.91668L11.8637 1.46968ZM5.19705 8.13634C5.48994 8.42924 5.48994 8.90411 5.19705 9.197L3.81071 10.5833L10 10.5833C10.4143 10.5833 10.75 10.9191 10.75 11.3333C10.75 11.7476 10.4143 12.0833 10 12.0833L3.81071 12.0833L5.19704 13.4697C5.48993 13.7626 5.48993 14.2374 5.19704 14.5303C4.90415 14.8232 4.42927 14.8232 4.13638 14.5303L1.46972 11.8637C1.39781 11.7918 1.34356 11.7089 1.30696 11.6204C1.27029 11.532 1.25005 11.435 1.25005 11.3333C1.25005 11.2316 1.27029 11.1347 1.30696 11.0462C1.34272 10.9598 1.39535 10.8787 1.46484 10.8079M1.47031 10.8024L4.13639 8.13634C4.42928 7.84345 4.90415 7.84345 5.19705 8.13634"
                      fill="#64748B"
                    />
                  </svg>
                </button>
              </div>
            </div>
          {/if}
        {/each}
      </div>

      <button
        type="button"
        aria-label="Resize row"
        class="row-resize-handle w-full h-[3px] cursor-row-resize bg-transparent hover:bg-primary-300 z-[50] opacity-0 hover:opacity-100 pointer-events-auto"
        on:mousedown|stopPropagation={(e) =>
          handleRowResizeStart(e, index, row.height * gridCell)}
      />
    {/each}
  </div>

  {#if dropTarget && draggedComponent}
    {@const targetItem = items[dropTarget.index]}
    {#if targetItem && targetItem.x !== undefined && targetItem.y !== undefined && targetItem.width !== undefined && targetItem.height !== undefined}
      <DropIndicator
        height={dropTarget.position === "bottom" ||
        dropTarget.position === "top"
          ? 2
          : targetItem.height * gridCell}
        top={dropTarget.position === "bottom"
          ? (targetItem.y + targetItem.height) * gridCell
          : dropTarget.position === "top"
            ? targetItem.y * gridCell
            : targetItem.y * gridCell}
        left={dropTarget.position === "right"
          ? (targetItem.x + targetItem.width) * gridCell
          : dropTarget.position === "bottom" || dropTarget.position === "top"
            ? 0
            : targetItem.x * gridCell}
        width={dropTarget.position === "bottom" || dropTarget.position === "top"
          ? defaults.DASHBOARD_WIDTH
          : undefined}
        orientation={dropTarget.position === "bottom" ||
        dropTarget.position === "top"
          ? "horizontal"
          : "vertical"}
      />
    {/if}
  {/if}
</DashboardWrapper>

<svelte:window
  on:mousemove={(e) => {
    handleRowResize(e);
    handleColResize(e);
  }}
  on:mouseup={() => {
    handleRowResizeEnd();
    handleColResizeEnd();
  }}
  on:click={handleDocumentClick}
/>

<style lang="postcss">
  :global(body.resizing-row) {
    cursor: row-resize !important;
  }

  .grid-container {
    position: relative;
    width: 100%;
  }

  .row {
    position: relative;
    min-height: 0;
  }

  .row-resize-handle {
    position: relative;
    transition: opacity 0.2s;
  }

  :global(body.resizing-col) {
    cursor: col-resize !important;
  }

  .col-resize-group {
    position: relative;
  }

  .col-resize-group.active .col-resize-handle {
    opacity: 1;
  }

  .floating-buttons {
    z-index: 51;
  }

  .floating-buttons.show {
    opacity: 1;
  }
</style>
