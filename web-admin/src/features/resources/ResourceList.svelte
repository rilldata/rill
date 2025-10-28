<script lang="ts">
  import type { ColumnDef, TableOptions } from "@tanstack/svelte-table";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import ResourceListToolbar from "./ResourceListToolbar.svelte";

  export let data: unknown[] = [];
  export let columns: ColumnDef<unknown, unknown>[] = [];
  export let columnVisibility: Record<string, boolean> = {};
  export let kind: string;
  export let toolbar: boolean = true;
  export let fixedRowHeight: boolean = true;

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

  // Check if we're in a filtered state (search is active)
  $: isFiltered = $table.getState().globalFilter?.length > 0;
</script>

<div class="flex flex-col gap-y-3 w-full">
  {#if toolbar}
    <slot name="toolbar">
      <ResourceListToolbar />
    </slot>
  {/if}

  <div class="w-full">
    <slot name="header" />
    <ul role="list" class="resource-list">
      {#each $table.getRowModel().rows as row (row.id)}
        <li class="resource-list-item" class:fixed-height={fixedRowHeight}>
          {#each row.getVisibleCells() as cell (cell.id)}
            <svelte:component
              this={flexRender(cell.column.columnDef.cell, cell.getContext())}
            />
          {/each}
        </li>
      {:else}
        <li class="resource-list-item-empty">
          <div class="text-center py-16">
            {#if isFiltered}
              <!-- Filtered empty state: no results match search -->
              <div class="flex flex-col gap-y-2 items-center text-sm">
                <div class="text-gray-600 font-semibold">
                  No {kind}s match your search
                </div>
                <div class="text-gray-500">Try adjusting your search terms</div>
              </div>
            {:else}
              <!-- Custom empty state via slot, or fallback -->
              <slot name="empty">
                <div class="text-gray-600 text-sm font-semibold">
                  You don't have any {kind}s yet
                </div>
              </slot>
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  </div>
</div>

<style lang="postcss">
  .resource-list {
    @apply list-none p-0 m-0 w-full;
  }

  .resource-list-item,
  .resource-list-item-empty {
    @apply block w-full border border-gray-200;
  }

  .resource-list-item.fixed-height {
    @apply h-[60px];
  }

  /* Remove top border on non-first items to avoid double borders */
  .resource-list-item + .resource-list-item {
    @apply border-t-0;
  }

  /* Rounded corners on first and last items */
  .resource-list-item:first-child,
  .resource-list-item-empty:first-child {
    @apply rounded-t-lg;
  }

  .resource-list-item:last-child,
  .resource-list-item-empty:last-child {
    @apply rounded-b-lg;
  }

  /* Hover effect on list items */
  .resource-list-item:hover {
    @apply bg-slate-50;
  }
</style>
