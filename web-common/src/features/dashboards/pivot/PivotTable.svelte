<script lang="ts" context="module">
  import { writable } from "svelte/store";
  const measureLengths = writable(new Map<string, number>());
</script>

<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import VirtualTooltip from "@rilldata/web-common/components/virtualized-table/VirtualTooltip.svelte";
  import { getMeasureColumnProps } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import {
    calculateFirstColumnWidth,
    calculateMeasureWidth,
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import {
    NUM_COLUMNS_PER_PAGE,
    NUM_ROWS_PER_PAGE,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { dev } from "$app/environment";
  import {
    type Cell,
    type ExpandedState,
    type SortingState,
    type TableOptions,
    type Updater,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
  } from "@tanstack/svelte-table";
  import {
    createVirtualizer,
    defaultRangeExtractor,
  } from "@tanstack/svelte-virtual";
  import { onMount } from "svelte";
  import type { Readable } from "svelte/motion";
  import { derived } from "svelte/store";
  import { getPivotConfig } from "./pivot-data-config";
  import type { PivotDataRow, PivotDataStore } from "./types";
  import { slugify } from "@rilldata/web-common/lib/string-utils";

  // Distance threshold (in pixels) for triggering data fetch
  const ROW_THRESHOLD = 200;
  const OVERSCAN = 60;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;

  export let pivotDataStore: PivotDataStore;

  const stateManagers = getStateManagers();

  const { dashboardStore, exploreName } = stateManagers;

  const config = getPivotConfig(stateManagers);

  const pivotDashboardStore = derived(dashboardStore, (dashboard) => {
    return dashboard?.pivot;
  });

  const { cloudDataViewer, readOnly } = featureFlags;
  $: isRillDeveloper = $readOnly === false;
  $: canShowDataViewer = Boolean($cloudDataViewer || isRillDeveloper);

  const options: Readable<TableOptions<PivotDataRow>> = derived(
    [pivotDashboardStore, pivotDataStore],
    ([pivotConfig, pivotData]) => {
      let tableData = [...pivotData.data];
      if (pivotData.totalsRowData) {
        tableData = [pivotData.totalsRowData, ...pivotData.data];
      }
      return {
        data: tableData,
        columns: pivotData.columnDef,
        state: {
          expanded: pivotConfig.expanded,
          sorting: pivotConfig.sorting,
        },
        onExpandedChange,
        getSubRows: (row) => row.subRows,
        onSortingChange,
        getExpandedRowModel: getExpandedRowModel(),
        getCoreRowModel: getCoreRowModel(),
        enableSortingRemoval: false,
        enableExpanding: true,
      };
    },
  );

  const table = createSvelteTable(options);

  let containerRefElement: HTMLDivElement;
  let stickyRows = [0];
  let rowScrollOffset = 0;
  let scrollLeft = 0;
  let initialMeasureIndexOnResize = 0;
  let initLengthOnResize = 0;
  let initScrollOnResize = 0;
  let percentOfChangeDuringResize = 0;
  let resizingMeasure = false;
  let resizing = false;

  $: ({
    expanded,
    sorting,
    columnPage,
    rowPage,
    rows: rowPills,
    columns: columnPills,
    activeCell,
  } = $pivotDashboardStore);

  $: timeDimension = $config.time.timeDimension;
  $: hasDimension = rowPills.dimension.length > 0;
  $: hasColumnDimension = columnPills.dimension.length > 0;
  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;
  $: assembled = $pivotDataStore.assembled;
  $: dataRows = $pivotDataStore.data;
  $: totalsRow = $pivotDataStore.totalsRowData;

  $: measures = getMeasureColumnProps($config);
  $: measureCount = measures.length;
  $: measures.forEach(({ name, label, formatter }) => {
    if (!$measureLengths.has(name)) {
      const estimatedWidth = calculateMeasureWidth(
        name,
        label,
        formatter,
        totalsRow,
        dataRows,
      );
      measureLengths.update((measureLengths) => {
        return measureLengths.set(name, estimatedWidth);
      });
    }
  });

  $: subHeaders = [
    {
      subHeaders: measures.map((m) => ({
        column: { columnDef: { name: m.name } },
      })),
    },
  ];

  let measureGroups: {
    subHeaders: { column: { columnDef: { name: string } } }[];
  }[];
  // @ts-expect-error - I have manually added the name property in pivot-column-definition.ts
  $: measureGroups =
    headerGroups[headerGroups.length - 2]?.headers?.slice(
      hasDimension ? 1 : 0,
    ) ?? subHeaders;
  // $: console.log("measureGroups: ", measureGroups);
  $: measureGroupsLength = measureGroups.length;
  $: totalMeasureWidth = measures.reduce(
    (acc, { name }) => acc + ($measureLengths.get(name) ?? 0),
    0,
  );
  $: totalLength = measureGroupsLength * totalMeasureWidth;
  // $: console.log("totalLength: ", totalLength);

  $: headerGroups = $table.getHeaderGroups();
  // $: console.log("headerGroups: ", headerGroups);
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;
  $: headers = headerGroups[0].headers;
  $: firstColumnName = hasDimension
    ? String(headers[0]?.column.columnDef.header)
    : null;
  $: firstColumnWidth =
    hasDimension && firstColumnName
      ? calculateFirstColumnWidth(firstColumnName, timeDimension, dataRows)
      : 0;
  // $: console.log("firstColumnWidth: ", firstColumnWidth);

  $: rows = $table.getRowModel().rows;
  $: rowVirtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
    count: rows.length,
    getScrollElement: () => containerRefElement,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
    initialOffset: rowScrollOffset,
    rangeExtractor: (range) => {
      const next = new Set([...stickyRows, ...defaultRangeExtractor(range)]);
      return [...next].sort((a, b) => a - b);
    },
  });

  $: columns = $table.getVisibleLeafColumns();
  $: columnVirtualizer = createVirtualizer<
    HTMLDivElement,
    HTMLTableCellElement
  >({
    horizontal: true,
    count: columns.length,
    getScrollElement: () => containerRefElement,
    estimateSize: (index) => columns[index].getSize(),
    overscan: OVERSCAN,
    rangeExtractor: (range) => {
      const next = new Set([...defaultRangeExtractor(range)]);
      return [...next].sort((a, b) => a - b);
    },
  });

  $: virtualRows = $rowVirtualizer.getVirtualItems();
  $: totalRowSize = $rowVirtualizer.getTotalSize();

  $: virtualColumns = $columnVirtualizer.getVirtualItems();
  // $: totalColumnSize = $columnVirtualizer.getTotalSize();

  let virtualPaddingLeft: number | undefined;
  let virtualPaddingRight: number | undefined;

  $: if (columnVirtualizer && virtualColumns?.length) {
    virtualPaddingLeft = virtualColumns[0]?.start ?? 0;
    virtualPaddingRight =
      $columnVirtualizer.getTotalSize() -
      (virtualColumns[virtualColumns.length - 1]?.end ?? 0);
  }

  $: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0;

  // See: https://github.com/TanStack/virtual/issues/585#issuecomment-1716247313
  // In this virtualization model, we create buffer rows before and after our real data
  // This maintains the "correct" scroll position when the user scrolls
  $: [before, after] = virtualRows.length
    ? [
        (virtualRows[1]?.start ?? virtualRows[0].start) - ROW_HEIGHT,
        totalRowSize - virtualRows[virtualRows.length - 1].end,
      ]
    : [0, 0];

  $: if (resizingMeasure && containerRefElement && measureLengths) {
    containerRefElement.scrollTo({
      left:
        initScrollOnResize +
        percentOfChangeDuringResize * (totalLength - initLengthOnResize),
    });
  }

  let customShortcuts: { description: string; shortcut: string }[] = [];
  $: if (canShowDataViewer) {
    customShortcuts = [
      { description: "View raw data for aggregated cell", shortcut: "Click" },
    ];
  }
  function onExpandedChange(updater: Updater<ExpandedState>) {
    if (updater instanceof Function) {
      expanded = updater(expanded);
    } else {
      expanded = updater;
    }
    metricsExplorerStore.setPivotExpanded($exploreName, expanded);
  }

  function onSortingChange(updater: Updater<SortingState>) {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }
    metricsExplorerStore.setPivotSort($exploreName, sorting);
    rowScrollOffset = 0;
  }

  function handleScroll(containerRefElement?: HTMLDivElement | null) {
    if (containerRefElement) {
      if (hovering) hovering = null;
      const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
      const bottomEndDistance = scrollHeight - scrollTop - clientHeight;
      scrollLeft = containerRefElement.scrollLeft;

      const isReachingPageEnd = bottomEndDistance < ROW_THRESHOLD;
      const canFetchMoreData =
        !$pivotDataStore.isFetching && !reachedEndForRows;
      const hasMoreRowsDataThanOnePage = rows.length >= NUM_ROWS_PER_PAGE;
      if (isReachingPageEnd && hasMoreRowsDataThanOnePage && canFetchMoreData) {
        console.log("fetching more rowPage: ", rowPage);
        metricsExplorerStore.setPivotRowPage($exploreName, rowPage + 1);
      }

      // FIXME: when uncommented, the row page will increase as we scroll right
      const rightEndDistance =
        containerRefElement.scrollWidth -
        scrollLeft -
        containerRefElement.clientWidth;
      const isReachingColumnEnd = rightEndDistance < ROW_THRESHOLD;
      const hasMoreColumnsThanOnePage = columns.length >= NUM_COLUMNS_PER_PAGE;
      if (isReachingColumnEnd && hasMoreColumnsThanOnePage) {
        console.log("fetching more columns [columnPage]: ", columnPage);
        // metricsExplorerStore.setPivotColumnPage($exploreName, columnPage + 1);
      }
    }
  }

  function onResizeStart(e: MouseEvent) {
    initLengthOnResize = totalLength;
    initScrollOnResize = scrollLeft;

    const offset =
      e.clientX -
      containerRefElement.getBoundingClientRect().left -
      firstColumnWidth -
      measures.reduce((rollingSum, { name }, i) => {
        return i <= initialMeasureIndexOnResize
          ? rollingSum + ($measureLengths.get(name) ?? 0)
          : rollingSum;
      }, 0) +
      4;

    percentOfChangeDuringResize = (scrollLeft + offset) / totalLength;
  }

  let showTooltip = false;
  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;
  let timer: ReturnType<typeof setTimeout>;

  type HoveringData = {
    value: string | number | null;
  };

  function handleCellClick(cell: Cell<PivotDataRow, unknown>) {
    if (!canShowDataViewer) return;
    const rowId = cell.row.id;
    const columnId = cell.column.id;

    metricsExplorerStore.setPivotActiveCell($exploreName, rowId, columnId);
  }

  function handleHover(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLElement;
    },
  ) {
    hoverPosition = e.currentTarget.getBoundingClientRect();

    const value = e.currentTarget.dataset.value;

    if (value === undefined) return;

    hovering = {
      value,
    };

    timer = setTimeout(() => {
      showTooltip = true;
    }, 250);
  }

  function handleLeave() {
    clearTimeout(timer);
    showTooltip = false;
    hovering = null;
  }

  function handleClick(e: MouseEvent) {
    if (!isElement(e.target)) return;

    const value = e.target.dataset.value;

    if (value === undefined) return;

    copyToClipboard(value);
  }

  function isElement(target: EventTarget | null): target is HTMLElement {
    return target instanceof HTMLElement;
  }

  function isMeasureColumn(header, colNumber: number) {
    // Measure columns are the last columns in the header group
    if (header.depth !== headerGroups.length) return;
    // If there is a row dimension, the first column is not a measure column
    if (!hasDimension) {
      return true;
    } else return colNumber > 0;
  }

  function isCellActive(cell: Cell<PivotDataRow, unknown>) {
    return (
      cell.row.id === activeCell?.rowId &&
      cell.column.id === activeCell?.columnId
    );
  }

  onMount(() => {
    // wait for layout to be calculated
    requestAnimationFrame(() => {
      handleScroll(containerRefElement);
    });
  });

  $: tableWidth = totalLength + firstColumnWidth;
  $: tableHeight = totalRowSize + totalHeaderHeight + headerGroups.length;
</script>

<!-- FIXME: columns and columnPage should be increasing as we scroll right -->
<!-- DEBUG ONLY -->
{#if dev}
  <span
    >({columns.length} columns) ({columnPage} Column Page) ({tableWidth}px Table
    Width)</span
  >
  <span
    >({rows.length} rows) ({rowPage} Row Page) ({tableHeight}px Table Height)</span
  >
{/if}

<div
  class="table-wrapper relative"
  class:with-row-dimension={hasDimension}
  class:with-col-dimension={hasColumnDimension}
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  style:--total-header-height="{totalHeaderHeight + headerGroups.length}px"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
  class:pointer-events-none={resizing}
>
  <div
    class="w-full absolute top-0 z-50 flex pointer-events-none"
    style:width="{tableWidth}px"
    style:height="{tableHeight}px"
  >
    <div
      style:width="{firstColumnWidth}px"
      class="sticky left-0 flex-none flex"
    >
      <Resizer
        side="right"
        direction="EW"
        min={WIDTHS.MIN_COL_WIDTH}
        max={WIDTHS.MAX_COL_WIDTH}
        dimension={firstColumnWidth}
        onUpdate={(d) => (firstColumnWidth = d)}
        onMouseDown={(e) => {
          resizingMeasure = false;
          resizing = true;
          onResizeStart(e);
        }}
        onMouseUp={() => {
          resizing = false;
          resizingMeasure = false;
        }}
      >
        <div class="resize-bar" />
      </Resizer>
    </div>

    {#each measureGroups as { subHeaders }, groupIndex (groupIndex)}
      <div class="h-full z-50 flex" style:width="{totalMeasureWidth}px">
        {#each subHeaders as { column: { columnDef: { name } } }, i (name)}
          {@const length =
            $measureLengths.get(name) ?? WIDTHS.INIT_MEASURE_WIDTH}
          {@const last =
            i === subHeaders.length - 1 &&
            groupIndex === measureGroups.length - 1}
          <div style:width="{length}px" class="h-full relative">
            <Resizer
              side="right"
              direction="EW"
              min={WIDTHS.MIN_MEASURE_WIDTH}
              max={WIDTHS.MAX_MEASURE_WIDTH}
              dimension={length}
              justify={last ? "end" : "center"}
              hang={!last}
              onUpdate={(d) => {
                measureLengths.update((measureLengths) => {
                  return measureLengths.set(name, d);
                });
              }}
              onMouseDown={(e) => {
                resizingMeasure = true;
                resizing = true;
                initialMeasureIndexOnResize = i;
                onResizeStart(e);
              }}
              onMouseUp={() => {
                resizing = false;
                resizingMeasure = false;
              }}
            >
              <div class="resize-bar" />
            </Resizer>
          </div>
        {/each}
      </div>
    {/each}
  </div>

  <table
    role="presentation"
    style:width="{tableWidth}px"
    style:height="{tableHeight}px"
    on:click={modified({ shift: handleClick })}
  >
    <colgroup>
      {#if firstColumnName && firstColumnWidth}
        <col
          style:width="{firstColumnWidth}px"
          style:max-width="{firstColumnWidth}px"
        />
      {/if}

      {#each measureGroups as { subHeaders }, i (i)}
        {#each subHeaders as { column: { columnDef: { name } } } (name)}
          {@const length =
            $measureLengths.get(name) ?? WIDTHS.INIT_MEASURE_WIDTH}
          <col style:width="{length}px" style:max-width="{length}px" />
        {/each}
      {/each}
    </colgroup>

    <thead>
      {#each headerGroups as headerGroup (headerGroup.id)}
        <tr>
          {#each headerGroup.headers as header, i (header.id)}
            {@const sortDirection = header.column.getIsSorted()}
            <th
              colSpan={header.colSpan}
              data-id={slugify(header.id)}
              data-index={header.index}
            >
              <button
                class="header-cell"
                class:cursor-pointer={header.column.getCanSort()}
                class:select-none={header.column.getCanSort()}
                class:flex-row-reverse={isMeasureColumn(header, i)}
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  <p class="truncate">
                    {header.column.columnDef.header}
                  </p>
                  {#if sortDirection}
                    <span
                      class="transition-transform -mr-1"
                      class:-rotate-180={sortDirection === "asc"}
                    >
                      <ArrowDown />
                    </span>
                  {/if}
                {/if}
              </button>
            </th>
          {/each}
        </tr>
      {/each}
    </thead>
    <tbody>
      <tr style:height="{before}px" />
      {#each virtualRows as row (row.index)}
        {@const cells = rows[row.index].getVisibleCells()}
        <tr>
          {#each cells as cell, i (cell.id)}
            {@const result =
              typeof cell.column.columnDef.cell === "function"
                ? cell.column.columnDef.cell(cell.getContext())
                : cell.column.columnDef.cell}
            {@const isActive = isCellActive(cell)}
            <td
              class="ui-copy-number"
              class:active-cell={isActive}
              class:interactive-cell={canShowDataViewer}
              class:border-r={i % measureCount === 0 && i}
              on:click={() => handleCellClick(cell)}
              on:mouseenter={handleHover}
              on:mouseleave={handleLeave}
              data-value={cell.getValue()}
              data-id={slugify(cell.id)}
              class:totals-column={i > 0 && i <= measureCount}
            >
              <div
                class="cell pointer-events-none truncate"
                role="presentation"
              >
                {#if result?.component && result?.props}
                  <svelte:component
                    this={result.component}
                    {...result.props}
                    {assembled}
                  />
                {:else if typeof result === "string" || typeof result === "number"}
                  {result}
                {:else}
                  <svelte:component
                    this={flexRender(
                      cell.column.columnDef.cell,
                      cell.getContext(),
                    )}
                  />
                {/if}
              </div>
            </td>
          {/each}
        </tr>
      {/each}
      <tr style:height="{after}px" />
    </tbody>
  </table>
</div>

{#if showTooltip && hovering}
  <VirtualTooltip
    sortable={true}
    {hovering}
    {hoverPosition}
    pinned={false}
    {customShortcuts}
  />
{/if}

<style lang="postcss">
  * {
    @apply border-slate-200;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal;
    @apply bg-white table-fixed;
  }

  .table-wrapper {
    @apply overflow-auto h-fit max-h-full w-fit max-w-full;
    @apply border rounded-md z-40;
  }

  /* Pin header */
  thead {
    @apply sticky top-0;
    @apply z-30 bg-white;
  }

  tbody .cell {
    height: var(--row-height);
  }

  th {
    @apply p-0 m-0 text-xs;
    @apply border-r border-b relative;
  }

  th:last-of-type,
  td:last-of-type {
    @apply border-r-0;
  }

  th,
  td {
    @apply whitespace-nowrap text-xs;
  }

  td {
    @apply text-right;
    @apply p-0 m-0;
  }

  .header-cell {
    @apply px-2 bg-white size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    @apply font-medium;
    height: var(--header-height);
  }

  .cell {
    @apply size-full p-1 px-2;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  .with-row-dimension thead > tr:not(:last-of-type) > th:first-of-type {
    @apply border-b-0;
  }

  .with-row-dimension tr > th:first-of-type {
    @apply sticky left-0 z-20;
    @apply bg-white;
  }

  .with-row-dimension tr > td:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  tr > td:first-of-type:not(:last-of-type) {
    @apply border-r font-normal;
  }

  /* The totals row */
  tbody > tr:nth-of-type(2) {
    @apply bg-slate-50 sticky z-20 font-semibold;
    top: var(--total-header-height);
  }

  /* The totals row header */
  tbody > tr:nth-of-type(2) > td:first-of-type {
    @apply font-semibold;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }

  tr:hover .active-cell .cell {
    @apply bg-primary-100;
  }

  .totals-column {
    @apply bg-slate-50;
  }
  .with-col-dimension .totals-column {
    @apply font-semibold;
  }
  .interactive-cell {
    @apply cursor-pointer;
  }
  .interactive-cell:hover .cell {
    @apply bg-primary-100;
  }
  .active-cell .cell {
    @apply bg-primary-50;
  }

  .resize-bar {
    @apply bg-primary-500 w-1 h-full;
  }
</style>
