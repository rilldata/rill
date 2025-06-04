<script lang="ts">
  import { writable } from "svelte/store";
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
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";

  export let data: any[];
  export let columns: ColumnDef<any, any>[];
  export let hasNextPage: boolean;
  export let isFetchingNextPage: boolean;
  export let emptyStateMessage = "No items found";
  export let rowHeight = 69;
  export let overscan = 5;
  export let maxHeight = "auto";
  export let headerIcons: Record<string, { icon: any; href: string }> = {};
  export let onLoadMore: () => void;

  let virtualListEl: HTMLDivElement;
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

  // Optimize data handling
  $: safeData = Array.isArray(data) ? data : [];

  const setSorting: OnChangeFn<SortingState> = (updater) => {
    if (updater instanceof Function) {
      sorting = updater(sorting);
    } else {
      sorting = updater;
    }
  };

  // Create table options store once and update it reactively
  const options = writable<TableOptions<any>>({
    data: safeData,
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    enableSortingRemoval: false,
  });

  // Update options only when necessary
  $: options.update((old) => ({
    ...old,
    data: safeData,
    columns,
    state: {
      ...old.state,
      sorting,
    },
  }));

  $: table = createSvelteTable(options);
  $: rows = $table.getRowModel().rows;

  // Create virtualizer once and update options separately
  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => rowHeight,
    overscan,
  });

  // Memoize virtual items to avoid repeated computations
  $: virtualItems = $virtualizer.getVirtualItems();

  // Optimize virtualizer count updates
  $: virtualCount = hasNextPage ? rows.length + 1 : rows.length;
  $: $virtualizer.setOptions({ count: virtualCount });

  // Optimize infinite scroll trigger - avoid array spread
  $: {
    if (virtualItems.length > 0) {
      const lastItem = virtualItems[virtualItems.length - 1];

      if (
        lastItem &&
        lastItem.index > rows.length - 1 &&
        hasNextPage &&
        !isFetchingNextPage
      ) {
        onLoadMore();
      }
    }
  }

  // Memoize header groups to avoid repeated calls
  $: headerGroups = $table.getHeaderGroups();

  // Memoize empty state check
  $: isEmpty = rows.length === 0;
</script>

<div
  class={`list scroll-container`}
  bind:this={virtualListEl}
  style:max-height={maxHeight}
>
  <div class="table-wrapper" style="position: relative;">
    <table>
      <thead>
        {#each headerGroups as headerGroup}
          <tr class="h-10">
            {#each headerGroup.headers as header (header.id)}
              {@const widthPercent = header.column.columnDef.meta?.widthPercent}
              {@const marginLeft = header.column.columnDef.meta?.marginLeft}
              {@const canSort = header.column.getCanSort()}
              {@const isSorted = header.column.getIsSorted()}
              <th
                colSpan={header.colSpan}
                style={widthPercent ? `width: ${widthPercent}%;` : ""}
                class="px-4 py-2 text-left"
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  <div
                    style={marginLeft ? `margin-left: ${marginLeft};` : ""}
                    class:cursor-pointer={canSort}
                    class:select-none={canSort}
                    class="font-semibold text-gray-500 flex flex-row items-center gap-x-1 text-sm"
                  >
                    <svelte:component
                      this={flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                    />
                    {#if headerIcons[header.column.id]}
                      <a
                        href={headerIcons[header.column.id].href}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="hover:text-gray-700"
                      >
                        <svelte:component
                          this={headerIcons[header.column.id].icon}
                          class="text-gray-500"
                          size="11px"
                          strokeWidth={2}
                        />
                      </a>
                    {/if}
                    {#if isSorted}
                      <span>
                        <ArrowDown
                          flip={isSorted.toString() === "asc"}
                          size="12px"
                        />
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
        {#if isEmpty}
          <tr>
            <td
              colspan={columns.length}
              class="px-4 py-4 text-center text-gray-500"
            >
              {emptyStateMessage}
            </td>
          </tr>
        {:else}
          {#each virtualItems as virtualRow, idx (virtualRow.index)}
            {@const rowData = rows[virtualRow.index]}
            {@const transformY = virtualRow.start - idx * virtualRow.size}
            <tr
              style="height: {virtualRow.size}px; transform: translateY({transformY}px);"
            >
              {#each rowData?.getVisibleCells() ?? [] as cell (cell.id)}
                {@const isActionsColumn = cell.column.id === "actions"}
                <td
                  class={`px-4 py-2 max-w-[200px] truncate ${isActionsColumn ? "w-1" : ""}`}
                  data-label={cell.column.columnDef.header}
                >
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
</div>

<style lang="postcss">
  table {
    @apply border-separate border-spacing-0 w-full;
  }
  table th,
  table td {
    @apply border-b border-gray-200;
  }
  thead {
    @apply sticky top-0 z-30 bg-white;
  }
  thead tr th {
    @apply border-t border-gray-200;
  }
  thead tr th:first-child {
    @apply border-l;
    @apply rounded-tl-sm;
  }
  thead tr th:last-child {
    @apply border-r;
    @apply rounded-tr-sm;
  }
  thead tr:last-child th {
    @apply border-b;
  }
  tbody tr:first-child {
    @apply border-t-0;
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
    width: 100%;
    overflow-y: auto;
  }
</style>
