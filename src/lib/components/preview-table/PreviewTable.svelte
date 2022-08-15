<!-- @component 
Creates a virtualized preview table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
RowHeaders – a sticky row number header.
TableCells – the cell contents.
PinnedColumns – any reference columns pinned on the right side of the overall table.
-->
<script lang="ts">
  import { TIMESTAMPS } from "$lib/duckdb-data-types";

  import type { ProfileColumn } from "$lib/types";

  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { config } from "./config";
  import ColumnHeaders from "./sections/ColumnHeaders.svelte";
  import PinnedColumns from "./sections/PinnedColumns.svelte";
  import RowHeaders from "./sections/RowHeaders.svelte";
  import TableCells from "./sections/TableCells.svelte";

  export let rows;
  export let columnNames: ProfileColumn[];
  /** if this is set to true, we will use the data passed in as rows
   * to calculate the column widths. Otherwise, we use the table / view's
   * largest values in each column, which is useful if we're building an
   * infinite-scroll table and need to compute the largest possible column width
   * ahead of time.
   */
  export let inferColumnWidthFromData = true;

  let rowVirtualizer;
  let columnVirtualizer;
  let container;
  let pinnedColumns = [];
  let virtualRows;
  let virtualColumns;
  let virtualWidth;
  let virtualHeight;

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_WIDTH = 7;
  const CHARACTER_X_PAD = 16 * 2;
  const HEADER_ICON_WIDTHS = 16 * 2;
  const HEADER_X_PAD = CHARACTER_X_PAD;
  const HEADER_FLEX_SPACING = 16 * 2;

  $: if (rows && columnNames) {
    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => config.rowHeight,
      overscan: 40,
      paddingStart: config.rowHeight,
    });

    /** if we're inferring the column widths from static-ish data, let's
     * find the largest strings in the column and use that to bootstrap the
     * column widths.
     */
    let columnWidths: { [key: string]: number } = {};
    if (inferColumnWidthFromData) {
      columnNames.forEach((column) => {
        // get values
        const largest = Math.max(
          ...rows.map((row) => `${row[column.name]}`.length)
        );
        columnWidths[column.name] = TIMESTAMPS.has(column.type) ? 22 : largest;
      });
    }

    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columnNames.length,
      getItemKey: (index) => columnNames[index].name,
      estimateSize: (index) => {
        const column = columnNames[index];
        /** if we are inferring column widths from the data,
         * let's utilize columnWidths, calculated above.
         */
        const largestStringLength =
          (inferColumnWidthFromData
            ? columnWidths[column.name]
            : column?.largestStringLength) *
            CHARACTER_WIDTH +
          CHARACTER_X_PAD;

        return largestStringLength
          ? /** the largest value for a column should be config.maxColumnWidth.
             * the smallest value should either be the largestStringLength (which comes from the actual)
             * table values, the header string, or the config.minColumnWidth.
             */
            Math.min(
              config.maxColumnWidth,
              Math.max(
                largestStringLength,
                column.name.length * CHARACTER_WIDTH +
                  HEADER_ICON_WIDTHS +
                  HEADER_X_PAD +
                  HEADER_FLEX_SPACING,
                config.minColumnWidth
              )
            )
          : /** if there isn't a longet string length for some reason, let's go with a
             * default column width. We should not be in this state.
             */
            config.defaultColumnWidth;
      },
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
        virtualColumnItems={virtualColumns}
        virtualRowItems={virtualRows}
        on:pin={handlePin}
        on:inspect={setActiveIndex}
      />
    {/if}
  {/if}
</div>
