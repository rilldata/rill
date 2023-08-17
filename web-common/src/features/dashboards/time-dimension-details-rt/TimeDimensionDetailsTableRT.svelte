<script lang="ts">
  import "regular-table";
  import "regular-table/dist/css/material.css";

  let table;

  const DATA = [
    [0, 1, 2, 3, 4, 5],
    ["A", "B", "C", "D", "E", "F"],
    [true, false, true, false, true, false],
  ];

  const NUM_ROWS = 1000;
  const NUM_COLUMNS = 106;
  function range(x0, x1, f) {
    return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
  }
  function dataListener(num_rows, num_columns) {
    return (x0, y0, x1, y1) => ({
      num_rows,
      num_columns,
      // row_headers: range(y0, y1, (d) => d),
      // column_headers: range(x0, x1, (d) => d),
      data: range(x0, x1, (x) => range(y0, y1, (y) => `${x},${y}`)),
    });
  }

  // function getDataSlice(x0, y0, x1, y1) {
  //   console.log("getDataSlice", x0, y0, x1, y1);
  //   return {
  //     num_rows: DATA[0].length,
  //     num_columns: DATA.length,
  //     data: DATA.slice(x0, x1).map((col) => col.slice(y0, y1)),
  //   };
  // }

  $: {
    console.log(table);
    if (table) {
      function style_td(td) {
        const meta = table.getMeta(td);
        td.setAttribute("__col", String(meta.x));
        td.setAttribute("__row", String(meta.y));
        td.classList.toggle("bg-slate-50", meta.x > 0 && meta.x < 6);
        if (meta.x === 1) {
          td.innerHTML = `<div class="flex">#,### <svg height="24px" width="60px" viewBox="0 0 9 15" preserveAspectRatio="none">
  <path
    d="M 0 14 L 1 15 L 2 10 L 3 11 L 4 7 L 5 5 L 6 0 L 7 5 L 8 10 L 9 11"
    stroke-width="1"
    stroke="black"
    fill="none"
    vector-effect="non-scaling-stroke"
  />
</svg></div>`;
        }
      }
      table.setDataListener(dataListener(NUM_ROWS, NUM_COLUMNS));
      table.addStyleListener(() => {
        for (const td of table.querySelectorAll("tbody td")) {
          style_td(td);
        }
      });
      table.draw();
    }
  }
</script>

<regular-table
  class="border m-8"
  style="height: 400px; width: 2000px;"
  bind:this={table}
/>

<style>
  :global(regular-table td) {
    height: 24px;
    width: 60px;
    min-width: 60px;
    max-width: 60px;
    padding: 0px;
  }

  /* const columnWidths = [120, 126, 64, 64, 64]; */

  :global(regular-table td[__col="0"]) {
    width: 120px;
    min-width: 120px;
    max-width: 120px;
  }

  :global(regular-table td[__col="1"]) {
    width: 126px;
    min-width: 126px;
    max-width: 126px;
  }

  :global(regular-table td[__col="2"]) {
    width: 64px;
    min-width: 64px;
    max-width: 64px;
  }

  :global(regular-table td[__col="3"]) {
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
