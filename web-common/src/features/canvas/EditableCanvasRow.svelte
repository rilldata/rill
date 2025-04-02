<script lang="ts">
  import type {
    V1CanvasRow as APIV1CanvasRow,
    V1CanvasItem,
  } from "@rilldata/web-common/runtime-client";
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
  import type { Unsubscriber } from "svelte/store";
  import { clamp } from "@rilldata/web-common/lib/clamp";

  type V1CanvasRow = Omit<APIV1CanvasRow, "items"> & {
    items: (V1CanvasItem | null)[];
  };

  export let row: V1CanvasRow;
  export let zIndex = 50;
  export let maxWidth: number;
  export let heightUnit: string = "px";
  export let rowIndex: number;
  export let movingWidget: boolean;
  export let onDrop: (row: number, column: number | null) => void;
  export let initializeRow: (row: number, type: CanvasComponentType) => void;
  export let updateRowHeight: (newHeight: number, index: number) => void;
  export let updateComponentWidths: (
    index: number,
    newWidths: number[],
  ) => void;
  export let columnWidth: number;

  let rowHeight = row.height ?? MIN_HEIGHT;
  let componentWidths: number[] = normalizeSizeArray(
    row.items?.map((item) => item?.width ?? 0),
  );
  let hasLocalChange = false;
  let initialMousePosition: { x: number; y: number };
  let initialHeight: number;
  let unsubscriber: Unsubscriber | undefined = undefined;

  $: items = row.items ?? [];

  $: isSpreadEvenly = componentWidths.every((w) => w === componentWidths[0]);

  $: heightUnit = row.heightUnit ?? "px";

  $: updateHeightFromSpec(row.height);
  $: updateWidthsFromSpec(row.items?.map((item) => item?.width ?? 0));

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
    const matchesLocal = normalized.every((w, i) => w === componentWidths[i]);

    if (hasLocalChange && matchesLocal) {
      hasLocalChange = false;
    } else if (!hasLocalChange && !matchesLocal) {
      componentWidths = normalized;
      hasLocalChange = false;
    }
  }

  function onColumnResizeStart(columnIndex: number) {
    initialMousePosition = $mousePosition;
    hasLocalChange = true;

    const nextElementWidth = componentWidths[columnIndex + 1];
    if (nextElementWidth === undefined) return;

    const maxWidth = componentWidths.reduce((acc, el, i) => {
      if (i === columnIndex) {
        return acc;
      } else if (i === columnIndex + 1) {
        return acc - MIN_WIDTH;
      } else {
        return acc - el;
      }
    }, COLUMN_COUNT);

    const width = componentWidths[columnIndex];

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

      componentWidths[columnIndex] = newValue;

      componentWidths[columnIndex + 1] = nextElementWidth - clampedDelta;
    });

    window.addEventListener(
      "mouseup",
      () => {
        unsubscriber?.();
        unsubscriber = undefined;
        updateComponentWidths(rowIndex, componentWidths);
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
  gridTemplate={componentWidths.map((w) => `${w}fr`).join(" ")}
>
  <slot
    widths={componentWidths}
    {isSpreadEvenly}
    {items}
    {onColumnResizeStart}
  />

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
