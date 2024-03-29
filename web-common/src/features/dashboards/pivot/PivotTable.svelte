<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    ExpandedState,
    SortingState,
    TableOptions,
    Updater,
    createSvelteTable,
    getCoreRowModel,
    getExpandedRowModel,
  } from "@tanstack/svelte-table";
  import {
    createVirtualizer,
    defaultRangeExtractor,
  } from "@tanstack/svelte-virtual";
  import { derived } from "svelte/store";
  import type { PivotDataRow, PivotDataStore } from "./types";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { isTimeDimension } from "./pivot-utils";
  import { getPivotConfig } from "./pivot-data-store";
  import PivotExpandableRow from "./PivotExpandableRow.svelte";

  export let pivotDataStore: PivotDataStore;

  const OVERSCAN = 60;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;
  const MEASURE_PADDING = 16;
  // Distance threshold (in pixels) for triggering data fetch
  const ROW_THRESHOLD = 200;
  const MIN_COL_WIDTH = 150;
  const MAX_COL_WIDTH = 600;
  const MAX_INIT_COL_WIDTH = 400;

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    metricsViewName,
    selectors: {
      pivot: { rows: rowPills },
    },
  } = stateManagers;

  const config = getPivotConfig(stateManagers);

  const pivotDashboardStore = derived(dashboardStore, (dashboard) => {
    return dashboard?.pivot;
  });

  const options = derived(
    [pivotDashboardStore, pivotDataStore],
    ([pivotConfig, pivotData]) => {
      const options: TableOptions<PivotDataRow> = {
        data: pivotData.data,
        columns: pivotData.columnDef,
        state: {
          expanded: pivotConfig.expanded,
          sorting: pivotConfig.sorting,
        },
        onExpandedChange: (e) => handleExpandedChange(e),
        getSubRows: (row) => row.subRows,
        onSortingChange: handleSorting,
        getExpandedRowModel: getExpandedRowModel(),
        getCoreRowModel: getCoreRowModel(),
        enableSortingRemoval: false,
        enableExpanding: true,
      };

      return options;
    },
  );

  const table = createSvelteTable(options);

  let containerRefElement: HTMLDivElement;
  let stickyRows = [0];
  let rowScrollOffset = 0;

  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;

  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];

  $: rowPage = $dashboardStore?.pivot?.rowPage;
  $: headerGroups = $table.getHeaderGroups();
  $: measureCount = $dashboardStore.pivot?.columns?.measure?.length ?? 0;
  $: rows = $table.getRowModel().rows;
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;
  $: headers = headerGroups[0].headers;
  $: firstColumnName = hasDimension
    ? String(headers[0]?.column.columnDef.header)
    : null;

  $: hasDimension = $rowPills.dimension.length > 0;
  $: timeDimension = $config.time.timeDimension;
  $: calculatedFirstColumnWidth =
    hasDimension && firstColumnName
      ? calculateFirstColumnWidth(firstColumnName)
      : 0;

  $: firstColumnWidth = calculatedFirstColumnWidth;
  $: measureGroups = headerGroups[headerGroups.length - 2]?.headers?.slice(
    1,
  ) ?? [null];
  $: measureGroupsLength = measureGroups.length;

  $: measureLengths =
    $dashboardStore.pivot?.columns?.measure?.map((measure) => {
      return measure.title.length * 7 + MEASURE_PADDING;
    }) ?? [];

  $: totalMeasureWidth = measureLengths.reduce((acc, val) => acc + val, 0);
  $: totalLength = measureGroupsLength * totalMeasureWidth;

  $: totalsRow = rows[0];

  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
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

  $: virtualRows = $virtualizer.getVirtualItems();
  $: totalRowSize = $virtualizer.getTotalSize();

  $: rowScrollOffset = $virtualizer?.scrollOffset || 0;

  // In this virtualization model, we create buffer rows before and after our real data
  // This maintains the "correct" scroll position when the user scrolls
  $: [before, after] = virtualRows.length
    ? [
        (virtualRows[1]?.start ?? virtualRows[0].start) - ROW_HEIGHT,
        totalRowSize - virtualRows[virtualRows.length - 1].end,
      ]
    : [0, 0];

  function handleExpandedChange(updater: Updater<ExpandedState>) {
    // Something is off with tanstack's types
    //@ts-expect-error-free
    expanded = updater(expanded);
    metricsExplorerStore.setPivotExpanded($metricsViewName, expanded);
  }

  function handleSorting(updater: Updater<SortingState>) {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }
    metricsExplorerStore.setPivotSort($metricsViewName, sorting);
  }

  function handleScroll(containerRefElement?: HTMLDivElement | null) {
    if (containerRefElement) {
      const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
      const bottomEndDistance = scrollHeight - scrollTop - clientHeight;

      // Fetch more data when scrolling near the bottom end
      if (
        bottomEndDistance < ROW_THRESHOLD &&
        !$pivotDataStore.isFetching &&
        !reachedEndForRows
      ) {
        metricsExplorerStore.setPivotRowPage($metricsViewName, rowPage + 1);
      }
    }
  }

  function calculateFirstColumnWidth(firstColumnName: string) {
    const rows = $pivotDataStore.data;

    // Dates are displayed as shorter values
    if (isTimeDimension(firstColumnName, timeDimension)) return MIN_COL_WIDTH;

    const samples = extractSamples(
      rows.map((row) => row[firstColumnName]),
    ).filter((v): v is string => typeof v === "string");

    const maxValueLength = samples.reduce((max, value) => {
      return Math.max(max, value.length);
    }, 0);

    const finalBasis = Math.max(firstColumnName.length, maxValueLength);
    const pixelLength = finalBasis * 7;
    const final = Math.max(
      MIN_COL_WIDTH,
      Math.min(MAX_INIT_COL_WIDTH, pixelLength + 16),
    );

    return final;
  }

  function extractSamples<T>(arr: T[], sampleSize: number = 30) {
    if (arr.length <= sampleSize) {
      return arr.slice();
    }

    const sectionSize = Math.floor(sampleSize / 3);

    const lastSectionSize = sampleSize - sectionSize * 2;

    const first = arr.slice(0, sectionSize);

    let middleStartIndex = Math.floor((arr.length - sectionSize) / 2);
    const middle = arr.slice(middleStartIndex, middleStartIndex + sectionSize);

    const last = arr.slice(-lastSectionSize);

    return [...first, ...middle, ...last];
  }

  // $: console.log(rows[0].getVisibleCells());
</script>

<div
  class="table-wrapper"
  class:with-row-dimension={hasDimension}
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  style:--total-header-height="{totalHeaderHeight + headerGroups.length}px"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
>
  <table style:width="{totalLength}px">
    {#if firstColumnName && firstColumnWidth}
      <colgroup>
        <col
          style:width="{firstColumnWidth}px"
          style:max-width="{firstColumnWidth}px"
        />
      </colgroup>
    {/if}

    {#each measureGroups as _}
      <colgroup>
        {#each measureLengths as length}
          <col style:width="{length}px" style:max-width="{length}px" />
        {/each}
      </colgroup>
    {/each}

    <thead>
      {#each headerGroups as headerGroup (headerGroup.id)}
        <tr>
          {#each headerGroup.headers as header, i (header.id)}
            {@const sortDirection = header.column.getIsSorted()}
            {@const canResize = i === 0 && hasDimension}
            <th colSpan={header.colSpan}>
              {#if canResize}
                <Resizer
                  min={MIN_COL_WIDTH}
                  max={MAX_COL_WIDTH}
                  basis={MIN_COL_WIDTH}
                  bind:dimension={firstColumnWidth}
                  side="right"
                  direction="EW"
                />
              {/if}

              <button
                class="header-cell"
                class:cursor-pointer={header.column.getCanSort()}
                class:select-none={header.column.getCanSort()}
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  {header.column.columnDef.header}
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

      <PivotExpandableRow
        row={totalsRow}
        {measureCount}
        {hasDimension}
        totals
      />
    </thead>
    <tbody>
      <tr style:height="{before}px" />

      {#each virtualRows.slice(1) as virtualRow (rows[virtualRow.index].id)}
        {@const row = rows[virtualRow.index]}
        <PivotExpandableRow {row} {measureCount} {hasDimension} />
      {/each}

      <tr style:height="{after}px" />
    </tbody>
  </table>
</div>

<style lang="postcss">
  * {
    @apply border-slate-200;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal select-none;
    @apply bg-white table-fixed;
  }

  .table-wrapper {
    @apply overflow-auto h-fit max-h-full w-fit max-w-full;
    @apply border rounded-md z-40;
  }

  /* Pin header */
  thead {
    @apply sticky top-0;
    @apply z-30;
  }

  thead tr {
    height: var(--header-height);
  }

  tbody tr {
    height: var(--row-height);
  }

  .header-cell {
    @apply px-2 bg-white size-full;
    @apply flex items-center gap-x-1 w-full truncate;
    height: var(--header-height);
  }

  th {
    @apply p-0 m-0 text-xs;
  }

  th {
    @apply border-r border-b relative;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  .with-row-dimension thead > tr:not(:nth-last-of-type(2)) > th:first-of-type {
    @apply border-b-0;
  }

  thead > tr:last-of-type > th {
    @apply text-right;
  }

  .with-row-dimension tr > th:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  th:last-of-type {
    @apply border-r-0;
  }
</style>
