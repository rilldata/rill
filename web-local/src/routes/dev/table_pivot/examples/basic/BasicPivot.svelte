<script lang="ts">
  import { get, writable } from "svelte/store";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import { getMetadata } from "../mock-api";
  import { basicNestedPivot, basicPivot } from "../configs";
  import {
    createColumnHeaderDataGetter,
    createRowHeaderDataGetter,
    getBodyData,
  } from "./data-providers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  let config = writable(structuredClone(basicPivot));
  $: metadata = getMetadata($config);
  let table;

  $: getRowHeaderData = createRowHeaderDataGetter($config);
  $: getColumnHeaderData = createColumnHeaderDataGetter($config);

  const renderCell = (data) => {
    // Can add styles/classes directly to the cell
    data.element.classList.toggle("bg-gray-50", data.y % 2);
    data.element.classList.add("text-right");

    // Can manipulate attributes of cell
    // data.element.setAttribute(...);

    // Can return a string of HTML to render in the cell instead of using data value
    // return data.y % 2
    //   ? `<div class="text-red-500">${data.value}</div>`
    //   : data.value;
  };

  const renderColumnHeader = (data) => {
    if (data.y === 1) {
      data.element.classList.add("text-right");
    }
  };

  const renderFlatRowHeader = ({ element }) => {
    element.classList.add("truncate");
  };

  const renderNestedRowHeader = (data) => {
    data.element.classList.add("truncate");
    if (data.value === "") return data.value;
    if (data.value.expandable) {
      const marker = data.value.isExpanded ? "-" : "+";
      return `<div class='w-full h-full' data-expandable="${data.value.idx}">${marker} ${data.value.text}</div>`;
    }
    return data.value.text;
  };
  $: renderRowHeader =
    $config.rowJoinType === "flat"
      ? renderFlatRowHeader
      : renderNestedRowHeader;

  const getColumnWidth = (x: number) => {
    if (x === 0) return 100;
    return 70;
  };

  let getRowHeaderWidth = (x: number) => 100;
  $: {
    if ($config.rowJoinType === "flat" || $config.expanded.length === 0) {
      getRowHeaderWidth = (x: number) => 100;
    } else getRowHeaderWidth = (x: number) => [24, 100][x];
  }

  const handleMouseDown = (evt, table) => {
    if (evt.target.hasAttribute("data-expandable")) {
      const idx = parseInt(evt.target.getAttribute("data-expandable"));
      config.update((c) => {
        if (c.expanded.includes(idx)) {
          c.expanded = c.expanded.filter((i) => i !== idx);
        } else {
          c.expanded.push(idx);
        }
        return c;
      });
    }
  };

  $: renderRowCorner = ({ x, element, y }) => {
    if (y === 1) {
      element.classList.add("text-left");
      if ($config.rowJoinType === "flat") return ["Product", "Country"][x];
      return x === 0 ? "Product > Country" : undefined;
    }
  };

  // Make sure to force a redraw when the data changes, or the scroll position could get messed up
  $: {
    $config;
    if (table) table.draw();
  }
</script>

<div class="flex flex-col gap-2">
  <div class="flex items-center gap-2">
    <Button on:click={() => config.set(structuredClone(basicPivot))}
      >Flat rows</Button
    >
    <Button on:click={() => config.set(structuredClone(basicNestedPivot))}
      >Nested rows</Button
    >
  </div>
  <Pivot
    bind:api={table}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    rowCount={metadata.rowCt}
    columnCount={metadata.colCt}
    rowHeaderDepth={2}
    columnHeaderDepth={2}
    {renderCell}
    {renderColumnHeader}
    {renderRowHeader}
    {renderRowCorner}
    {getColumnWidth}
    {getRowHeaderWidth}
    onMouseDown={handleMouseDown}
  />
</div>

<style>
  /* Example of using custom classes. Global classnames must be used */
  :global(.basic-pivot-column) {
  }
</style>
