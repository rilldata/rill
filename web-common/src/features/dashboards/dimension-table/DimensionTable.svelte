<!-- @component 
Creates a virtualized dimension table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
TableCells – the cell contents.
-->
<script lang="ts">
  import ColumnHeaders from "@rilldata/web-common/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "@rilldata/web-common/components/virtualized-table/sections/TableCells.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { createEventDispatcher, setContext } from "svelte";
  import DimensionFilterGutter from "./DimensionFilterGutter.svelte";
  import { DimensionTableConfig } from "./DimensionTableConfig";
  import DimensionValueHeader from "./DimensionValueHeader.svelte";

  const dispatch = createEventDispatcher();

  export let rows;
  export let columns: VirtualizedTableColumns[];
  export let selectedValues: Array<unknown> = [];
  export let sortByColumn: string;
  export let dimensionName: string;
  export let excludeMode = false;

  /** the overscan values tell us how much to render off-screen. These may be set by the consumer
   * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
   * to render on initial load.
   */
  export let rowOverscanAmount = 40;
  export let columnOverscanAmount = 5;

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
  const CHARACTER_LIMIT_FOR_WRAPPING = 9;
  const FILTER_COLUMN_WIDTH = DimensionTableConfig.indexWidth;

  $: selectedIndex = selectedValues
    .map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    })
    .filter((i) => i >= 0);

  $: rowScrollOffset = 0;
  $: colScrollOffset = 0;

  /** if we're inferring the column widths from static-ish data, let's
   * find the largest strings in the column and use that to bootstrap the
   * column widths.
   */
  let columnWidths: { [key: string]: number } = {};
  let largestColumnLength = 0;
  columns.forEach((column, i) => {
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
    if (i != 0) {
      largestColumnLength = Math.max(
        largestColumnLength,
        column.label?.length || column.name.length
      );
    }
  });

  /* check if column header requires extra space for larger column names  */
  const config = DimensionTableConfig;
  if (largestColumnLength > CHARACTER_LIMIT_FOR_WRAPPING) {
    config.columnHeaderHeight = 46;
  }

  /* set context for child components */
  setContext("config", config);

  let estimateColumnSize;
  let measureColumns = [];
  let dimensionColumn;
  let horizontalScrolling = false;

  $: if (rows && columns) {
    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => config.rowHeight,
      overscan: rowOverscanAmount,
      paddingStart: config.columnHeaderHeight,
      initialOffset: rowScrollOffset,
    });

    estimateColumnSize = columns.map((column, i) => {
      if (column.name.includes("delta")) return config.comparisonColumnWidth;
      if (i != 0) return config.defaultColumnWidth;

      const largestStringLength =
        columnWidths[column.name] * CHARACTER_WIDTH + CHARACTER_X_PAD;

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
          ? config.minHeaderWidthWhenColumsAreSmall
          : headerWidth;

      return largestStringLength
        ? Math.min(
            config.maxColumnWidth,
            Math.max(
              largestStringLength,
              effectiveHeaderWidth,
              /** All columns must be minColumnWidth regardless of user settings. */
              config.minColumnWidth
            )
          )
        : /** if there isn't a longet string length for some reason, let's go with a
           * default column width. We should not be in this state.
           */
          config.defaultColumnWidth;
    });

    const measureColumnSizeSum = estimateColumnSize
      .slice(1)
      .reduce((a, b) => a + b, 0);

    /* Dimension column should expand to cover whole container */
    estimateColumnSize[0] = Math.max(
      containerWidth - measureColumnSizeSum - FILTER_COLUMN_WIDTH,
      estimateColumnSize[0]
    );

    /* Separate out dimension column */
    dimensionColumn = columns.find((c) => c.name == dimensionName);
    measureColumns = columns.filter((c) => c.name !== dimensionName);

    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: measureColumns.length,
      getItemKey: (index) => measureColumns[index].name,
      estimateSize: (index) => {
        return estimateColumnSize[index + 1];
      },
      overscan: columnOverscanAmount,
      paddingStart: estimateColumnSize[0] + FILTER_COLUMN_WIDTH,
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
    on:scroll={() => {
      horizontalScrolling = container?.scrollLeft > 0;
    }}
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
          noPin={true}
          selectedColumn={sortByColumn}
          columns={measureColumns}
          on:click-column={handleColumnHeaderClick}
        />

        <div class="flex">
          <!-- Gutter for Include Exlude Filter -->
          <DimensionFilterGutter
            virtualRowItems={virtualRows}
            totalHeight={virtualHeight}
            {selectedIndex}
            {excludeMode}
            on:select-item={(event) => onSelectItem(event)}
          />
          <DimensionValueHeader
            virtualRowItems={virtualRows}
            totalHeight={virtualHeight}
            width={estimateColumnSize[0]}
            column={dimensionColumn}
            {rows}
            {activeIndex}
            {selectedIndex}
            {excludeMode}
            {scrolling}
            {horizontalScrolling}
            on:select-item={(event) => onSelectItem(event)}
            on:inspect={setActiveIndex}
          />
        </div>
        {#if rows.length}
          <!-- VirtualTableBody -->
          <TableCells
            virtualColumnItems={virtualColumns}
            virtualRowItems={virtualRows}
            columns={measureColumns}
            selectedColumn={sortByColumn}
            {rows}
            {activeIndex}
            {selectedIndex}
            {scrolling}
            {excludeMode}
            on:select-item={(event) => onSelectItem(event)}
            on:inspect={setActiveIndex}
          />
        {:else}
          <div class="flex text-gray-500 justify-center mt-[30vh]">
            No results to show
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
