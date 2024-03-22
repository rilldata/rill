<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    TableOptions,
    createSvelteTable,
    flexRender,
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

  export let pivotDataStore: PivotDataStore;

  const OVERSCAN = 80;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;
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
        onExpandedChange: handleExpandedChange,
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

  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;
  $: assembled = $pivotDataStore.assembled;
  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];

  $: rowPage = $dashboardStore?.pivot?.rowPage;
  $: headerGroups = $table.getHeaderGroups();
  $: measureCount = $dashboardStore.pivot?.columns?.measure?.length ?? 0;
  $: rows = $table.getRowModel().rows;
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;
  $: headers = headerGroups[0].headers;
  $: firstColumnName = headers[0]?.column.columnDef.id;

  $: hasDimension = $rowPills.dimension.length > 0;
  $: timeDimension = $config.time.timeDimension;
  $: calculatedFirstColumnWidth =
    hasDimension && firstColumnName
      ? calculateFirstColumnWidth(firstColumnName)
      : 0;

  $: firstColumnWidth = calculatedFirstColumnWidth;

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

  let rowScrollOffset = 0;
  $: rowScrollOffset = $virtualizer?.scrollOffset || 0;

  // In this virtualization model, we create buffer rows before and after our real data
  // This maintains the "correct" scroll position when the user scrolls
  $: [before, after] = virtualRows.length
    ? [
        (virtualRows[1]?.start ?? virtualRows[0].start) - ROW_HEIGHT,
        totalRowSize - virtualRows[virtualRows.length - 1].end,
      ]
    : [0, 0];

  function handleExpandedChange(updater) {
    expanded = updater(expanded);
    metricsExplorerStore.setPivotExpanded($metricsViewName, expanded);
  }

  function handleSorting(updater) {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }
    metricsExplorerStore.setPivotSort($metricsViewName, sorting);
  }

  const handleScroll = (containerRefElement?: HTMLDivElement | null) => {
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
  };

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
</script>

<div
  class="table-wrapper"
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  class:with-row-dimension={hasDimension}
  style:--total-header-height="{totalHeaderHeight + headerGroups.length}px"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
>
  <table>
    <thead>
      {#each headerGroups as headerGroup}
        <tr>
          {#each headerGroup.headers as header, i}
            {@const sortDirection = header.column.getIsSorted()}
            <th colSpan={header.colSpan}>
              {#if i === 0 && hasDimension}
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
                style:width={i === 0 && hasDimension
                  ? `${firstColumnWidth}px`
                  : "100%"}
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
            <td
              class="ui-copy-number"
              class:border-r={i % measureCount === 0 && i}
            >
              <div class="cell">
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

<style lang="postcss">
  * {
    @apply border-slate-200;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal select-none;
    @apply bg-white;
  }

  .table-wrapper {
    @apply overflow-auto h-fit max-h-full w-fit max-w-full;
    @apply border rounded-md z-40;
  }

  th,
  td {
    @apply p-0 m-0;
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
    @apply border-r border-b;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  .with-row-dimension thead > tr:not(:last-of-type) > th:first-of-type {
    @apply border-b-0;
  }

  thead > tr:last-of-type > th {
    @apply text-right;
  }

  th {
    @apply relative;
  }

  td {
    @apply text-right;
  }

  .with-row-dimension tr > td:first-of-type,
  .with-row-dimension tr > th:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  tr > td:first-of-type:not(:last-of-type) {
    @apply border-r font-medium;
  }

  th,
  td {
    @apply whitespace-nowrap text-xs;
  }

  .cell {
    @apply p-1 px-2;
  }

  tbody > tr:nth-of-type(2) {
    @apply bg-slate-100 sticky z-20 font-semibold;
    top: var(--total-header-height);
  }

  tbody > tr:first-of-type > td:first-of-type > .cell {
    @apply font-bold;
  }

  td:last-of-type,
  th:last-of-type {
    @apply border-r-0;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }
</style>
