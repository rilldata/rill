<!-- @component 
Creates a virtualized preview table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
RowHeaders – a sticky row number header.
TableCells – the cell contents.
PinnedColumns – any reference columns pinned on the right side of the overall table.
-->
<script lang="ts">
  import {
    DATES,
    TIMESTAMPS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { setContext } from "svelte";
  import { tweened } from "svelte/motion";
  import ColumnHeaders from "../virtualized-table/sections/ColumnHeaders.svelte";
  import PinnedColumns from "../virtualized-table/sections/PinnedColumns.svelte";
  import RowHeaders from "../virtualized-table/sections/RowHeaders.svelte";
  import TableCells from "../virtualized-table/sections/TableCells.svelte";
  import { config as defaultConfig } from "./config";
  import type { VirtualizedTableConfig } from "../virtualized-table/types";

  export let rows;
  export let configOverride: Partial<VirtualizedTableConfig> = {};
  export let columnNames: VirtualizedTableColumns[];

  /** the overscan values tell us how much to render off-screen. These may be set by the consumer
   * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
   * to render on initial load.
   */
  export let rowOverscanAmount = 40;
  export let columnOverscanAmount = 5;

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

  const config = {
    ...defaultConfig,
    ...configOverride,
  };

  /* set context for child components */
  setContext("config", config);

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_WIDTH = 7;
  const CHARACTER_X_PAD = 16 * 2;
  const HEADER_ICON_WIDTHS = 16;
  const HEADER_X_PAD = CHARACTER_X_PAD;
  const HEADER_FLEX_SPACING = 16;

  $: rowScrollOffset = 0;
  $: colScrollOffset = 0;

  let manuallyResizedColumns = tweened({});
  $: if (rows && columnNames) {
    // initialize resizers?
    if (Object.keys(manuallyResizedColumns).length === 0) {
      manuallyResizedColumns = tweened(
        columnNames.reduce((tbl, column) => {
          tbl[column.name] = undefined;
          return tbl;
        }),
        { duration: 200 }
      );
    }

    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => config.rowHeight,
      overscan: rowOverscanAmount,
      paddingStart: config.rowHeight,
      initialOffset: rowScrollOffset,
    });

    /** if we're inferring the column widths from static-ish data, let's
     * find the largest strings in the column and use that to bootstrap the
     * column widths.
     */
    let columnWidths: { [key: string]: number } = {};
    if (inferColumnWidthFromData) {
      columnNames.forEach((column) => {
        // get values
        const values = rows
          .filter((row) => row[column.name] !== null)
          .map((row) => `${row[column.name]}`.length);
        values.sort();
        let largest = Math.max(...values);

        // const largest = Math.max(values);
        columnWidths[column.name] = largest;

        if (TIMESTAMPS.has(column.type)) {
          columnWidths[column.name] = DATES.has(column.type) ? 13 : 22;
        }
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

        /** The header width is largely a function of the total number of characters in the column.*/
        const headerWidth =
          column.name.length * CHARACTER_WIDTH +
          HEADER_ICON_WIDTHS +
          HEADER_X_PAD +
          HEADER_FLEX_SPACING;

        /** If the header is bigger than the largestStringLength and that's not at threshold, default to threshold.
         * This will prevent the case where we have very long column names for very short column values.
         */
        let effectiveHeaderWidth =
          headerWidth > 160 && largestStringLength < 160
            ? config.minHeaderWidthWhenColumsAreSmall
            : headerWidth;

        let hasUserDefinedColumnWidth =
          $manuallyResizedColumns[column.name] !== undefined;

        return largestStringLength
          ? /** the largest value for a column should be config.maxColumnWidth.
             * the smallest value should either be the largestStringLength (which comes from the actual)
             * table values, the header string, or the config.minColumnWidth.
             */
            Math.min(
              /** Define the maximum column size. If the user has set the column width, go with that (meaning columns
               * can be infinitely large if the user wants it). Otherwise, set a default width that is sensible.
               */
              hasUserDefinedColumnWidth ? Infinity : config.maxColumnWidth,
              /** If iuser has set the column width, we'll go with that (as long as it is larger than minColumnWidth).
               * If they haven't set it, we'll go with the effectiveHeaderWidth.
               * In the case of TIMESTAMP columns, we are effectively skipping out on worrying about the header column
               * and going strictly with a fixed-width based on the time stamp representation.
               */
              hasUserDefinedColumnWidth
                ? $manuallyResizedColumns[column.name]
                : Math.max(
                    largestStringLength,
                    /** use effective header width, unless its a timestamp, in which case just use largest string length */
                    TIMESTAMPS.has(column.type) ? 0 : effectiveHeaderWidth,
                    /** All columns must be minColumnWidth regardless of user settings. */
                    config.minColumnWidth
                  )
            )
          : /** if there isn't a longet string length for some reason, let's go with a
             * default column width. We should not be in this state.
             */
            config.defaultColumnWidth;
      },
      overscan: columnOverscanAmount,
      paddingStart: config.indexWidth,
      initialOffset: colScrollOffset,
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

  async function handleResizeColumn(event) {
    rowScrollOffset = $rowVirtualizer.scrollOffset;
    colScrollOffset = $columnVirtualizer.scrollOffset;

    const { size, name } = event.detail;
    manuallyResizedColumns.update((state) => {
      state[name] = Math.max(config.minColumnWidth, size);
      return state;
    });
  }

  async function handleResetColumnSize(event) {
    const { name } = event.detail;
    manuallyResizedColumns.update((state) => {
      state[name] = undefined;
      return state;
    });
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
      class="relative surface"
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
        showDataIcon={true}
        {pinnedColumns}
        on:pin={handlePin}
        on:resize-column={handleResizeColumn}
        on:reset-column-size={handleResetColumnSize}
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
