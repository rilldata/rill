<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    TableOptions,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
  } from "@tanstack/svelte-table";
  import type { Readable } from "svelte/motion";
  import { derived } from "svelte/store";
  import type { PivotDataRow, PivotDataStore } from "./types";
  import { ChevronDown } from "lucide-svelte";
  import { getMeasureCountInColumn } from "./pivot-utils";

  export let pivotDataStore: PivotDataStore;

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    metricsViewName,
    selectors: {
      measures: { visibleMeasures },
    },
  } = stateManagers;

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

  $: assembled = $pivotDataStore.assembled;
  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];
  $: columnPage = $dashboardStore.pivot.columnPage;
  $: totalColumns = $pivotDataStore.totalColumns;

  $: headerGroups = $table.getHeaderGroups();

  $: measureCount = getMeasureCountInColumn(
    $dashboardStore.pivot,
    $visibleMeasures,
  );

  $: console.log(columnPage, totalColumns);

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

  // TODO: Ideally we would like to handle page changes by knowing the scroll
  // position of the container and getting x0, x1, y0, y1 from the table
  // Called when the user scrolls and possibly on mount to fetch more data as the user scrolls
  const handleScroll = (containerRefElement?: HTMLDivElement | null) => {
    if (containerRefElement) {
      // const { scrollWidth, scrollLeft, clientWidth } = containerRefElement;
      // const rightEndDistance = scrollWidth - scrollLeft - clientWidth;
      // const leftEndDistance = scrollLeft;
      // // Distance threshold (in pixels) for triggering data fetch
      // const threshold = 500;
      // // Fetch more data when scrolling near the right end
      // if (
      //   rightEndDistance < threshold &&
      //   !$pivotDataStore.isFetching &&
      //   30 * columnPage < totalColumns
      // ) {
      //   metricsExplorerStore.setPivotColumnPage(
      //     $metricsViewName,
      //     columnPage + 1,
      //   );
      // }
      // // Decrease page number when scrolling near the left end
      // else if (
      //   leftEndDistance < threshold &&
      //   columnPage > 1 // Ensure we don't go below the first page
      // ) {
      //   metricsExplorerStore.setPivotColumnPage(
      //     $metricsViewName,
      //     columnPage - 1,
      //   );
      // }
    }
  };
</script>

<div
  class="overflow-scroll h-full border rounded-md bg-white"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
>
  <table class="overflow-scroll">
    <thead>
      {#each headerGroups as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th colSpan={header.colSpan}>
              <div class="header-cell">
                {#if !header.isPlaceholder}
                  <button
                    class:cursor-pointer={header.column.getCanSort()}
                    class:select-none={header.column.getCanSort()}
                    on:click={header.column.getToggleSortingHandler()}
                  >
                    {header.column.columnDef.header}
                    {#if header.column.getIsSorted()}
                      {#if header.column.getIsSorted().toString() === "asc"}
                        <span>▼</span>
                        <ChevronDown />
                      {:else}
                        <span>▲</span>
                      {/if}
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
      {#each $table.getRowModel().rows as row}
        <tr>
          {#each row.getVisibleCells() as cell, i}
            {@const result =
              typeof cell.column.columnDef.cell === "function"
                ? cell.column.columnDef.cell(cell.getContext())
                : cell.column.columnDef.cell}
            <td
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
    </tbody>
  </table>
</div>

<style lang="postcss">
  table {
    @apply bg-white;
  }

  * {
    @apply border-slate-200;
  }

  thead {
    @apply sticky top-0;
    @apply z-10;
  }

  .header-cell {
    @apply w-full h-full;
    @apply bg-white;
    @apply p-2 px-2;
    @apply border-r border-b;
    @apply text-left;
  }

  thead > tr:first-of-type > th:first-of-type > .header-cell {
    @apply border-b-0;
  }

  thead > tr:last-of-type > th > div {
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

  tr > th:first-of-type,
  tr > td:first-of-type {
    @apply sticky left-0;
    @apply bg-white;
  }

  tr > td:first-of-type > .cell {
    @apply border-r font-medium;
  }

  th,
  td {
    @apply whitespace-nowrap text-xs;
  }

  .cell {
    @apply p-1 px-2;
  }

  tbody > tr:first-of-type {
    @apply bg-slate-100;
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
</style>
