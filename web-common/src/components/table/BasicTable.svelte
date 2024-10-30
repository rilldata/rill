<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type {
    ColumnDef,
    OnChangeFn,
    SortingState,
    TableOptions,
  } from "@tanstack/svelte-table";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import { writable } from "svelte/store";

  export let data: any[];
  export let columns: ColumnDef<any, any>[];
  export let emptyText = "No data available";
  export let scrollable = false;

  let sorting: SortingState = [];

  $: safeData = Array.isArray(data) ? data : [];
  $: {
    if (safeData) {
      options.update((old) => ({
        ...old,
        data: safeData,
      }));
    }
  }

  const setSorting: OnChangeFn<SortingState> = (updater) => {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }

    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        sorting,
      },
    }));
  };

  const options = writable<TableOptions<any>>({
    data: data,
    columns: columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);
</script>

<div class="overflow-x-auto" class:scroll-container={scrollable}>
  <table class="w-full">
    <thead class={scrollable ? "sticky top-0 z-30 bg-white" : ""}>
      {#each $table.getHeaderGroups() as headerGroup (headerGroup.id)}
        <tr>
          {#each headerGroup.headers as header (header.id)}
            {@const widthPercent = header.column.columnDef.meta?.widthPercent}
            {@const marginLeft = header.column.columnDef.meta?.marginLeft}
            <th
              colSpan={header.colSpan}
              style={`width: ${widthPercent}%;`}
              class="px-4 py-2 text-left"
              on:click={header.column.getToggleSortingHandler()}
            >
              {#if !header.isPlaceholder}
                <div
                  style={`margin-left: ${marginLeft};`}
                  class:cursor-pointer={header.column.getCanSort()}
                  class:select-none={header.column.getCanSort()}
                  class="font-semibold text-gray-500 flex flex-row items-center gap-x-1 truncate"
                >
                  <svelte:component
                    this={flexRender(
                      header.column.columnDef.header,
                      header.getContext(),
                    )}
                  />
                  {#if header.column.getIsSorted().toString() === "asc"}
                    <span>
                      <ArrowDown flip size="12px" />
                    </span>
                  {:else if header.column.getIsSorted().toString() === "desc"}
                    <span>
                      <ArrowDown size="12px" />
                    </span>
                  {/if}
                </div>
              {/if}
            </th>
          {/each}
        </tr>
      {/each}
    </thead>
    <tbody>
      {#if $table.getRowModel().rows.length === 0}
        <tr>
          <td
            colspan={columns.length}
            class="px-4 py-4 text-center text-gray-500"
          >
            {emptyText}
          </td>
        </tr>
      {:else}
        {#each $table.getRowModel().rows as row (row.id)}
          <tr>
            {#each row.getVisibleCells() as cell (cell.id)}
              <td class="px-4 py-2" data-label={cell.column.columnDef.header}>
                <svelte:component
                  this={flexRender(
                    cell.column.columnDef.cell,
                    cell.getContext(),
                  )}
                />
              </td>
            {/each}
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>

<style lang="postcss">
  table {
    @apply border-separate border-spacing-0;
  }
  table th,
  table td {
    @apply border-b border-gray-200;
  }

  thead tr th {
    @apply border-t border-gray-200;
  }
  thead tr th:first-child {
    @apply border-l rounded-tl-sm;
  }
  thead tr th:last-child {
    @apply border-r rounded-tr-sm;
  }
  thead tr:last-child th {
    @apply border-b;
  }
  tbody tr {
    @apply border-t border-gray-200;
  }
  tbody tr:first-child {
    @apply border-t-0;
  }
  tbody td {
    @apply border-b border-gray-200;
  }
  tbody td:first-child {
    @apply border-l;
  }
  tbody td:last-child {
    @apply border-r;
  }
  tbody tr:last-child td:first-child {
    @apply rounded-bl-sm;
  }
  tbody tr:last-child td:last-child {
    @apply rounded-br-sm;
  }

  .scroll-container {
    height: 680px;
    width: 100%;
    overflow-y: auto;
  }
</style>
