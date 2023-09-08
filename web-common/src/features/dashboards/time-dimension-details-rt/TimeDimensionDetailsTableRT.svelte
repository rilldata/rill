<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";
  import { getBlock } from "../time-dimension-details/util";
  import {
    CreateQueryResult,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { fetchData } from "../time-dimension-details/mock-data";
  import { get, writable } from "svelte/store";

  let table;
  const __state = writable({ cachedBlock: null });
  const NUM_ROWS = 1000;
  const NUM_COLUMNS = 105;
  const STICKY_COLUMNS = 5;
  function range(x0, x1, f) {
    return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
  }

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };

  let debouncedPos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const debounce = (cb, t) => {
    let timeout = null;
    return (...args) => {
      const next = () => cb(...args);
      clearTimeout(timeout);
      timeout = setTimeout(next, t);
    };
  };

  const updateDebouncedPos = debounce((pos) => {
    debouncedPos = pos;
  }, 0);

  $: {
    const posIsEmpty = pos.x1 - pos.x0 === 0 || pos.y1 - pos.y0 === 0;
    if (!posIsEmpty) {
      updateDebouncedPos(pos);
    }
  }

  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  let block = getBlock(100, debouncedPos.y0, debouncedPos.y1);
  $: {
    const nextBlock = getBlock(100, debouncedPos.y0, debouncedPos.y1);
    if (block[0] !== nextBlock[0] || block[1] !== nextBlock[1]) {
      block = nextBlock;
    }
  }

  let cellQuery = createQuery({
    queryKey: ["time-dimension-details", block[0], block[1]],
    queryFn: fetchData(block, 1000),
  }) as CreateQueryResult<{
    block: number[];
    data: { text?: string; value?: number; sparkline?: number[] }[][];
  }>;

  $: cellQuery = createQuery({
    queryKey: ["time-dimension-details", block[0], block[1]],
    queryFn: fetchData(block, 1000),
  }) as CreateQueryResult<{
    block: number[];
    data: { text?: string; value?: number; sparkline?: number[] }[][];
  }>;

  const columnRenderers = [
    (d) => `<div class="">${d.text}</div>`,
    (d) =>
      `<div class="flex items-center justify-end">${d.value.toFixed(
        0
      )} <svg height="24px" width="60px" viewBox="0 0 9 15" preserveAspectRatio="none"> <path  d="M 0 14 L 1 15 L 2 10 L 3 11 L 4 7 L 5 5 L 6 0 L 7 5 L 8 10 L 9 11"  stroke-width="1"  stroke="black"  fill="none"  vector-effect="non-scaling-stroke" /></svg></div>`,
    (d) => (d.value * 10).toFixed(2) + "%",
    (d) =>
      `<span class="${d.value < 0 ? "text-red-500" : ""}">${d.value.toFixed(
        2
      )}</span>`,
    (d) =>
      `<span class="${d.value < 0 ? "text-red-500" : ""}">${
        (d.value * 10).toFixed(2) + "%"
      }</span>`,
  ];
  const defaultRender = (d) => d.value.toFixed(2);

  function renderCell(cell, data) {
    // read __row and __col attribute of td
    const x = Number(cell.getAttribute("__col"));
    const y = Number(cell.getAttribute("__row"));
    if (data) {
      // if (data && data.block[0] <= y && data.block[1] > y) {
      const yOffset = y - data.block[0];
      const v = data.data[yOffset][x];
      cell.innerHTML = columnRenderers[x]?.(v) ?? defaultRender(v);
    } else
      cell.innerHTML = `<div class="w-full h-4 bg-gray-100 rounded loading-cell"/>`; // "...";
  }

  $: {
    const data = $cellQuery.data;
    // If data and don't have a cachedBlock that was already rendered on scroll, render the cells
    if (data && table && !get(__state).cachedBlock) {
      for (const cell of table.querySelectorAll("tbody td, tbody th")) {
        renderCell(cell, data);
      }
    }
  }

  $: {
    if (table) {
      function style_td(td) {
        const meta = table.getMeta(td);
        const x = meta.x + STICKY_COLUMNS;
        const y = meta.y;
        td.setAttribute("__col", String(x));
        td.setAttribute("__row", String(y));
        const data = get(__state).cachedBlock;
        renderCell(td, data);
      }

      function style_th(th) {
        const meta = table.getMeta(th);
        const x = meta.row_header_x;
        const y = meta.y;
        th.setAttribute("__col", String(x));
        th.setAttribute("__row", String(y));
        th.classList.toggle("bg-slate-50", x > 0 && x < 6);
        const data = get(__state).cachedBlock;
        renderCell(th, data);
      }

      function dataListener(num_rows, num_columns) {
        return (x0, y0, x1, y1) => {
          pos = { x0, x1, y0, y1 };
          const block = getBlock(100, y0, y1);
          const cachedBlock = cache.find([
            "time-dimension-details",
            block[0],
            block[1],
          ])?.state?.data;
          __state.set({ cachedBlock });
          return {
            num_rows,
            num_columns,
            data: range(x0, x1, (x) =>
              range(y0, y1, (y) => `${x + STICKY_COLUMNS},${y}`)
            ),
            row_height: 24,
            row_headers: range(y0, y1, (y) =>
              new Array(STICKY_COLUMNS).fill(0).map((n, i) => [`${i},${y}`])
            ),
            column_headers: new Array(NUM_COLUMNS - STICKY_COLUMNS)
              .fill(0)
              .map((n, i) => [`${i + x0}`]),
          };
        };
      }

      table.setDataListener(
        dataListener(NUM_ROWS, NUM_COLUMNS - STICKY_COLUMNS)
      );
      table.addStyleListener(() => {
        for (const td of table.querySelectorAll("tbody td")) {
          style_td(td);
        }

        for (const th of table.querySelectorAll("tbody th")) {
          style_th(th);
        }
      });
      table.draw();
    }
  }
</script>

<div class="border m-8 relative" style="height: 400px; width: 100%">
  <regular-table class="w-full h-full tdd-table" bind:this={table} />
</div>

<style>
  :global(regular-table tr td) {
    height: 24px;
    width: 60px;
    min-width: 60px;
    max-width: 60px;
    padding-block: 0px;
    text-align: right;
    background-color: white;
  }

  :global(.tdd-table tbody tr td) {
    background-color: white;
    color: black;
    padding-inline: 8px;
  }

  :global(.tdd-table tbody tr th) {
    color: black;
    font-weight: normal;
    padding-inline: 8px;
  }

  :global(.tdd-table tbody tr th:nth-child(1)) {
    text-align: left;
  }

  :global(regular-table th) {
    height: 24px;
    width: 60px;
    min-width: 60px;
    max-width: 60px;
    padding: 0px;
    text-align: right;
  }

  :global(regular-table tbody th) {
    height: 24px;
    text-align: right;
  }

  /* const columnWidths = [120, 126, 64, 64, 64]; */

  :global(regular-table th[__col="0"]) {
    width: 120px;
    min-width: 120px;
    max-width: 120px;
    left: 0px;
  }

  :global(regular-table th[__col="1"]) {
    width: 126px;
    min-width: 126px;
    max-width: 126px;
  }

  :global(regular-table th[__col="2"]) {
    width: 64px;
    min-width: 64px;
    max-width: 64px;
  }

  :global(regular-table th[__col="3"]) {
    width: 64px;
    min-width: 64px;
    max-width: 64px;
  }

  :global(regular-table td[__col="4"]) {
    width: 64px;
    min-width: 64px;
    max-width: 64px;
  }

  :global(regular-table table) {
    table-layout: fixed;
    font-family: Inter;
    font-feature-settings: "case" 0, "cpsp" 0, "dlig" 0, "frac" 0, "dnom" 0,
      "numr" 0, "salt" 0, "subs" 0, "sups" 0, "tnum", "zero" 0, "ss01", "ss02" 0,
      "ss03" 0, "ss04" 0, "cv01" 0, "cv02" 0, "cv03" 0, "cv04" 0, "cv05" 0,
      "cv06" 0, "cv07" 0, "cv08" 0, "cv09" 0, "cv10" 0, "cv11" 0, "calt", "ccmp",
      "kern";
    /* width: 2000px; */
  }

  :global(regular-table *) {
    box-sizing: border-box;
  }
</style>
