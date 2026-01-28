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
  export let emptyIcon: any | null = null;
  export let emptyText = "No data available";
  export let columnLayout = `repeat(${columns.length}, 1fr)`;
  export let rowPadding = "py-3";

  let sorting: SortingState = [];

  // Initialize sorting for sortDescFirst column
  const sortDescFirstColumn = columns.find((col) => col.sortDescFirst);
  if (sortDescFirstColumn) {
    const columnId =
      "id" in sortDescFirstColumn
        ? sortDescFirstColumn.id
        : "accessorKey" in sortDescFirstColumn
          ? sortDescFirstColumn.accessorKey
          : "accessorFn" in sortDescFirstColumn
            ? (sortDescFirstColumn.header as string)
            : Object.keys(sortDescFirstColumn)[0];

    sorting = [
      {
        id: columnId as string,
        desc: true,
      },
    ];
  }

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
    data: safeData,
    columns: columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    enableSortingRemoval: false,
  });

  const table = createSvelteTable(options);

  $: ({ getHeaderGroups, getRowModel } = $table);

  $: rows = getRowModel().rows;
  $: headers = getHeaderGroups();
</script>

<div
  class="flex flex-col border rounded-sm overflow-hidden overflow-x-auto"
  style:--grid-template-columns={columnLayout}
>
  {#each headers as headerGroup (headerGroup.id)}
    <div class="row sticky top-0 z-30 bg-transparent">
      {#each headerGroup.headers as header (header.id)}
        <svelte:element
          this={header.column.getCanSort() ? "button" : "div"}
          role="columnheader"
          tabindex="0"
          class="pl-{header.column.columnDef.meta?.marginLeft ||
            '4'} py-2 font-semibold text-fg-secondary text-left flex flex-row items-center gap-x-1 truncate text-sm"
          on:click={header.column.getToggleSortingHandler()}
        >
          {#if !header.isPlaceholder}
            <span class="truncate">
              <svelte:component
                this={flexRender(
                  header.column.columnDef.header,
                  header.getContext(),
                )}
              />
            </span>
            {#if header.column.getIsSorted()}
              <span>
                <ArrowDown
                  flip={header.column.getIsSorted().toString() === "asc"}
                  size="12px"
                />
              </span>
            {/if}
          {/if}
        </svelte:element>
      {/each}
    </div>
  {/each}

  {#each rows as row (row.id)}
    <div class="row {rowPadding}">
      {#each row.getVisibleCells() as cell (cell.id)}
        <div
          class="pl-{cell.column.columnDef.meta?.marginLeft ||
            '4'} pr-1 flex items-center truncate"
        >
          <svelte:component
            this={flexRender(cell.column.columnDef.cell, cell.getContext())}
          />
        </div>
      {/each}
    </div>
  {:else}
    <div class="flex flex-col items-center gap-y-1 py-10">
      {#if emptyIcon}
        <svelte:component this={emptyIcon} size={32} color="#CBD5E1" />
      {/if}
      <span class="text-fg-secondary font-semibold text-sm">{emptyText}</span>
    </div>
  {/each}
</div>

<style lang="postcss">
  * {
    @apply border-gray-200;
  }

  .row {
    @apply w-fit min-w-full;
    display: grid;
    grid-template-columns: var(--grid-template-columns);
  }

  .row:not(:last-of-type) {
    @apply border-b;
  }
</style>
