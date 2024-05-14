<script lang="ts">
  import "regular-table";
  import type { RegularTableElement } from "regular-table";
  import { createEventDispatcher, onMount } from "svelte";
  import "./regular-table-style.css";
  import type { PivotPos, PivotRenderCallback } from "./types";
  import { isEmptyPos, range } from "./util";

  const LOADING_CELL = `<div load class="loading-cell h-4 bg-gray-50 rounded" style="width: 100%; min-width: 32px;"/>`;
  const NULL_CELL = `<div style="font-size:0.925em" class="null-cell opacity-50">-</div>`;

  export let getColumnHeaderData: (pos: PivotPos) => any = () => [];
  export let getRowHeaderData: (pos: PivotPos) => any = () => [];
  export let getBodyData: (pos: PivotPos) => any = () => [];
  export let rowCount = 0;
  export let columnCount = 0;
  export let rowHeaderDepth = 0;
  export let columnHeaderDepth = 0;
  export let onMouseDown: undefined | ((evt: MouseEvent, table: any) => any) =
    undefined;
  export let onMouseHover: undefined | ((evt: MouseEvent, table: any) => any) =
    undefined;
  export let renderCell: PivotRenderCallback = () => undefined;
  export let renderColumnHeader: PivotRenderCallback = () => undefined;
  export let renderRowHeader: PivotRenderCallback = () => undefined;
  export let renderRowCorner: PivotRenderCallback = () => undefined;
  export let rowHeight = 24;
  export let getColumnWidth: (x: number) => number | void = () => undefined;
  export let getRowHeaderWidth: (x: number) => number | void = () => undefined;

  const dispatch = createEventDispatcher();

  let table: RegularTableElement;
  let initialized = false;
  export function draw() {
    if (initialized && table) table.draw();
  }
  export const isInitialized = () => initialized;

  export function scrollToCell(x: number, y: number) {
    if (initialized) {
      table?.scrollToCell(x, y, columnCount, rowCount);
      table?.invalidate();
    }
  }

  let pos = {
    x0: 0,
    x1: 0,
    y0: 0,
    y1: 0,
  };

  $: reactiveDataListener = (x0, y0, x1, y1) => {
    if (!isEmptyPos({ x0, x1, y0, y1 })) {
      pos = { x0, x1, y0, y1 };
    }

    let column_headers = getColumnHeaderData({ x0, x1, y0, y1 });
    // Replace any nulls with loading placeholders
    for (let i = 0; i < column_headers.length; i++) {
      if (!column_headers[i]) {
        column_headers[i] = range(0, columnHeaderDepth, () =>
          structuredClone(LOADING_CELL),
        );
      }
    }

    let row_headers = getRowHeaderData({ x0, x1, y0, y1 });
    // Replace any null rows with loading placeholders
    row_headers.forEach((r, i) => {
      if (!r) {
        row_headers[i] = range(0, rowHeaderDepth, () =>
          structuredClone(LOADING_CELL),
        );
      }
      // Replace null values inside the header with null placeholders
      // Assumes that the second item in the array is always the measure
      // value
      else if (r?.[1]?.value === null) {
        row_headers[i][1].value = NULL_CELL;
      }
    });

    let data = getBodyData({ x0, x1, y0, y1 });
    // Replace undefined with loading placeholders and nulls with null placeholder
    data.forEach((c, i) => {
      c.forEach((r, j) => {
        if (r === undefined) {
          data[i][j] = LOADING_CELL;
        } else if (r === null) {
          data[i][j] = NULL_CELL;
        }
      });
    });

    const dataSlice = {
      num_rows: rowCount,
      num_columns: columnCount,
      data,
      row_headers,
      column_headers,
    };

    return Promise.resolve(dataSlice);
  };

  function style_row_th(th: HTMLTableCellElement) {
    const meta = table.getMeta(th);
    const numFixedCols = meta?.row_header?.length;
    const x = meta?.row_header_x;
    const y = meta?.y;

    if (typeof x !== "number" || typeof y !== "number") return;
    th.setAttribute("__col", String(x - numFixedCols!));
    th.setAttribute("__row", String(y));

    const maybeWidth = getRowHeaderWidth(x);
    if (maybeWidth) {
      th.style.width = `${maybeWidth}px`;
      th.style.minWidth = `${maybeWidth}px`;
      th.style.maxWidth = `${maybeWidth}px`;
    }

    if (
      typeof meta.value === "string" &&
      meta.value?.includes("loading-cell")
    ) {
      th.innerHTML = meta.value;
    }

    const maybeVal = renderRowHeader({
      x,
      y,
      value: meta.value,
      element: th,
    });
    if (maybeVal) th.innerHTML = maybeVal;
  }

  function style_td(td: HTMLTableCellElement) {
    const meta = table.getMeta(td);
    const x = meta?.x;
    const y = meta?.y;
    if (typeof x !== "number" || typeof y !== "number") return;

    const value = meta?.value;
    td.setAttribute("__col", String(x));
    td.setAttribute("__row", String(y));

    const maybeWidth = getColumnWidth(x);
    if (maybeWidth) {
      td.style.width = `${maybeWidth}px`;
      td.style.minWidth = `${maybeWidth}px`;
      td.style.maxWidth = `${maybeWidth}px`;
    }

    if (
      typeof value === "string" &&
      (value?.includes("loading-cell") || value?.includes("null-cell"))
    ) {
      td.innerHTML = value;
    }

    const maybeVal = renderCell({ x, y, value, element: td });
    if (maybeVal) td.innerHTML = maybeVal;
  }

  function style_column_th(th: HTMLTableCellElement) {
    const meta = table.getMeta(th);
    const x = meta.x;
    const y = meta.column_header_y;
    if (typeof x !== "number" || typeof y !== "number") return;

    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));

    if (y === columnHeaderDepth - 1) {
      const maybeWidth = getColumnWidth(x);
      if (maybeWidth) {
        th.style.width = `${maybeWidth}px`;
      }
    }

    if (
      typeof meta.value === "string" &&
      meta.value?.includes("loading-cell")
    ) {
      th.innerHTML = meta.value;
    }

    const maybeVal = renderColumnHeader({
      x,
      y,
      value: meta.value,
      element: th,
    });
    if (maybeVal)
      th.innerHTML = maybeVal + `<span class="rt-column-resize"></span>`;
  }

  function style_row_corner(th: HTMLTableCellElement) {
    const meta = table.getMeta(th);

    const x = meta.row_header_x;
    if (typeof x !== "number") return;

    if (meta.column_header_y === columnHeaderDepth - 1) {
      const maybeWidth = getRowHeaderWidth(x);
      if (maybeWidth) {
        th.style.width = `${maybeWidth}px`;
      }
    }

    const maybeVal = renderRowCorner({
      x: meta.row_header_x!,
      y: meta.column_header_y!,
      value: meta.value,
      element: th,
    });

    if (maybeVal)
      th.innerHTML = maybeVal + `<span class="rt-column-resize"></span>`;
  }

  $: {
    if (table) {
      table.setDataListener(reactiveDataListener);
      table.draw();
      initialized = true;
    }
  }

  const handlerCache = new Map();
  function addHandler(type: string, handler: (evt: MouseEvent) => any) {
    table?.addEventListener(type, handler);
    const prevHandler = handlerCache.get(type);
    if (prevHandler) {
      table?.removeEventListener(type, prevHandler);
    }
    handlerCache.set(type, handler);
  }

  $: {
    if (table && onMouseDown) {
      const handler = (evt: MouseEvent) =>
        onMouseDown ? onMouseDown(evt, table) : undefined;
      const hoverHandler = (evt: MouseEvent) =>
        onMouseHover ? onMouseHover(evt, table) : undefined;
      addHandler("mouseover", hoverHandler);
      addHandler("mouseout", hoverHandler);
      addHandler("mousedown", handler);
    }
  }

  let lastColumnSizer: null | ((x: number) => number | void) = null;
  let lastRowHeaderSizer: null | ((x: number) => number | void) = null;
  function styleListener() {
    for (const td of Array.from(table?.querySelectorAll("tbody td") || [])) {
      style_td(td as HTMLTableCellElement);
    }

    for (const th of Array.from(table?.querySelectorAll("tbody th") || [])) {
      style_row_th(th as HTMLTableCellElement);
    }

    for (const th of Array.from(
      table?.querySelectorAll("thead th:not(.rt-group-corner)") || [],
    )) {
      style_column_th(th as HTMLTableCellElement);
    }

    for (const th of Array.from(
      table?.querySelectorAll("thead th.rt-group-corner") || [],
    )) {
      style_row_corner(th as HTMLTableCellElement);
    }
    /**
     * If the column sizer or row header sizer function has
     * changed since last style call, invalidate the table column
     * width caches so horizontal scrolling is  properly calculated
     * */
    if (
      lastColumnSizer !== getColumnWidth ||
      lastRowHeaderSizer !== getRowHeaderWidth
    ) {
      table?.invalidate();
      lastColumnSizer = getColumnWidth;
      lastRowHeaderSizer = getRowHeaderWidth;
    }

    dispatch("pos", pos);
  }

  onMount(() => {
    table?.addStyleListener(styleListener);
  });

  $: cssVarStyles = `--row-height: ${rowHeight}px;`;
</script>

<div class="relative w-full h-full" style={cssVarStyles}>
  <regular-table class="w-full h-full tdd-table" bind:this={table} />
</div>

<style>
  :global(regular-table) {
    padding-left: 0px;
    padding-right: 0px;
  }
  :global(regular-table table) {
    color: rgb(55 65 81);
    table-layout: fixed;
    border-collapse: separate;
    font-family: Inter;
    font-feature-settings:
      "case" 0,
      "cpsp" 0,
      "dlig" 0,
      "frac" 0,
      "dnom" 0,
      "numr" 0,
      "salt" 0,
      "subs" 0,
      "sups" 0,
      "tnum",
      "zero" 0,
      "ss01",
      "ss02" 0,
      "ss03" 0,
      "ss04" 0,
      "cv01" 0,
      "cv02" 0,
      "cv03" 0,
      "cv04" 0,
      "cv05" 0,
      "cv06" 0,
      "cv07" 0,
      "cv08" 0,
      "cv09" 0,
      "cv10" 0,
      "cv11" 0,
      "calt",
      "ccmp",
      "kern";
  }

  :global(regular-table thead tr) {
    border-bottom: 2px solid rgb(229 231 235);
  }

  :global(regular-table *) {
    box-sizing: border-box;
  }

  :global(regular-table tr td) {
    height: var(--row-height);
  }

  :global(regular-table thead th) {
    height: var(--row-height);
    user-select: none;
  }
</style>
