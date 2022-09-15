<!-- @component 
Creates a virtualized dimension table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
TableCells – the cell contents.
-->
<script lang="ts">
  import { setContext } from "svelte";
  import { createEventDispatcher } from "svelte";
  import type { VirtualizedTableColumns } from "$lib/types";

  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { tweened } from "svelte/motion";
  import { DimensionTableConfig } from "./DimensionTableConfig";
  import ColumnHeaders from "$lib/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "$lib/components/virtualized-table/sections/TableCells.svelte";

  const dispatch = createEventDispatcher();

  export let rows;
  export let columns: VirtualizedTableColumns[];
  export let activeValues: Array<unknown> = [];
  export let sortByColumn: string;

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
  let virtualRows;
  let virtualColumns;
  let virtualWidth;
  let virtualHeight;
  let containerWidth;

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_WIDTH = 7;
  const CHARACTER_X_PAD = 16 * 2;
  const HEADER_ICON_WIDTHS = 16;
  const HEADER_X_PAD = CHARACTER_X_PAD;
  const HEADER_FLEX_SPACING = 14;

  /* set context for child components */
  setContext("config", DimensionTableConfig);

  $: dimensionName = columns[0]?.name;
  $: selectedIndex = activeValues
    .map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    })
    .filter((i) => i >= 0);

  $: rowScrollOffset = 0;
  $: colScrollOffset = 0;
  let manuallyResizedColumns = tweened({});
  $: if (rows && columns) {
    // initialize resizers?
    if (Object.keys(manuallyResizedColumns).length === 0) {
      manuallyResizedColumns = tweened(
        columns.reduce((tbl, column) => {
          tbl[column.name] = undefined;
          return tbl;
        }),
        { duration: 200 }
      );
    }

    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => DimensionTableConfig.rowHeight,
      overscan: rowOverscanAmount,
      paddingStart: DimensionTableConfig.columnHeaderHeight,
      initialOffset: rowScrollOffset,
    });

    /** if we're inferring the column widths from static-ish data, let's
     * find the largest strings in the column and use that to bootstrap the
     * column widths.
     */
    let columnWidths: { [key: string]: number } = {};
    if (inferColumnWidthFromData) {
      columns.forEach((column) => {
        // get values
        const values = rows
          .filter((row) => row[column.name] !== null)
          .map(
            (row) =>
              `${row["__formatted_" + column.name] || row[column.name]}`.length
          );
        values.sort();
        let largest = Math.max(...values);
        columnWidths[column.name] = largest;
      });
    }

    const estimateColumnSize = columns.map((column) => {
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
        (column.label?.length || column.name.length) * CHARACTER_WIDTH +
        HEADER_ICON_WIDTHS +
        HEADER_X_PAD +
        HEADER_FLEX_SPACING;

      /** If the header is bigger than the largestStringLength and that's not at threshold, default to threshold.
       * This will prevent the case where we have very long column names for very short column values.
       */
      let effectiveHeaderWidth =
        headerWidth > 160 && largestStringLength < 160
          ? DimensionTableConfig.minHeaderWidthWhenColumsAreSmall
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
            hasUserDefinedColumnWidth
              ? Infinity
              : DimensionTableConfig.maxColumnWidth,
            /** If user has set the column width, we'll go with that (as long as it is larger than minColumnWidth).
             * If they haven't set it, we'll go with the effectiveHeaderWidth.
             */
            hasUserDefinedColumnWidth
              ? $manuallyResizedColumns[column.name]
              : Math.max(
                  largestStringLength,
                  effectiveHeaderWidth,
                  /** All columns must be minColumnWidth regardless of user settings. */
                  DimensionTableConfig.minColumnWidth
                )
          )
        : /** if there isn't a longet string length for some reason, let's go with a
           * default column width. We should not be in this state.
           */
          DimensionTableConfig.defaultColumnWidth;
    });

    const measureColumnSizeSum = estimateColumnSize
      .slice(1)
      .reduce((a, b) => a + b, 0);

    // Dimension column should expand to cover whole container
    estimateColumnSize[0] = Math.max(
      containerWidth - measureColumnSizeSum,
      estimateColumnSize[0]
    );

    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columns.length,
      getItemKey: (index) => columns[index].name,
      estimateSize: (index) => {
        return estimateColumnSize[index];
      },
      overscan: columnOverscanAmount,
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

  function onSelectItem(event) {
    // store previous scroll position before re-render
    rowScrollOffset = $rowVirtualizer.scrollOffset;
    colScrollOffset = $columnVirtualizer.scrollOffset;
    dispatch("select-item", event.detail);
  }

  async function handleResizeColumn(event) {
    const { size, name } = event.detail;
    manuallyResizedColumns.update((state) => {
      state[name] = Math.max(DimensionTableConfig.minColumnWidth, size);
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

  async function handleColumnHeaderClick(event) {
    colScrollOffset = $columnVirtualizer.scrollOffset;
    dispatch("sort", event.detail);
  }
</script>

<div
  bind:clientWidth={containerWidth}
  style:height="calc(100vh - var(--header, 130px) - 8rem)"
>
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
          {columns}
          noPin={true}
          selectedColumn={sortByColumn}
          on:resize-column={handleResizeColumn}
          on:reset-column-size={handleResetColumnSize}
          on:click-column={handleColumnHeaderClick}
        />
        <!-- VirtualTableBody -->
        <TableCells
          virtualColumnItems={virtualColumns}
          virtualRowItems={virtualRows}
          {rows}
          {columns}
          {activeIndex}
          {selectedIndex}
          {scrolling}
          on:select-item={(event) => onSelectItem(event)}
          on:inspect={setActiveIndex}
        />
      </div>
    {/if}
  </div>
</div>
