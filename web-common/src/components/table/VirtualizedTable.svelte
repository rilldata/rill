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
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { writable } from "svelte/store";

  export let data: any[];
  export let columns: ColumnDef<any, any>[];
  export let emptyIcon: any | null = null;
  export let emptyText = "No data available";
  export let columnLayout = `repeat(${columns.length}, 1fr)`;
  export let rowPadding = "py-3";
  export let rowHeight = 46;
  export let containerHeight = 400;
  export let overscan = 1;
  export let tableId: string | undefined = undefined;

  let containerElement: HTMLDivElement;
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

  $: ({ getHeaderGroups } = $table);
  $: headers = getHeaderGroups();

  $: rows = $table.getRowModel().rows;
  $: virtualizer = createVirtualizer({
    count: rows.length,
    getScrollElement: () => containerElement,
    estimateSize: () => rowHeight,
    overscan: overscan,
    initialOffset: rowScrollOffset,
  });

  $: virtualRows = $virtualizer.getVirtualItems();
  $: rowScrollOffset = $virtualizer?.scrollOffset || 0;

  $: dynamicContainerHeight =
    rows.length <= 10 ? rowHeight * rows.length : containerHeight;
</script>

<div
  id={tableId}
  class="flex flex-col border rounded-sm overflow-hidden"
  style:--grid-template-columns={columnLayout}
>
  {#each headers as headerGroup (headerGroup.id)}
    <div class="row sticky top-0 z-30 bg-surface-subtle">
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

  <div
    bind:this={containerElement}
    class="relative overflow-auto"
    style="height: {dynamicContainerHeight}px;"
  >
    {#if !rows || rows.length === 0}
      <div class="flex flex-col items-center gap-y-1 py-10">
        {#if emptyIcon}
          <svelte:component this={emptyIcon} size={32} color="#CBD5E1" />
        {/if}
        <span class="text-fg-secondary font-semibold text-sm">{emptyText}</span>
      </div>
    {:else}
      <div
        class="relative w-full"
        style="height: {$virtualizer.getTotalSize()}px;"
      >
        {#each virtualRows as virtualRow, i (i)}
          {@const row = rows[virtualRow.index]}
          <div
            class="row {rowPadding} absolute top-0 left-0 w-full"
            style="transform: translateY({virtualRow.start}px);"
          >
            {#each row.getVisibleCells() as cell (cell.id)}
              <div
                class="pl-{cell.column.columnDef.meta?.marginLeft ||
                  '4'} pr-1 flex items-center truncate"
              >
                <svelte:component
                  this={flexRender(
                    cell.column.columnDef.cell,
                    cell.getContext(),
                  )}
                />
              </div>
            {/each}
          </div>
        {/each}
      </div>
    {/if}
  </div>
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
