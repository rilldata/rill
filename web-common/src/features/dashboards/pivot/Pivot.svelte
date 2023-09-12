<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";
  import { basicPivot } from "./configs";
  import { getColumnHeaders, getMetadata, getRowHeaders } from "./api";
  import { get, writable } from "svelte/store";
  import { createEventDispatcher } from "svelte";

  // TODO: replace with w-full once you have fixed widths
  const LOADING_EL = `<div class="w-8 h-4 bg-gray-100 rounded loading-cell"/>`;

  function range(x0, x1, f) {
    return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
  }

  export let rowHeaderData = {
    block: [0, 0],
    data: [],
  };

  export let getRowHeaderData = (pos) => null;

  const dispatch = createEventDispatcher();

  const config = basicPivot;
  $: metadata = getMetadata(config);

  // util
  const isEmptyPos = (pos) =>
    pos.x0 === 0 && pos.x1 === 0 && pos.y0 === 0 && pos.y1 === 0;

  let table;
  const store = writable({
    pos: {
      x0: 0,
      x1: 0,
      y0: 0,
      y1: 0,
    },
    cachedBlock: null,
  });

  function dataListener(num_rows, num_columns) {
    return (x0, y0, x1, y1) => {
      if (!isEmptyPos({ x0, x1, y0, y1 })) {
        store.update((s) => {
          s.pos = { x0, x1, y0, y1 };
          // s.cachedBlock = getRowHeaderData(s.pos);
          return s;
        });
        // dispatch("pos", { x0, x1, y0, y1 });
      }
      return {
        num_rows,
        num_columns,
        data: range(x0, x1, (x) => range(y0, y1, (y) => `${y},${x}`)),
        row_headers: getRowHeaders(config, y0, y1),
        column_headers: getColumnHeaders(config, x0, x1),
      };
    };
  }

  function renderRowHeader(cell, data, log) {
    const x = Number(cell.getAttribute("__col"));
    const y = Number(cell.getAttribute("__row"));
    if (data) {
      const yOffset = y - data.block[0];
      try {
        const v = data.data[yOffset][x];
        cell.innerHTML = v;
      } catch (e) {
        console.log({
          e,
          data,
          y,
          yOffset,
          log,
        });
      }
    } else cell.innerHTML = LOADING_EL; // "loading...";
  }

  $: {
    // If data and don't have a cachedBlock that was already rendered on scroll, render the cells
    // TODO: this cachedBlock check might break for when data is invalidated by filter? need to check
    if (
      rowHeaderData &&
      rowHeaderData.data.length > 0 &&
      !get(store).cachedBlock // Tried removing cached block, but then other stuff broke :/
    ) {
      const data = getRowHeaderData(get(store).pos);
      /*
        Problem: sometimes reactive statement tries to run before the new pos rows are drawn
        The reason the cachedBlock thing fixes this is that it skips the instant update from cache
        that tanstack is giving otherwise here.

        So, def need to know if rendering from cache or not
        
        Other ideas: 
        - check if rows are drawn yet?
        - dont update pos externally until AFTER rows are drawn, via the style listener
        ^ this seems to work. review tomorrow
      */
      for (const th of table.querySelectorAll("tbody th")) {
        renderRowHeader(th, data, "reactive");
        // renderRowHeader(th, rowHeaderData);
      }
    }
  }

  function map_row_th(th, data) {
    const meta = table.getMeta(th);
    const x = meta.row_header_x;
    const y = meta.y;
    th.setAttribute("__col", String(x));
    th.setAttribute("__row", String(y));
    renderRowHeader(th, data, "map");
  }

  $: {
    function map_td(td) {
      const meta = table.getMeta(td);
      const x = meta.x;
      const y = meta.y;
      td.setAttribute("__col", String(x));
      td.setAttribute("__row", String(y));
    }

    function map_column_th(th) {
      const meta = table.getMeta(th);
      const x = meta.x;
      const y = meta.column_header_y;
      th.setAttribute("__col", String(x));
      th.setAttribute("__row", String(y));
    }

    if (table) {
      table.setDataListener(dataListener(metadata.rowCt, metadata.colCt));
      table.addStyleListener(() => {
        const data = getRowHeaderData(get(store).pos);
        for (const td of table.querySelectorAll("tbody td")) {
          map_td(td);
        }

        for (const th of table.querySelectorAll("tbody th")) {
          map_row_th(th, data);
        }

        for (const th of table.querySelectorAll("thead th")) {
          map_column_th(th);
        }
        dispatch("pos", get(store).pos);
      });

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
</style>
