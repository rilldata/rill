<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";
  import { basicPivot } from "./configs";
  import { getColumnHeaders, getMetadata, getRowHeaders } from "./api";
  import { createEventDispatcher } from "svelte";
  import type { PivotDataProvider } from "./pivot-data-provider";
  import type { PivotPos } from "./types";
  import { isEmptyPos, range } from "./util";

  // TODO: replace with w-full once you have fixed widths
  const LOADING_EL = `<div class="w-8 h-4 bg-gray-100 rounded loading-cell"/>`;
  const LOADING_EL_FW = `<div class="min-w-8 w-8 h-4 bg-gray-100 rounded loading-cell"/>`;

  export let rowHeaderDataProvider: PivotDataProvider;
  export let columnHeaderDataProvider: PivotDataProvider;
  export let bodyDataProvider: PivotDataProvider;

  const getRowHeaderData = (pos: PivotPos) =>
    rowHeaderDataProvider.getData(pos);
  $: rowHeaderQuery = rowHeaderDataProvider.query;

  const getColumnHeaderData = (pos: PivotPos) =>
    columnHeaderDataProvider.getData(pos);
  $: columnHeaderQuery = columnHeaderDataProvider.query;

  const getBodyData = (pos: PivotPos) => bodyDataProvider.getData(pos);
  $: bodyQuery = bodyDataProvider.query;

  const dispatch = createEventDispatcher();

  const config = basicPivot;
  $: metadata = getMetadata(config);

  let table;
  const cachedState = {
    pos: {
      x0: 0,
      x1: 0,
      y0: 0,
      y1: 0,
    },
  };

  // Getters to avoid triggering reactivity when reading
  const getCachedPos = () => cachedState.pos;

  $: reactiveDataListener = (x0, y0, x1, y1) => {
    if (!isEmptyPos({ x0, x1, y0, y1 })) {
      cachedState.pos = { x0, x1, y0, y1 };
    }
    const depth = 2;
    const placeholderEvenColumnHeader = new Array(depth).fill("LOADING");
    const placeholderOddColumnHeader = new Array(depth).fill("\u200BLOADING");
    let column_headers = [];
    const lastColumnBlock = getColumnHeaderData({ x0, x1, y0, y1 });
    if (lastColumnBlock) {
      const xOffset = x0 - lastColumnBlock.block[0];
      column_headers = lastColumnBlock.data.slice(xOffset);
    }
    while (column_headers.length < x1 - x0) {
      column_headers.push(
        column_headers.length % 2
          ? placeholderEvenColumnHeader
          : placeholderOddColumnHeader
      );
    }

    const rowDepth = 2;
    const placeholderEvenRowHeader = new Array(rowDepth)
      .fill("LOADING")
      .map((_, i) => `${i % 2 ? "\u200B" : ""}LOADING`);
    const placeholderOddRowHeader = [
      ...placeholderEvenRowHeader.slice(1),
      placeholderEvenRowHeader.at(0),
    ];

    let row_headers = [];
    const rowHeaderData = getRowHeaderData({ x0, x1, y0, y1 });
    if (rowHeaderData && !isEmptyPos({ x0, x1, y0, y1 })) {
      const yOffset = y0 - rowHeaderData.block[0];
      const yEnd = y1 - rowHeaderData.block[0];
      row_headers = rowHeaderData.data.slice(yOffset, yEnd);
    }
    row_headers.forEach((r, i) => {
      if (!r) {
        row_headers[i] =
          i % 2 ? placeholderEvenRowHeader : placeholderOddRowHeader;
      }
    });
    while (row_headers.length < y1 - y0) {
      row_headers.push(
        row_headers.length % 2
          ? placeholderEvenRowHeader
          : placeholderOddRowHeader
      );
    }

    let data = [];
    const placeholderData = new Array(x1 - x0).fill("LOADING");
    const bodyData = getBodyData({ x0, x1, y0, y1 });
    if (bodyData && !isEmptyPos({ x0, x1, y0, y1 })) {
      const yOffset = y0 - bodyData.block.y[0];
      const yEnd = y1 - bodyData.block.y[0];
      const xOffset = x0 - bodyData.block.x[0];
      const xEnd = x1 - bodyData.block.x[0];
      data = bodyData.data
        .slice(yOffset, yEnd)
        .map((r) => r.slice(xOffset, xEnd));
    }
    data.forEach((r, i) => {
      if (!r) {
        data[i] = placeholderData;
      }
    });
    while (data.length < y1 - y0) {
      data.push(placeholderData);
    }

    const dataSlice = {
      num_rows: metadata.rowCt,
      num_columns: metadata.colCt,
      data, //: range(x0, x1, (x) => range(y0, y1, (y) => `${y},${x}`)),
      row_headers,
      column_headers,
    };

    return dataSlice;
  };

  function map_row_th(th: Element, data: any) {
    const meta = table.getMeta(th);
    const x = meta.row_header_x;
    const y = meta.y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));
    if (th.textContent === "LOADING" || th.textContent === "\u200BLOADING")
      th.innerHTML = LOADING_EL_FW;
  }

  function map_td(td: Element) {
    const meta = table.getMeta(td);
    const x = meta.x;
    const y = meta.y;
    td.setAttribute("__col", String(x));
    td.setAttribute("__row", String(y));
    if (td.textContent === "LOADING" || td.textContent === "\u200BLOADING")
      td.innerHTML = LOADING_EL_FW;
  }

  function map_column_th(th: Element) {
    const meta = table.getMeta(th);
    const x = meta.x;
    const y = meta.column_header_y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));
    if (th.textContent === "LOADING" || th.textContent === "\u200BLOADING")
      th.innerHTML = LOADING_EL_FW;
  }

  function renderRowHeader(cell: Element, data: any) {
    const x = Number(cell.getAttribute("__col"));
    const y = Number(cell.getAttribute("__row"));
    if (data) {
      const yOffset = y - data.block[0];
      const v = data.data[yOffset][x];
      cell.innerHTML = v;
    } else cell.innerHTML = LOADING_EL;
  }

  $: {
    if (table) {
      table.setDataListener(reactiveDataListener);
      table.draw();
    }
  }

  $: {
    if (table) {
      table.addStyleListener(() => {
        const data = getRowHeaderData(getCachedPos());
        for (const td of table.querySelectorAll("tbody td")) {
          map_td(td);
        }

        for (const th of table.querySelectorAll("tbody th")) {
          map_row_th(th, data);
        }

        for (const th of table.querySelectorAll("thead th")) {
          map_column_th(th);
        }
        dispatch("pos", getCachedPos());
      });
    }
  }

  let lastSeenRowQuery = null;
  const getLastSeenRowQuery = () => lastSeenRowQuery;
  $: {
    // If data and we haven't drawn this data already, redraw the table
    if (
      $rowHeaderQuery.data &&
      $rowHeaderQuery.data !== getLastSeenRowQuery()
    ) {
      lastSeenRowQuery = $rowHeaderQuery.data;
      table.draw();
    }
  }

  let lastSeenColumnQuery = null;
  const getLastSeenColumnQuery = () => lastSeenColumnQuery;
  $: {
    if (
      $columnHeaderQuery.data &&
      getLastSeenColumnQuery() !== $columnHeaderQuery.data
    ) {
      lastSeenColumnQuery = $columnHeaderQuery.data;
      table.draw();

      if ($columnHeaderQuery.data.block[0] >= 175) {
        metadata.colCt = 500;
      }
    }
  }

  let lastSeenBodyData = null;
  const getLastSeenBodyData = () => lastSeenBodyData;
  $: {
    if ($bodyQuery.data && getLastSeenBodyData() !== $bodyQuery.data) {
      lastSeenBodyData = $bodyQuery.data;
      table.draw();
    }
  }

  $: {
    console.log({ metadata });
  }
</script>

<div class="max-h-64 overflow-auto w-fit">
  <pre>
  {JSON.stringify(basicPivot, null, 2)}
</pre>
</div>
<div class="border m-8 relative" style="height: 400px; width: 100%">
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
    height: 24px;
    width: 60px;
    min-width: 60px;
    max-width: 60px;
    padding-block: 0px;
    text-align: right;
    background-color: white;
  }

  :global(regular-table thead th) {
    height: 24px;
  }

  /* Example of fixing column width in the header */
  :global(regular-table thead tr:last-child th:not([__col="0"])) {
    min-width: 70px !important;
    width: 70px;
    max-width: 70px;
  }

  :global(regular-table thead tr:last-child th[__col="0"]) {
    min-width: 92px !important;
    width: 92px;
    max-width: 92px;
  }
</style>
