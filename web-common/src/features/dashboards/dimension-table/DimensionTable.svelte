<!-- @component 
Creates a virtualized dimension table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
TableCells – the cell contents.
-->
<script lang="ts">
  import ColumnHeaders from "@rilldata/web-common/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "@rilldata/web-common/components/virtualized-table/sections/TableCells.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { createEventDispatcher, setContext } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { DimensionTableRow } from "./dimension-table-types";
  import {
    estimateColumnCharacterWidths,
    estimateColumnSizes,
  } from "./dimension-table-utils";
  import DimensionFilterGutter from "./DimensionFilterGutter.svelte";
  import { DIMENSION_TABLE_CONFIG as config } from "./DimensionTableConfig";
  import DimensionValueHeader from "./DimensionValueHeader.svelte";

  const dispatch = createEventDispatcher();

  export let rows: DimensionTableRow[];
  export let columns: VirtualizedTableColumns[];
  export let selectedValues: ReturnType<typeof selectedDimensionValues>;
  export let dimensionName: string;
  export let isFetching: boolean;

  const {
    actions: {
      dimensionTable,
      comparison: { toggleComparisonDimension },
    },
    selectors: {
      sorting: { sortByMeasure },
      dimensionFilters: { isFilterExcludeMode },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
  } = getStateManagers();

  $: excludeMode = $isFilterExcludeMode(dimensionName);
  $: isBeingCompared = $isBeingComparedReadable(dimensionName);

  /** the overscan values tell us how much to render off-screen. These may be set by the consumer
   * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
   * to render on initial load.
   */
  export let rowOverscanAmount = 120;
  export let columnOverscanAmount = 12;

  let container: HTMLDivElement;

  let containerWidth: number;

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_LIMIT_FOR_WRAPPING = 9;
  const FILTER_COLUMN_WIDTH = config.indexWidth;

  $: selectedIndex =
    $selectedValues.data?.map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    }) ?? [];

  let rowScrollOffset = 0;
  $: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0;
  let colScrollOffset = 0;
  $: colScrollOffset = $columnVirtualizer?.scrollOffset || 0;

  /** if we're inferring the column widths from static-ish data, let's
   * find the largest strings in the column and use that to bootstrap the
   * column widths.
   */
  const { columnWidths, largestColumnLength } = estimateColumnCharacterWidths(
    columns,
    rows,
  );

  /* check if column header requires extra space for larger column names  */
  if (largestColumnLength > CHARACTER_LIMIT_FOR_WRAPPING) {
    config.columnHeaderHeight = 46;
  }

  /* set context for child components */
  setContext("config", config);

  let estimateColumnSize: number[] = [];

  /* Separate out dimension column */
  $: dimensionColumn = columns?.find(
    (c) => c.name == dimensionName,
  ) as VirtualizedTableColumns;
  $: measureColumns = columns?.filter((c) => c.name !== dimensionName) ?? [];

  let horizontalScrolling = false;

  let manualDimensionColumnWidth: number | null = null;

  $: rowVirtualizer = createVirtualizer({
    getScrollElement: () => container,
    count: rows.length,
    estimateSize: () => config.rowHeight,
    // Provides a stable identity for each virtualized row so the virtualizer can reuse DOM nodes
    // instead of remounting them during fast scrolls or data updates. This reduces blank frames,
    // preserves focus/hover state, and avoids unnecessary re-renders.
    getItemKey: (index) => String(rows?.[index]?.[dimensionName] ?? index),
    overscan: rowOverscanAmount,
    paddingStart: config.columnHeaderHeight,
    initialOffset: rowScrollOffset,
  });

  $: if (rows && columns) {
    estimateColumnSize = estimateColumnSizes(
      columns,
      columnWidths,
      containerWidth,
      config,
    );

    if (manualDimensionColumnWidth !== null) {
      estimateColumnSize[0] = manualDimensionColumnWidth;
    }
  }

  $: columnVirtualizer = createVirtualizer({
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

  $: virtualRows = $rowVirtualizer?.getVirtualItems() ?? [];
  $: virtualHeight = $rowVirtualizer?.getTotalSize() ?? 0;

  $: virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];
  $: virtualWidth = $columnVirtualizer?.getTotalSize() ?? 0;

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
    rowScrollOffset = $rowVirtualizer?.scrollOffset ?? 0;
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;
    dispatch("select-item", event.detail);
  }

  async function handleColumnHeaderClick(event) {
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;
    const columnName = event.detail;
    dimensionTable.handleDimensionMeasureColumnHeaderClick(columnName);
  }

  async function handleResizeDimensionColumn(event) {
    rowScrollOffset = $rowVirtualizer?.scrollOffset ?? 0;
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;

    const { size } = event.detail;
    manualDimensionColumnWidth = Math.max(config.minColumnWidth, size);
  }
</script>

<div
  bind:clientWidth={containerWidth}
  style="height: 100%;"
  role="table"
  aria-label="Dimension table"
>
  <div
    bind:this={container}
    on:scroll={() => {
      horizontalScrolling = container?.scrollLeft > 0;
      /** capture to suppress cell tooltips. Otherwise,
       * there's quite a bit of rendering jank.
       */
      scrolling = true;
    }}
    style:width="100%"
    style:height="100%"
    class="overflow-auto grid max-w-fit"
    style:grid-template-columns="max-content auto"
  >
    {#if $rowVirtualizer}
      <div
        role="grid"
        tabindex="0"
        class="relative bg-surface"
        on:mouseleave={clearActiveIndex}
        on:blur={clearActiveIndex}
        style:will-change="transform, contents"
        style:contain="content"
        style:width="{virtualWidth}px"
        style:height="{virtualHeight}px"
      >
        <!-- measure column headers -->
        <ColumnHeaders
          virtualColumnItems={virtualColumns}
          noPin={true}
          sortByMeasure={$sortByMeasure}
          columns={measureColumns}
          on:click-column={handleColumnHeaderClick}
        />
        <!-- dimension value and gutter column -->
        <div class="flex">
          <!-- Gutter for Include Exlude Filter -->
          <DimensionFilterGutter
            virtualRowItems={virtualRows}
            totalHeight={virtualHeight}
            {selectedIndex}
            {isBeingCompared}
            {excludeMode}
            {dimensionName}
            {toggleComparisonDimension}
            on:select-item={(event) => onSelectItem(event)}
          />
          {#if dimensionColumn}
            <DimensionValueHeader
              on:resize-column={handleResizeDimensionColumn}
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
              on:dimension-sort
              on:select-item={(event) => onSelectItem(event)}
              on:inspect={setActiveIndex}
            />
          {/if}
        </div>
        {#if rows.length}
          <!-- VirtualTableBody -->
          <TableCells
            virtualColumnItems={virtualColumns}
            virtualRowItems={virtualRows}
            columns={measureColumns}
            {rows}
            {activeIndex}
            {selectedIndex}
            {scrolling}
            {excludeMode}
            on:select-item={(event) => onSelectItem(event)}
            on:inspect={setActiveIndex}
            cellLabel="Filter dimension value"
          />
        {:else if isFetching || $selectedValues.isFetching}
          <div class="flex text-gray-500 justify-center mt-[30vh]">
            Loading...
          </div>
        {:else}
          <div class="flex text-gray-500 justify-center mt-[30vh]">
            No results to show
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
