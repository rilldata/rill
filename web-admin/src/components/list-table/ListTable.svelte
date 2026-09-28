<script lang="ts">
  import type {
    ColumnDef,
    SortingState,
    TableOptions,
  } from "tanstack-table-8-svelte-5";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
  } from "tanstack-table-8-svelte-5";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import ListTableToolbar from "./ListTableToolbar.svelte";
  import { flip } from "svelte/animate";

  // Renders a tanstack table as a bordered list of rows with client-side search, sorting and pinning.
  // Rows are opaque to this component: callers supply the columns and a stable row id.
  export let data: unknown[] = [];
  export let columns: ColumnDef<unknown, unknown>[] = [];
  export let columnVisibility: Record<string, boolean> = {};
  // Singular noun for the empty states, e.g. "dashboard".
  export let kind: string;
  export let toolbar: boolean = true;
  export let fixedRowHeight: boolean = true;
  export let sorting: SortingState = [];
  // Ids (as returned by getRowId) of rows pinned to the top.
  export let pinnedRows: string[] = [];
  export let maxRows: number | undefined = undefined;
  export let getRowId: (row: unknown, index: number) => string;

  function setSorting(newSorting: SortingState) {
    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        sorting: newSorting,
      },
    }));
  }
  $: setSorting(sorting);

  function setPinned(newPinnedRows: string[]) {
    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        rowPinning: {
          top: [...newPinnedRows],
        },
      },
    }));
  }
  $: setPinned(pinnedRows);

  const options = writable<TableOptions<unknown>>({
    data: data,
    columns: columns,
    globalFilterFn: "auto",
    enableSorting: true,
    enableFilters: true,
    enableGlobalFilter: true,
    enableRowPinning: true,
    state: {
      sorting,
      columnVisibility,
      rowPinning: {},
    },
    getRowId,
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

  $: allRows = [...$table.getTopRows(), ...$table.getCenterRows()];
  $: limitedRows = allRows.slice(0, maxRows ?? allRows.length);
</script>

<div class="flex flex-col gap-y-3 w-full">
  {#if toolbar}
    <slot name="toolbar">
      <ListTableToolbar />
    </slot>
  {/if}

  <div class="w-full">
    <slot name="header" />
    <ul role="list" class="list-table">
      {#each limitedRows as row (row.id)}
        <li
          class="list-table-item"
          class:fixed-height={fixedRowHeight}
          animate:flip={{ duration: 200 }}
        >
          {#each row.getVisibleCells() as cell (cell.id)}
            <svelte:component
              this={flexRender(cell.column.columnDef.cell, cell.getContext())}
            />
          {/each}
        </li>
      {:else}
        <li class="list-table-item-empty">
          <div class="text-center py-16">
            {#if isFiltered}
              <!-- Filtered empty state: no results match search -->
              <div class="flex flex-col gap-y-2 items-center text-sm">
                <div class="text-fg-secondary font-semibold">
                  No {kind}s match your search
                </div>
                <div class="text-fg-secondary">
                  Try adjusting your search terms
                </div>
              </div>
            {:else}
              <!-- Custom empty state via slot, or fallback -->
              <slot name="empty">
                <div class="text-fg-secondary text-sm font-semibold">
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
  .list-table {
    @apply list-none p-0 m-0 w-full;
  }

  .list-table-item,
  .list-table-item-empty {
    @apply block w-full border bg-surface-background;
  }

  .list-table-item.fixed-height {
    @apply h-[60px];
  }

  /* Remove top border on non-first items to avoid double borders */
  .list-table-item + .list-table-item {
    @apply border-t-0;
  }

  /* Rounded corners on first and last items */
  .list-table-item:first-child,
  .list-table-item-empty:first-child {
    @apply rounded-t-lg;
  }

  .list-table-item:last-child,
  .list-table-item-empty:last-child {
    @apply rounded-b-lg;
  }

  /* Hover effect on list items */
  .list-table-item:hover {
    @apply bg-surface-hover;
  }
</style>
