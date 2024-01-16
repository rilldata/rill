<script lang="ts">
  import { Writable, writable } from "svelte/store";
  import {
    TableOptions,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
  } from "@tanstack/svelte-table";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";

  export let data;
  export let columns;

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName } = stateManagers;

  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: sorting = $dashboardStore?.pivot?.sorting ?? [];

  function handleExpandedChange(updater) {
    expanded = updater(expanded);
    metricsExplorerStore.setPivotExpanded($metricsViewName, expanded);

    options.update((options) => ({
      ...options,
      state: {
        expanded,
      },
    }));
  }

  function handleSorting(updater) {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }

    metricsExplorerStore.setPivotSort($metricsViewName, sorting);
    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        sorting,
      },
    }));
  }

  const options: Writable<TableOptions<PivotDataRow>> = writable({
    data: data,
    columns: columns,
    state: {
      expanded,
    },
    onExpandedChange: handleExpandedChange,
    getSubRows: (row) => row.subRows,
    onSortingChange: handleSorting,
    getExpandedRowModel: getExpandedRowModel(),
    getCoreRowModel: getCoreRowModel(),
    enableSortingRemoval: false,
    enableExpanding: true,
  });

  let table = createSvelteTable(options);

  function rerender() {
    options.update((options) => ({
      ...options,
      data: data,
    }));

    // FIXME: This is a hack to force the table to rerender, upadting
    // the options in itself doesn't seem to work
    table = createSvelteTable(options);
  }

  // Whenever the input data changes, rerender the table
  $: data && rerender();
</script>

<div class="p-2">
  <table>
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr>
          {#each headerGroup.headers as header}
            <th colSpan={header.colSpan}>
              {#if !header.isPlaceholder}
                <!-- TODO: Fix svelte a11y issues -->
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <!-- svelte-ignore a11y-no-static-element-interactions -->
                <div
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
                </div>
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
                <svelte:component this={result.component} {...result.props} />
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
  <div class="h-4" />
  <button on:click={() => rerender()} class="border p-2"> Rerender </button>
</div>

<style>
  table {
    width: 100%;
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
