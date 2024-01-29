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

  export let pivotDataStore: PivotDataStore;

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName } = stateManagers;

  $: assembled = $pivotDataStore.assembled;
  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];
  $: columnPage = $dashboardStore.pivot.columnPage;
  $: totalColumns = $pivotDataStore.totalColumns;

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

  let table = createSvelteTable(options);

  $: console.log(columnPage, totalColumns);

  let containerRefElement;

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
  class="overflow-x-auto relative h-full"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
>
  <table class="mx-2">
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th colSpan={header.colSpan}>
              {#if !header.isPlaceholder}
                <button
                  class:cursor-pointer={header.column.getCanSort()}
                  class:select-none={header.column.getCanSort()}
                  on:click={header.column.getToggleSortingHandler()}
                >
                  <svelte:component
                    this={flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                  />
                  {#if header.column.getIsSorted()}
                    {#if header.column.getIsSorted().toString() === "asc"}
                      <span>▼</span>
                    {:else}
                      <span>▲</span>
                    {/if}
                  {/if}
                </button>
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </thead>
    <tbody>
      {#each $table.getRowModel().rows as row}
        <tr>
          {#each row.getVisibleCells() as cell}
            {@const result =
              typeof cell.column.columnDef.cell === "function"
                ? cell.column.columnDef.cell(cell.getContext())
                : cell.column.columnDef.cell}
            <td class="ui-copy-number">
              {#if result?.component && result?.props}
                <svelte:component
                  this={result.component}
                  {...result.props}
                  {assembled}
                />
              {:else if typeof result === "string" || typeof result === "number"}
                {result}
              {:else}
                <!-- flexRender is REALLY slow https://github.com/TanStack/table/issues/4962#issuecomment-1821011742 -->
                <svelte:component
                  this={flexRender(
                    cell.column.columnDef.cell,
                    cell.getContext(),
                  )}
                />
              {/if}
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  table {
    min-width: 300px;
    border-collapse: collapse;
    color: #333;
  }

  tbody {
    border-bottom: 1px solid lightgray;
  }

  th,
  td {
    padding: 10px;
    border: 1px solid #ddd;
    text-align: left;
  }

  th {
    background-color: #f2f2f2;
    font-weight: bold;
    outline: 1px solid #ddd;
  }

  tr:nth-child(even) {
    background-color: #f9f9f9;
  }
  tr:first-child {
    font-weight: bold;
  }
  tr:hover {
    background-color: #e8e8e8;
  }

  thead {
    border-bottom: 2px solid #333;
    position: sticky;
    top: 0;
  }

  td {
    text-align: right;
  }
</style>
