<!-- @component 
Creates a virtualized preview table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
RowHeaders – a sticky row number header.
TableCells – the cell contents.
PinnedColumns – any reference columns pinned on the right side of the overall table.
-->
<script lang="ts">
  import type { ProfileColumn } from "$lib/types";

  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import ColumnHeaders from "./ColumnHeaders.svelte";
  import { config } from "./config";
  import PinnedColumns from "./PinnedColumns.svelte";
  import RowHeaders from "./RowHeaders.svelte";
  import TableCells from "./TableCells.svelte";

  export let rows;
  export let columnNames: ProfileColumn[];

  let rowVirtualizer;
  let columnVirtualizer;
  let container;
  let pinnedColumns = [];
  let virtualRows;
  let virtualColumns;
  let virtualWidth;
  let virtualHeight;

  $: if (rows && columnNames) {
    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => config.rowHeight,
      overscan: 90,
      paddingStart: config.rowHeight,
    });
    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columnNames.length,
      estimateSize: () => config.columnWidth,
      overscan: 10,
      paddingStart: config.indexWidth,
    });
  }

  $: if (rowVirtualizer) {
    virtualRows = $rowVirtualizer.getVirtualItems();
    virtualHeight = $rowVirtualizer.getTotalSize();
  }
  $: if (columnVirtualizer) {
    virtualColumns = $columnVirtualizer.getVirtualItems();
    virtualWidth = $columnVirtualizer.getTotalSize();
  }

  let activeIndex;
  function setActiveIndex(event) {
    activeIndex = event.detail;
  }
  function clearActiveIndex() {
    activeIndex = false;
  }

  /** handle scrolling tooltip suppression */
  let scrolling = false;
  let timeoutID;
  $: {
    if (scrolling) {
      if (timeoutID) clearTimeout(timeoutID);
      timeoutID = setTimeout(() => {
        scrolling = false;
      }, 200);
    }
  }

  /** pinning functionality */
  function handlePin(event) {
    const column = event.detail;
    if (pinnedColumns.some((p) => p.name === column.name)) {
      pinnedColumns = [...pinnedColumns.filter((c) => c.name !== column.name)];
    } else {
      pinnedColumns = [...pinnedColumns, column];
    }
  }
</script>

<div
  bind:this={container}
  style:width="100%"
  style:height="100%"
  class="overflow-auto grid"
  style:grid-template-columns="max-content auto"
  on:scroll={() => {
    /** capture to suppress cell tooltips. Otherwise,
     * there's quite a bit of rendering jank.
     */
    scrolling = true;
  }}
>
  {#if rowVirtualizer}
    <div
      class="relative bg-white"
      on:mouseleave={clearActiveIndex}
      on:blur={clearActiveIndex}
      style:will-change="transform, contents"
      style:width="{virtualWidth}px"
      style:height="{virtualHeight}px"
    >
      <!-- ColumnHeaders -->
      <ColumnHeaders
        virtualColumnItems={virtualColumns}
        columns={columnNames}
        {pinnedColumns}
        on:pin={handlePin}
      />
      <!-- RowHeader -->
      <RowHeaders virtualRowItems={virtualRows} totalHeight={virtualHeight} />
      <!-- VirtualTableBody -->
      <TableCells
        virtualColumnItems={virtualColumns}
        virtualRowItems={virtualRows}
        {rows}
        columns={columnNames}
        {activeIndex}
        {scrolling}
        on:inspect={setActiveIndex}
      />
    </div>
    <!-- PinnedContent -->
    {#if pinnedColumns.length}
      <PinnedColumns
        {rows}
        {pinnedColumns}
        {scrolling}
        {activeIndex}
        virtualRowItems={virtualRows}
        on:pin={handlePin}
        on:inspect={setActiveIndex}
      />
    {/if}
  {/if}
</div>
