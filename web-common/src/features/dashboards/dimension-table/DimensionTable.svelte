<!-- @component
Creates a virtualized dimension table. This consists of four sub-components:
ColumnHeaders – sticky column headers. Utilizes the columnVirtualizer (for now).
TableCells – the cell contents.
-->
<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ColumnHeaders from "@rilldata/web-common/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "@rilldata/web-common/components/virtualized-table/sections/TableCells.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { setContext } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { DimensionTableRow } from "./dimension-table-types";
  import {
    estimateColumnCharacterWidths,
    estimateColumnSizes,
  } from "./dimension-table-utils";
  import DimensionFilterGutter from "./DimensionFilterGutter.svelte";
  import { DIMENSION_TABLE_CONFIG as config } from "./DimensionTableConfig";
  import DimensionValueHeader from "./DimensionValueHeader.svelte";

  let {
    rows,
    columns,
    selectedValues,
    dimensionName,
    isFetching,
    rowOverscanAmount = 120,
    columnOverscanAmount = 12,
    onSelectItem,
  }: {
    rows: DimensionTableRow[];
    columns: VirtualizedTableColumns[];
    selectedValues: ReturnType<typeof selectedDimensionValues>;
    dimensionName: string;
    isFetching: boolean;

    /** the overscan values tell us how much to render off-screen. These may be set by the consumer
     * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
     * to render on initial load.
     */
    rowOverscanAmount?: number;
    columnOverscanAmount?: number;

    onSelectItem: (data: { index: number; meta: boolean }) => void;
  } = $props();

  const {
    actions: {
      dimensionTable,
      comparison: { toggleComparisonDimension },
    },
    selectors: {
      sorting: { sortByMeasure },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    expressionFilterManager,
  } = getStateManagers();

  let excludeMode = $derived(
    expressionFilterManager.filterManagers.dimensions.find(
      (dfm) => dfm.name === dimensionName,
    )?.exclude ?? false,
  );
  let isBeingCompared = $derived($isBeingComparedReadable(dimensionName));

  let container: HTMLDivElement;

  let containerWidth: number = $state(0);

  /** this is a perceived character width value, in pixels, when our monospace
   * font is 12px high. */
  const CHARACTER_LIMIT_FOR_WRAPPING = 9;
  const FILTER_COLUMN_WIDTH = config.indexWidth;

  let selectedIndex = $derived(
    $selectedValues.data?.map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    }) ?? [],
  );

  let rowScrollOffset = $state(0);
  let colScrollOffset = $state(0);

  /** if we're inferring the column widths from static-ish data, let's
   * find the largest strings in the column and use that to bootstrap the
   * column widths.
   */
  // svelte-ignore state_referenced_locally
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

  let estimateColumnSize: number[] = $state([]);

  /* Separate out dimension column */
  let dimensionColumn = $derived(
    columns?.find((c) => c.name == dimensionName) as VirtualizedTableColumns,
  );
  let measureColumns = $derived(
    columns?.filter((c) => c.name !== dimensionName) ?? [],
  );

  let horizontalScrolling = $state(false);

  let manualDimensionColumnWidth: number | null = $state(null);

  let rowVirtualizer = $derived(
    createVirtualizer({
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
    }),
  );
  $effect(() => {
    rowScrollOffset = $rowVirtualizer?.scrollOffset || 0;
  });
  $effect(() => {
    colScrollOffset = $columnVirtualizer?.scrollOffset || 0;
  });

  $effect(() => {
    if (rows && columns) {
      estimateColumnSize = estimateColumnSizes(
        columns,
        columnWidths,
        rows,
        config,
      );

      if (manualDimensionColumnWidth !== null) {
        estimateColumnSize[0] = manualDimensionColumnWidth;
      }
    }
  });

  let columnVirtualizer = $derived(
    createVirtualizer({
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
    }),
  );

  let virtualRows = $derived($rowVirtualizer?.getVirtualItems() ?? []);
  let virtualHeight = $derived($rowVirtualizer?.getTotalSize() ?? 0);

  let virtualColumns = $derived($columnVirtualizer?.getVirtualItems() ?? []);
  let virtualWidth = $derived($columnVirtualizer?.getTotalSize() ?? 0);

  let activeIndex: number = $state(-1);
  function setActiveIndex(index: number) {
    activeIndex = index;
  }
  function clearActiveIndex() {
    activeIndex = -1;
  }

  /** handle scrolling tooltip suppression */
  let scrolling = $state(false);
  let timeoutID: ReturnType<typeof setTimeout> = $state(0 as any);
  $effect(() => {
    if (scrolling) {
      if (timeoutID) clearTimeout(timeoutID);
      timeoutID = setTimeout(() => {
        scrolling = false;
      }, 200);
    }
  });

  function onSelectItemHandler(data: { index: number; meta: boolean }) {
    // store previous scroll position before re-render
    rowScrollOffset = $rowVirtualizer?.scrollOffset ?? 0;
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;
    onSelectItem(data);
  }

  function handleColumnHeaderClick(columnName: string) {
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;

    dimensionTable.handleDimensionMeasureColumnHeaderClick(columnName);
  }

  function handleResizeDimensionColumn(size: number) {
    rowScrollOffset = $rowVirtualizer?.scrollOffset ?? 0;
    colScrollOffset = $columnVirtualizer?.scrollOffset ?? 0;

    manualDimensionColumnWidth = Math.max(config.minColumnWidth, size);
  }
</script>

<div
  bind:clientWidth={containerWidth}
  style="height: 100%;"
  role="table"
  class="relative"
  aria-label={m.dashboard_dimension_table_aria()}
>
  <div
    bind:this={container}
    onscroll={() => {
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
        class="relative"
        onmouseleave={clearActiveIndex}
        onblur={clearActiveIndex}
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
          onClickColumn={handleColumnHeaderClick}
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
          />
          {#if dimensionColumn}
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
              onSelectItem={onSelectItemHandler}
              onInspect={setActiveIndex}
              onResizeColumn={handleResizeDimensionColumn}
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
            onSelectItem={onSelectItemHandler}
            onInspect={setActiveIndex}
            cellLabel={m.dashboard_filter_dimension_value()}
          />
        {/if}
      </div>
    {/if}
  </div>
  {#if !rows.length}
    <!-- The virtualized grid above is sized to its rows and uses paint containment,
      so empty states must render outside it to be visible. -->
    <div
      class="absolute inset-0 flex items-center justify-center pointer-events-none"
      style:padding-top="{config.columnHeaderHeight}px"
    >
      {#if isFetching || $selectedValues.isFetching}
        <DelayedSpinner isLoading size="24px" />
      {:else}
        <span class="text-fg-secondary">
          {m.dashboard_no_results_to_show()}
        </span>
      {/if}
    </div>
  {/if}
</div>
