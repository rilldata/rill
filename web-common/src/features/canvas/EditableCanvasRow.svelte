<script lang="ts">
  import type { CanvasComponentType } from "./components/types";
  import RowDropZone from "./RowDropZone.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import {
    COLUMN_COUNT,
    MIN_HEIGHT,
    MIN_WIDTH,
    normalizeSizeArray,
  } from "./layout-util";
  import { mousePosition } from "./layout-util";
  import { get, type Unsubscriber } from "svelte/store";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import type { CanvasEntity, LayoutRow } from "./stores/canvas-entity";
  import ComponentError from "./components/ComponentError.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import ElementDivider from "./ElementDivider.svelte";
  import DropZone from "./components/DropZone.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import { activeDivider } from "./stores/ui-stores";

  export let row: LayoutRow;
  export let zIndex = 50;
  export let maxWidth: number;

  export let heightUnit: string = "px";
  export let rowIndex: number;
  export let movingWidget: boolean;

  export let components: CanvasEntity["components"];
  export let dragItemInfo;
  export let addItems;
  export let spreadEvenly;
  export let selectedComponent;

  export let onComponentMouseDown;
  export let onDuplicate;
  export let onDelete;
  export let onDrop: (row: number, column: number | null) => void;
  export let initializeRow: (row: number, type: CanvasComponentType) => void;
  export let updateRowHeight: (newHeight: number, index: number) => void;
  export let updateComponentWidths: (
    index: number,
    newWidths: number[],
  ) => void;
  export let columnWidth: number;

  let rowHeight = get(row.height) ?? MIN_HEIGHT;
  let hasLocalChange = false;
  let initialMousePosition: { x: number; y: number };
  let initialHeight: number;
  let unsubscriber: Unsubscriber | undefined = undefined;

  $: ({ height, itemIds: _itemIds, itemWidths } = row);

  $: widths = normalizeSizeArray($itemWidths);

  $: itemIds = $_itemIds;

  $: isSpreadEvenly = widths.every((w) => w === widths[0]);

  $: updateHeightFromSpec($height);
  $: updateWidthsFromSpec($itemWidths);

  $: id = `canvas-row-${rowIndex}`;

  function onRowResizeStart() {
    initialMousePosition = $mousePosition;
    hasLocalChange = true;

    initialHeight =
      document.querySelector(`#${id}`)?.getBoundingClientRect().height ??
      rowHeight ??
      MIN_HEIGHT;

    unsubscriber = mousePosition.subscribe((position) => {
      const diff = position.y - initialMousePosition.y;

      rowHeight = Math.floor(diff + initialHeight);
    });

    window.addEventListener(
      "mouseup",
      () => {
        unsubscriber?.();
        unsubscriber = undefined;
        const actualHeight =
          document.querySelector(`#${id}`)?.getBoundingClientRect().height ??
          rowHeight;

        rowHeight = actualHeight;
        updateRowHeight(actualHeight, rowIndex);
      },
      { once: true },
    );
  }

  function updateHeightFromSpec(newSpecHeight: number | undefined) {
    if (newSpecHeight === undefined) return;

    const matchesLocal = newSpecHeight === rowHeight;

    if (hasLocalChange && matchesLocal) {
      hasLocalChange = false;
    } else if (!hasLocalChange && !matchesLocal) {
      rowHeight = newSpecHeight;
      hasLocalChange = false;
    }
  }

  function updateWidthsFromSpec(newWidths: number[]) {
    if (newWidths === undefined) return;
    const normalized = normalizeSizeArray(newWidths);
    const matchesLocal = normalized.every((w, i) => w === widths[i]);

    if (hasLocalChange && matchesLocal) {
      hasLocalChange = false;
    } else if (!hasLocalChange && !matchesLocal) {
      widths = normalized;
      hasLocalChange = false;
    }
  }

  function onColumnResizeStart(columnIndex: number) {
    initialMousePosition = $mousePosition;
    hasLocalChange = true;

    const nextElementWidth = widths[columnIndex + 1];
    if (nextElementWidth === undefined) return;

    const maxWidth = widths.reduce((acc, el, i) => {
      if (i === columnIndex) {
        return acc;
      } else if (i === columnIndex + 1) {
        return acc - MIN_WIDTH;
      } else {
        return acc - el;
      }
    }, COLUMN_COUNT);

    const width = widths[columnIndex];

    initialHeight =
      document.querySelector(`#canvas-row-${rowIndex}`)?.getBoundingClientRect()
        .height ??
      rowHeight ??
      MIN_HEIGHT;

    unsubscriber = mousePosition.subscribe((position) => {
      const delta = position.x - (initialMousePosition?.x ?? 0);
      const columnDelta = Math.round(delta / columnWidth);

      const newValue = clamp(MIN_WIDTH, width + columnDelta, maxWidth);

      const clampedDelta = newValue - width;

      widths[columnIndex] = newValue;

      widths[columnIndex + 1] = nextElementWidth - clampedDelta;
    });

    window.addEventListener(
      "mouseup",
      () => {
        unsubscriber?.();
        unsubscriber = undefined;
        updateComponentWidths(rowIndex, widths);
      },
      { once: true },
    );
  }
</script>

<RowWrapper
  {zIndex}
  {maxWidth}
  height={rowHeight}
  {heightUnit}
  {id}
  gridTemplate={widths.map((w) => `${w}fr`).join(" ")}
>
  {#each itemIds as id, columnIndex (columnIndex)}
    {@const component = components.get(id)}
    {@const type = component?.type}
    {@const itemCount = itemIds.length}

    <ItemWrapper type={component?.type} zIndex={4 - columnIndex}>
      {#if columnIndex === 0}
        <ElementDivider
          {rowIndex}
          resizeIndex={-1}
          addIndex={columnIndex}
          rowLength={itemCount}
          dragging={!!dragItemInfo}
          {isSpreadEvenly}
          {spreadEvenly}
          {addItems}
        />
      {/if}

      <ElementDivider
        {isSpreadEvenly}
        columnWidth={widths[columnIndex]}
        {rowIndex}
        dragging={!!dragItemInfo}
        resizeIndex={columnIndex}
        addIndex={columnIndex + 1}
        rowLength={itemCount}
        {spreadEvenly}
        {addItems}
        {onColumnResizeStart}
      />

      <DropZone
        column={columnIndex}
        row={rowIndex}
        maxColumns={itemCount}
        allowDrop={!!dragItemInfo}
        {onDrop}
      />

      {#if component}
        <CanvasComponent
          {component}
          editable
          ghost={dragItemInfo?.position?.row === rowIndex &&
            dragItemInfo?.position?.column === columnIndex}
          selected={$selectedComponent === id}
          allowPointerEvents={!$activeDivider}
          onMouseDown={(event) => {
            onComponentMouseDown({
              event,
              id,
              columnIndex,
              type,
            });
          }}
          onDuplicate={() => {
            onDuplicate({ columnIndex });
          }}
          onDelete={() => {
            onDelete({ columnIndex });
          }}
        />
      {:else}
        <ComponentError error="No valid component {id} in project" />
      {/if}
    </ItemWrapper>
  {/each}
  <RowDropZone
    allowDrop={movingWidget}
    resizeIndex={rowIndex}
    dropIndex={rowIndex + 1}
    {onRowResizeStart}
    {onDrop}
    addItem={(type) => {
      initializeRow(rowIndex + 1, type);
    }}
  />

  {#if rowIndex === 0}
    <RowDropZone
      allowDrop={movingWidget}
      dropIndex={0}
      {onDrop}
      addItem={(type) => {
        initializeRow(rowIndex, type);
      }}
    />
  {/if}
</RowWrapper>
