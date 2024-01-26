<script lang="ts">
  import type { PivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
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
  import type { PivotDataRow } from "./types";

  export let pivotStore: PivotDataStore;

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName } = stateManagers;

  $: assembled = $pivotStore.assembled;
  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];

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

  const options: Readable<TableOptions<PivotDataRow>> = derived(
    pivotStore,
    (pivotData) => ({
      data: pivotData.data,
      columns: pivotData.columnDef,
      state: {
        expanded,
        sorting,
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
</script>

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
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            {/if}
          </td>
        {/each}
      </tr>
    {/each}
  </tbody>
</table>

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
