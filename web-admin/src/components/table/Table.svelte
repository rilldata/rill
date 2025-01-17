<script lang="ts">
  import type { ColumnDef, TableOptions } from "@tanstack/svelte-table";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
    type Row,
  } from "@tanstack/svelte-table";
  import { createEventDispatcher, setContext } from "svelte";
  import { writable } from "svelte/store";
  import Toolbar from "./Toolbar.svelte";

  export let data: unknown[] = [];
  export let columns: ColumnDef<unknown, unknown>[] = [];
  export let columnVisibility: Record<string, boolean> = {};
  export let kind: string;

  const dispatch = createEventDispatcher();

  let sorting = [];
  function setSorting(updater) {
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
  }

  const options = writable<TableOptions<unknown>>({
    data: data,
    columns: columns,
    globalFilterFn: "auto",
    enableSorting: true,
    enableFilters: true,
    enableGlobalFilter: true,
    state: {
      sorting,
      columnVisibility,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);

  // Expose the table API to the children components via Context
  setContext("table", table);

  function handleClickRow(row: Row<unknown>) {
    dispatch("click-row", row);
  }

  function rerender() {
    options.update((options) => ({
      ...options,
      data: data,
    }));
  }

  // Whenever the input data changes, rerender the table
  $: data && rerender();
</script>

<slot name="toolbar">
  <Toolbar />
</slot>

<table class="w-full">
  <slot name="header" />
  <tbody>
    {#each $table.getRowModel().rows as row (row.id)}
      <tr on:click={() => handleClickRow(row)}>
        {#each row.getVisibleCells() as cell (cell.id)}
          <td class="hover:bg-slate-50">
            <svelte:component
              this={flexRender(cell.column.columnDef.cell, cell.getContext())}
            />
          </td>
        {/each}
      </tr>
    {:else}
      <tr>
        <td class="text-center py-4">
          <span class="text-gray-500"> No {kind}s found. </span>
        </td>
      </tr>
    {/each}
  </tbody>
</table>

<!-- 
Rounded table corners are tricky:
- `border-radius` does not apply to table elements when `border-collapse` is `collapse`.
- You can only apply `border-radius` to <td>, not <tr> or <table>.
-->
<style lang="postcss">
  table {
    @apply border-separate border-spacing-0;
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
</style>
