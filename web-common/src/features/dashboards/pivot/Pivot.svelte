<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";
  import { createEventDispatcher } from "svelte";
  import type { PivotPos } from "./types";
  import { isEmptyPos } from "./util";

  // TODO: replace with w-full once you have fixed widths
  const LOADING_EL = `<div class="w-8 h-4 bg-gray-100 rounded loading-cell"/>`;
  const LOADING_EL_FW = `<div class="min-w-8 w-8 h-4 bg-gray-100 rounded loading-cell"/>`;

  export let getColumnHeaderData: (pos: PivotPos) => any = () => [];
  export let getRowHeaderData: (pos: PivotPos) => any = () => [];
  export let getBodyData: (pos: PivotPos) => any = () => [];
  export let rowCount = 0;
  export let columnCount = 0;
  export let rowHeaderDepth = 0;
  export let columnHeaderDepth = 0;
  export let onMouseDown: (evt: MouseEvent, table: any) => any = undefined;
  export let renderCell: ({ x, y, value, element }) => string | void = () => {};
  export let renderColumnHeader: ({
    x,
    y,
    value,
    element,
  }) => string | void = () => {};
  export let renderRowHeader: ({
    x,
    y,
    value,
    element,
  }) => string | void = () => {};
  export let renderRowCorner: ({
    x,
    y,
    value,
    element,
  }) => string | void = () => {};
  export let rowHeight = 24;
  export let getColumnWidth: (x: number) => number | void = () => {};
  export let getRowHeaderWidth: (x: number) => number | void = () => {};

  const dispatch = createEventDispatcher();

  let table = undefined;

  export let api = {
    initialized: false,
    draw() {
      if (this.initialized) {
        this.getTable().draw();
      }
    },
    getTable() {
      return table;
    },
  };

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
    const placeholderEvenColumnHeader = new Array(columnHeaderDepth).fill(
      "LOADING"
    );
    const placeholderOddColumnHeader = new Array(columnHeaderDepth).fill(
      "\u200BLOADING"
    );
    let column_headers = getColumnHeaderData({ x0, x1, y0, y1 });
    // Replace any nulls with loading placeholders
    for (let i = 0; i < column_headers.length; i++) {
      if (!column_headers[i]) {
        column_headers[i] =
          i % 2 ? placeholderEvenColumnHeader : placeholderOddColumnHeader;
      }
    }

    const placeholderEvenRowHeader = new Array(rowHeaderDepth)
      .fill("LOADING")
      .map((_, i) => `${i % 2 ? "\u200B" : ""}LOADING`);
    const placeholderOddRowHeader = [
      ...placeholderEvenRowHeader.slice(1),
      placeholderEvenRowHeader.at(0),
    ];
    let row_headers = getRowHeaderData({ x0, x1, y0, y1 });

    row_headers.forEach((r, i) => {
      if (!r) {
        row_headers[i] =
          i % 2 ? placeholderEvenRowHeader : placeholderOddRowHeader;
      }
    });

    let data = getBodyData({ x0, x1, y0, y1 });
    // Replace nulls with loading placeholders
    data.forEach((c, i) => {
      c.forEach((r, j) => {
        if (!r) {
          data[i][j] = "LOADING";
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

    return dataSlice;
  };

  function style_row_th(th: HTMLElement) {
    const meta = table.getMeta(th);
    const x = meta.row_header_x;
    const y = meta.y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));

    const maybeWidth = getRowHeaderWidth(x);
    if (maybeWidth) {
      th.style.width = `${maybeWidth}px`;
      th.style.minWidth = `${maybeWidth}px`;
      th.style.maxWidth = `${maybeWidth}px`;
    }

    if (meta.value === "LOADING" || meta.value === "\u200BLOADING")
      th.innerHTML = LOADING_EL_FW;

    const maybeVal = renderRowHeader({
      x,
      y,
      value: meta.value,
      element: th,
    });
    if (maybeVal) th.innerHTML = maybeVal;
  }

  function style_td(td: HTMLElement) {
    const { x, y, value } = table.getMeta(td);
    td.setAttribute("__col", String(x));
    td.setAttribute("__row", String(y));

    const maybeWidth = getColumnWidth(x);
    if (maybeWidth) {
      td.style.width = `${maybeWidth}px`;
      td.style.minWidth = `${maybeWidth}px`;
      td.style.maxWidth = `${maybeWidth}px`;
    }

    if (td.textContent === "LOADING" || td.textContent === "\u200BLOADING")
      td.innerHTML = LOADING_EL_FW;
    const maybeVal = renderCell({ x, y, value, element: td });
    if (maybeVal) td.innerHTML = maybeVal;
  }

  function style_column_th(th: HTMLElement) {
    const meta = table.getMeta(th);
    const x = meta.x;
    const y = meta.column_header_y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));

    if (y === columnHeaderDepth - 1) {
      const maybeWidth = getColumnWidth(x);
      if (maybeWidth) {
        th.style.width = `${maybeWidth}px`;
        th.style.minWidth = `${maybeWidth}px`;
        th.style.maxWidth = `${maybeWidth}px`;
      }
    }

    if (th.textContent === "LOADING" || th.textContent === "\u200BLOADING")
      th.innerHTML = LOADING_EL_FW;
    const maybeVal = renderColumnHeader({
      x,
      y,
      value: meta.value,
      element: th,
    });
    if (maybeVal) th.innerHTML = maybeVal;
  }

  function style_row_corner(th: HTMLElement) {
    const meta = table.getMeta(th);

    if (meta.column_header_y === columnHeaderDepth - 1) {
      const maybeWidth = getRowHeaderWidth(meta.row_header_x);
      if (maybeWidth) {
        th.style.width = `${maybeWidth}px`;
        th.style.minWidth = `${maybeWidth}px`;
        th.style.maxWidth = `${maybeWidth}px`;
      }
    }

    const maybeVal = renderRowCorner({
      x: meta.row_header_x,
      y: meta.column_header_y,
      value: meta.value,
      element: th,
    });
    if (maybeVal) th.innerHTML = maybeVal;
  }

  $: {
    if (table) {
      table.setDataListener(reactiveDataListener);
      table.draw();
      api.initialized = true;
    }
  }

  const handlerCache = new Map();
  function addHandler(type: string, handler: (evt: MouseEvent) => any) {
    table.addEventListener(type, handler);
    const prevHandler = handlerCache.get(type);
    if (prevHandler) {
      table.removeEventListener("mousedown", prevHandler);
    }
    handlerCache.set(type, handler);
  }

  $: {
    if (table && onMouseDown) {
      const handler = (evt: MouseEvent) => onMouseDown(evt, table);
      addHandler("mousedown", handler);
    }
  }

  let lastColumnSizer = null;
  let lastRowHeaderSizer = null;
  function styleListener() {
    for (const td of table.querySelectorAll("tbody td")) {
      style_td(td);
    }

    for (const th of table.querySelectorAll("tbody th")) {
      style_row_th(th);
    }

    for (const th of table.querySelectorAll("thead th:not(.rt-group-corner)")) {
      style_column_th(th);
    }

    for (const th of table.querySelectorAll("thead th.rt-group-corner")) {
      style_row_corner(th);
    }
    // If the column sizer or row header sizer function has changed since last style call, invalidate the table column width caches so horizontal scrolling is properly calculated
    if (
      lastColumnSizer !== getColumnWidth ||
      lastRowHeaderSizer !== getRowHeaderWidth
    ) {
      console.log("invalidate");
      table.invalidate();
      lastColumnSizer = getColumnWidth;
      lastRowHeaderSizer = getRowHeaderWidth;
    }

    dispatch("pos", pos);
  }
  $: {
    if (table) {
      table.addStyleListener(styleListener);
    }
  }

  $: cssVarStyles = `--row-height: ${rowHeight}px;`;
</script>

<div
  class="border relative"
  style={`height: 400px; width: 100%; ${cssVarStyles}`}
>
  <regular-table class="w-full h-full tdd-table" bind:this={table} />
</div>

<style>
  :global(regular-table table) {
    table-layout: fixed;
    border-collapse: collapse;
    font-family: Inter;
    font-feature-settings: "case" 0, "cpsp" 0, "dlig" 0, "frac" 0, "dnom" 0,
      "numr" 0, "salt" 0, "subs" 0, "sups" 0, "tnum", "zero" 0, "ss01", "ss02" 0,
      "ss03" 0, "ss04" 0, "cv01" 0, "cv02" 0, "cv03" 0, "cv04" 0, "cv05" 0,
      "cv06" 0, "cv07" 0, "cv08" 0, "cv09" 0, "cv10" 0, "cv11" 0, "calt", "ccmp",
      "kern";
  }

  :global(regular-table *) {
    box-sizing: border-box;
  }

  :global(regular-table tr td) {
    height: var(--row-height);
    /* width: 60px;
    min-width: 60px;
    max-width: 60px;
    padding-block: 0px;
    padding: 0px;
    text-align: right;
    background-color: white; */
  }

  :global(.tdd-table tbody tr td) {
    /* background-color: initial; */
    /* background-color: white;
    color: black; */
    /* padding-inline: 8px; */
  }

  /* Figure out how to get rid of that stupid hover color, or at least customize it */
  :global(.tdd-table tbody tr:hover td) {
    /* background-color: transparent; */
  }

  :global(regular-table thead th) {
    height: var(--row-height);
  }

  /* Example of fixing column width in the header */
  :global(
      regular-table
        thead
        tr:last-child
        th:not([__col="0"]):not(.rt-group-corner)
    ) {
    /* min-width: 70px !important;
    width: 70px;
    max-width: 70px; */
  }

  :global(regular-table thead tr:last-child th[__col="0"]) {
    /* min-width: 92px !important;
    width: 92px;
    max-width: 92px; */
  }

  :global(regular-table tbody th[__col="0"]) {
    /* color: red; */
    /* border-bottom: 1px solid black; */
  }
</style>
