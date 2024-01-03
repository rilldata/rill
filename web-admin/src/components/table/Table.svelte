<script lang="ts">
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import type { ColumnDef, TableOptions } from "@tanstack/table-core/src/types";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";

  export let data: unknown[] = [];
  export let columns: ColumnDef<unknown, unknown>[] = [];
  export let columnVisibility: Record<string, boolean> = {};
  export let maxWidthOverride: string | null = null;

  let maxWidth = maxWidthOverride ?? "max-w-[800px]";

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

  function rerender() {
    options.update((options) => ({
      ...options,
      data: data,
    }));
  }

  // Whenever the input data changes, rerender the table
  $: data && rerender();
</script>

<table class="w-full {maxWidth}">
  <slot name="header" />
  <tbody>
    {#if $table.getRowModel().rows.length === 0}
      <tr>
        <td class="text-center py-4">
          <slot name="empty" />
        </td>
      </tr>
    {:else}
      {#each $table.getRowModel().rows as row}
        <tr>
          {#each row.getVisibleCells() as cell}
            <td class="hover:bg-slate-50">
              <svelte:component
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            </td>
          {/each}
        </tr>
      {/each}
    {/if}
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
