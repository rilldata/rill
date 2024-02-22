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
  import type { Readable } from "svelte/motion";
  import { derived } from "svelte/store";
  import type { PivotDataRow, PivotDataStore } from "./types";

  export let pivotDataStore: PivotDataStore;

  const OVERSCAN = 80;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName } = stateManagers;

  const pivotDashboardStore = derived(dashboardStore, (dashboard) => {
    return dashboard?.pivot;
  });

  const options: Readable<TableOptions<PivotDataRow>> = derived(
    [pivotDashboardStore, pivotDataStore],
    ([pivotConfig, pivotData]) => ({
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
    }),
  );

  const table = createSvelteTable(options);

  let containerRefElement: HTMLDivElement;
  let stickyRows = [0];

  $: assembled = $pivotDataStore.assembled;
  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];

  $: headerGroups = $table.getHeaderGroups();
  $: measureCount = $dashboardStore.pivot?.columns?.measure?.length ?? 0;
  $: rows = $table.getRowModel().rows;
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;

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
</script>

<div
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-length="{totalHeaderHeight}px"
  class="overflow-scroll h-fit max-h-full border rounded-md bg-white"
  bind:this={containerRefElement}
>
  <div style:height="{totalRowSize + totalHeaderHeight}px">
    <table>
      <thead>
        {#each headerGroups as headerGroup}
          <tr>
            {#each headerGroup.headers as header}
              {@const sortDirection = header.column.getIsSorted()}
              <th
                colSpan={header.colSpan}
                class:with-row-dimension={rows.length > 1}
              >
                <div class="header-cell" style:height="{HEADER_HEIGHT}px">
                  {#if !header.isPlaceholder}
                    <button
                      class="flex items-center gap-x-1"
                      class:cursor-pointer={header.column.getCanSort()}
                      class:select-none={header.column.getCanSort()}
                      on:click={header.column.getToggleSortingHandler()}
                    >
                      {header.column.columnDef.header}
                      {#if sortDirection}
                        <span
                          class="transition-transform -mr-1"
                          class:-rotate-180={sortDirection === "desc"}
                        >
                          <ArrowDown />
                        </span>
                      {/if}
                    </button>
                  {:else}
                    <button class="w-full h-full"></button>
                  {/if}
                </div>
              </th>
            {/each}
          </tr>
        {/each}
      </thead>
      <tbody>
        <tr>
          <td colspan={headerGroups.length} style:height="{before}px"> </td>
        </tr>
        {#each virtualRows as row (row.index)}
          {@const cells = rows[row.index].getVisibleCells()}
          <tr>
            {#each cells as cell, i (cell.id)}
              {@const result =
                typeof cell.column.columnDef.cell === "function"
                  ? cell.column.columnDef.cell(cell.getContext())
                  : cell.column.columnDef.cell}
              <td
                class:with-row-dimension={rows.length > 1}
                class="ui-copy-number"
                class:border-right={i % measureCount === 0 && i}
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
        <tr>
          <td colspan={headerGroups.length} style:height="{after}px"></td>
        </tr>
      </tbody>
    </table>
  </div>
</div>

<style lang="postcss">
  table {
    @apply bg-white;
  }

  * {
    @apply border-slate-200;
  }

  /* Pin header */
  thead {
    @apply sticky top-0;
    @apply z-10;
  }

  .header-cell {
    @apply w-full h-full;
    @apply bg-white;
    @apply px-2;
    @apply flex items-center justify-start;
    @apply border-r border-b;
    @apply text-left;
    @apply text-ellipsis whitespace-nowrap overflow-hidden;
  }

  /* The leftmost header cells have no bottom border unless they're the last row */
  thead
    > tr:not(:last-of-type)
    > .with-row-dimension:first-of-type
    > .header-cell {
    @apply border-b-0;
  }

  thead > tr:last-of-type > th > .header-cell {
    @apply text-right;
  }

  th {
    @apply p-0 m-0;
  }

  td {
    @apply border-none;
    @apply text-right;
    @apply p-0 m-0;
  }

  tr > .with-row-dimension:first-of-type,
  tr > .with-row-dimension:first-of-type {
    @apply sticky left-0 z-0;
    @apply bg-white;
  }

  tr > td:first-of-type:not(:last-of-type) > .cell {
    @apply border-r font-medium;
  }

  th,
  td {
    @apply whitespace-nowrap text-xs;
  }

  .cell {
    @apply p-1 px-2;
    height: var(--row-height);
  }

  tbody > tr:nth-of-type(2) {
    @apply bg-slate-100 sticky z-10 font-semibold;
    top: var(--header-length);
  }

  .border-right {
    border-right: solid black 1px;
    @apply border-gray-200;
  }

  tbody > tr:first-of-type > td:first-of-type > .cell {
    @apply font-bold;
  }

  td:last-of-type,
  th:last-of-type > .header-cell {
    @apply border-r-0;
  }

  tr:hover,
  tr:hover .cell {
    @apply bg-slate-100;
  }
</style>
