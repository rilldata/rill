<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import { get, type Unsubscriber, type Writable } from "svelte/store";
  import CanvasComponent from "./CanvasComponent.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import DropZone from "./components/DropZone.svelte";
  import type { CanvasComponentType } from "./components/types";
  import ElementDivider from "./ElementDivider.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import {
    COLUMN_COUNT,
    MIN_HEIGHT,
    MIN_WIDTH,
    mousePosition,
    normalizeSizeArray,
  } from "./layout-util";
  import RowDropZone from "./RowDropZone.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import type { CanvasEntity } from "./stores/canvas-entity";
  import type { Row } from "./stores/row";
  import { activeDivider } from "./stores/ui-stores";

  export let row: Row;
  export let zIndex: number;
  export let maxWidth: number;
  export let heightUnit: string = "px";
  export let rowIndex: number;
  export let columnWidth: number;
  export let components: CanvasEntity["components"];
  export let dragComponent: BaseCanvasComponent | null;
  export let selectedComponent: Writable<string | null>;
  export let addItems: (
    position: { row: number; column: number },
    items: CanvasComponentType[],
  ) => void;
  export let spreadEvenly: (index: number) => void;
  export let onComponentMouseDown: (params: {
    id: string;
    event: MouseEvent;
  }) => void;
  export let onDuplicate: (params: { columnIndex: number }) => void;
  export let onDelete: (params: { component: BaseCanvasComponent }) => void;
  export let onDrop: (row: number, column: number | null) => void;
  export let initializeRow: (row: number, type: CanvasComponentType) => void;
  export let updateRowHeight: (newHeight: number, index: number) => void;
  export let updateComponentWidths: (
    index: number,
    newWidths: number[],
  ) => void;

  let rowHeight = get(row.height) ?? MIN_HEIGHT;
  let hasLocalChange = false;
  let initialMousePosition: { x: number; y: number };
  let initialHeight: number;
  let unsubscriber: Unsubscriber | undefined = undefined;

  $: ({ height, items: _itemIds, widths: itemWidths } = row);

  $: widths = normalizeSizeArray($itemWidths);

  $: itemIds = $_itemIds;

  $: isSpreadEvenly = widths.every((w) => w === widths[0]);
  $: activelyDragging = !!dragComponent;

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
        const actualHeight = Math.floor(
          document.querySelector(`#${id}`)?.getBoundingClientRect().height ??
            rowHeight,
        );

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
    {@const itemCount = itemIds.length}
    <ItemWrapper type={component?.type} zIndex={4 - columnIndex}>
      {#if columnIndex === 0}
        <ElementDivider
          {rowIndex}
          resizeIndex={-1}
          addIndex={columnIndex}
          rowLength={itemCount}
          dragging={activelyDragging}
          {isSpreadEvenly}
          {spreadEvenly}
          {addItems}
        />
      {/if}

      <ElementDivider
        {isSpreadEvenly}
        columnWidth={widths[columnIndex]}
        {rowIndex}
        dragging={activelyDragging}
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
        allowDrop={activelyDragging &&
          (itemCount < 4 || dragComponent?.pathInYAML?.[1] === rowIndex)}
        {onDrop}
      />

      {#if component}
        <CanvasComponent
          {component}
          editable
          ghost={dragComponent === component}
          selected={$selectedComponent === id}
          allowPointerEvents={!$activeDivider && !activelyDragging}
          onMouseDown={(event) => {
            onComponentMouseDown({
              event,
              id,
            });
          }}
          onDuplicate={() => {
            onDuplicate({ columnIndex });
          }}
          onDelete={() => {
            onDelete({ component });
          }}
        />
      {:else}
        <ComponentError error="No valid component {id} in project" />
      {/if}
    </ItemWrapper>
  {/each}
  <RowDropZone
    allowDrop={activelyDragging}
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
      allowDrop={activelyDragging}
      dropIndex={0}
      {onDrop}
      addItem={(type) => {
        initializeRow(rowIndex, type);
      }}
    />
  {/if}
</RowWrapper>
