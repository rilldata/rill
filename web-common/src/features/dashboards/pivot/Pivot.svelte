<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";
  import { basicPivot } from "./configs";
  import { getColumnHeaders, getMetadata, getRowHeaders } from "./api";
  import { createEventDispatcher } from "svelte";
  import type { PivotDataProvider } from "./pivot-data-provider";
  import type { PivotPos } from "./types";
  import { isEmptyPos, range } from "./util";
  import type { ColumnDataProvider } from "./column-data-provider";
  import type { RowDataProvider } from "./row-data-provider";
  import type { BodyDataProvider } from "./body-data-provider";
  import Row from "@rilldata/web-common/components/virtualized-table/core/Row.svelte";

  // TODO: replace with w-full once you have fixed widths
  const LOADING_EL = `<div class="w-8 h-4 bg-gray-100 rounded loading-cell"/>`;
  const LOADING_EL_FW = `<div class="min-w-8 w-8 h-4 bg-gray-100 rounded loading-cell"/>`;

  export let rowHeaderDataProvider: RowDataProvider;
  export let columnHeaderDataProvider: ColumnDataProvider;
  export let bodyDataProvider: BodyDataProvider;

  // const getRowHeaderData = (pos: PivotPos) =>
  //   rowHeaderDataProvider.getData(pos);
  // $: rowHeaderQuery = rowHeaderDataProvider.query;

  // const getBodyData = (pos: PivotPos) => bodyDataProvider.getData(pos);
  // $: bodyQuery = bodyDataProvider.query;

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
    let column_headers =
      columnHeaderDataProvider.getData({
        x0,
        x1,
        y0,
        y1,
      })?.data ?? [];
    // Replace any nulls with loading placeholders
    for (let i = 0; i < column_headers.length; i++) {
      if (!column_headers[i]) {
        column_headers[i] =
          i % 2 ? placeholderEvenColumnHeader : placeholderOddColumnHeader;
      }
    }

    const rowDepth = 2;
    const placeholderEvenRowHeader = new Array(rowDepth)
      .fill("LOADING")
      .map((_, i) => `${i % 2 ? "\u200B" : ""}LOADING`);
    const placeholderOddRowHeader = [
      ...placeholderEvenRowHeader.slice(1),
      placeholderEvenRowHeader.at(0),
    ];
    let row_headers =
      rowHeaderDataProvider.getData({ x0, x1, y0, y1 })?.data ?? [];
    row_headers.forEach((r, i) => {
      if (!r) {
        row_headers[i] =
          i % 2 ? placeholderEvenRowHeader : placeholderOddRowHeader;
      }
    });

    let data = bodyDataProvider.getData({ x0, x1, y0, y1 }).data;
    const placeholderData = new Array(x1 - x0).fill("LOADING");
    // Replace nulls with loading placeholders
    data.forEach((c, i) => {
      c.forEach((r, j) => {
        if (!r) {
          data[i][j] = "LOADING";
        }
      });
    });

    const dataSlice = {
      num_rows: metadata.rowCt,
      num_columns: metadata.colCt,
      data,
      // placeholder body data, overwrite during styling step
      // data: range(x0, x1, (x) => range(y0, y1, (y) => `${y},${x}`)),
      row_headers,
      column_headers,
    };

    return dataSlice;
  };

  function map_row_th(th: Element) {
    const meta = table.getMeta(th);
    const x = meta.row_header_x;
    const y = meta.y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));
    if (meta.value === "LOADING" || meta.value === "\u200BLOADING")
      th.innerHTML = LOADING_EL_FW;
    else if (x === 0) {
      th.innerHTML = `<div class="xborder-t xborder-b" style=" height: 100%; padding-top: 3px;">${meta.value}</div>`;
    }
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

  $: {
    if (table) {
      table.setDataListener(reactiveDataListener);
      table.draw();
    }
  }

  $: {
    if (table) {
      table.addStyleListener(() => {
        for (const td of table.querySelectorAll("tbody td")) {
          map_td(td);
        }

        for (const th of table.querySelectorAll("tbody th")) {
          map_row_th(th);
        }

        for (const th of table.querySelectorAll("thead th")) {
          map_column_th(th);
        }
        dispatch("pos", getCachedPos());
      });
    }
  }

  // Can experiment with debouncing forced draws from async data arriving
  let t = null;
  const forceDraw = () => {
    if (t) clearTimeout(t);
    t = setTimeout(() => {
      console.log("FORCE DRAW");
      table.draw();
    }, 0);
  };

  // let lastSeenRowQuery = null;
  // const getLastSeenRowQuery = () => lastSeenRowQuery;
  // $: {
  //   // If data and we haven't drawn this data already, redraw the table
  //   if (
  //     $rowHeaderQuery.data &&
  //     $rowHeaderQuery.data !== getLastSeenRowQuery()
  //   ) {
  //     lastSeenRowQuery = $rowHeaderQuery.data;

  //     console.log("DRAW ROWS");
  //     table.draw();
  //     // forceDraw();
  //   }
  // }

  const rowDataStore = rowHeaderDataProvider.data;
  $: {
    if (table && $rowDataStore) {
      console.count("DRAW ROWS");
      table.draw();
    }
  }

  const columnDataStore = columnHeaderDataProvider.data;
  $: {
    if (table && $columnDataStore) {
      console.count("DRAW COLUMNS");
      // TODO: any way to skip this draw if we know reading from cache will suffice?
      table.draw();
      // forceDraw();

      // Example of infinite scroll. expand available columns when reaching end
      // if ($columnHeaderQuery.data.block[0] >= 175) {
      //   metadata.colCt = 500;
      // }
    }
  }

  const bodyDataStore = bodyDataProvider.data;
  $: {
    if (table && $bodyDataStore) {
      console.count("DRAW BODY");
      table.draw();
    }
  }

  // let lastSeenBodyData = null;
  // const getLastSeenBodyData = () => lastSeenBodyData;
  // $: {
  //   if ($bodyQuery.data && getLastSeenBodyData() !== $bodyQuery.data) {
  //     lastSeenBodyData = $bodyQuery.data;
  //     // table.draw();
  //     // console.log("DRAW BODY");
  //   }
  // }

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
  :global(
      regular-table
        thead
        tr:last-child
        th:not([__col="0"]):not(.rt-group-corner)
    ) {
    min-width: 70px !important;
    width: 70px;
    max-width: 70px;
  }

  :global(regular-table thead tr:last-child th[__col="0"]) {
    min-width: 92px !important;
    width: 92px;
    max-width: 92px;
  }

  :global(regular-table tbody th[__col="0"]) {
    /* color: red; */
    /* border-bottom: 1px solid black; */
  }
</style>
