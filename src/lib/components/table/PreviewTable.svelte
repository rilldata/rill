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

  let activeIndex;
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
  function handlePin(column) {
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
      style:width="{$columnVirtualizer.getTotalSize()}px"
      style:height="{$rowVirtualizer.getTotalSize()}px"
    >
      <!-- ColumnHeaders -->
      <ColumnHeaders
        virtualColumnItems={$columnVirtualizer.getVirtualItems()}
        columns={columnNames}
        {pinnedColumns}
        on:pin={(event) => {
          handlePin(event.detail);
        }}
      />
      <!-- RowHeader -->
      <RowHeaders
        virtualRowItems={$rowVirtualizer.getVirtualItems()}
        totalHeight={$rowVirtualizer.getTotalSize()}
      />
      <!-- VirtualTableBody -->
      <TableCells
        virtualColumnItems={$columnVirtualizer.getVirtualItems()}
        virtualRowItems={$rowVirtualizer.getVirtualItems()}
        {rows}
        columns={columnNames}
        {activeIndex}
        {scrolling}
        on:inspect={(event) => {
          activeIndex = event.detail;
        }}
      />
    </div>
    <!-- PinnedContent -->
    {#if pinnedColumns.length}
      <PinnedColumns
        {rows}
        {pinnedColumns}
        {scrolling}
        {activeIndex}
        virtualRowItems={$rowVirtualizer.getVirtualItems()}
        on:pin={(event) => handlePin(event.detail)}
        on:inspect={(event) => {
          activeIndex = event.detail;
        }}
      />
    {/if}
  {/if}
</div>
