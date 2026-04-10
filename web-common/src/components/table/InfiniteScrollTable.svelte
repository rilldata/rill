<script lang="ts">
  /**
   * A scrollable table with infinite-loading support for paginated data.
   *
   * This component renders all rows in the DOM (no virtualization). It is
   * intended for datasets under ~1,000 rows (admin tables, settings lists,
   * etc.). For large datasets that need virtualization, use VirtualizedTable.
   */
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";
  import type {
    ColumnDef,
    OnChangeFn,
    SortingState,
    TableOptions,
  } from "tanstack-table-8-svelte-5";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "tanstack-table-8-svelte-5";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";

  export let data: any[];
  export let columns: ColumnDef<any, any>[];
  export let hasNextPage: boolean;
  export let isFetchingNextPage: boolean;
  export let onLoadMore: () => void;
  export let emptyStateMessage = "No items found";
  export let maxHeight = "auto";
  export let headerIcons: Record<string, { icon: any; href: string }> = {};
  export let scrollToTopTrigger: any = null;

  let scrollContainerEl: HTMLDivElement;
  let sentinelEl: HTMLDivElement;
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

  $: options = writable<TableOptions<any>>({
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

  $: table = createSvelteTable(options);

  $: rows = $table.getRowModel().rows;

  // IntersectionObserver-based infinite loading
  let observer: IntersectionObserver | null = null;

  function setupObserver() {
    observer?.disconnect();
    if (typeof IntersectionObserver === "undefined") return;

    observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          onLoadMore();
        }
      },
      { root: scrollContainerEl, rootMargin: "200px" },
    );

    if (sentinelEl) observer.observe(sentinelEl);
  }

  $: if (scrollContainerEl && sentinelEl) {
    setupObserver();
  }

  onDestroy(() => observer?.disconnect());

  // Auto scroll to top when scrollToTopTrigger changes
  $: if (scrollToTopTrigger !== null && scrollContainerEl) {
    scrollContainerEl.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  }
</script>

<div
  class="list scroll-container"
  bind:this={scrollContainerEl}
  style:max-height={maxHeight}
>
  <table>
    <thead>
      {#each $table.getHeaderGroups() as headerGroup}
        <tr class="h-10">
          {#each headerGroup.headers as header (header.id)}
            {@const widthPercent = header.column.columnDef.meta?.widthPercent}
            {@const marginLeft = header.column.columnDef.meta?.marginLeft}
            <th
              colSpan={header.colSpan}
              style={`width: ${widthPercent}%;`}
              class="px-4 py-2 text-left"
              onclick={header.column.getToggleSortingHandler()}
            >
              {#if !header.isPlaceholder}
                <div
                  style={`margin-left: ${marginLeft};`}
                  class:cursor-pointer={header.column.getCanSort()}
                  class:select-none={header.column.getCanSort()}
                  class="font-semibold text-fg-secondary flex flex-row items-center gap-x-1 text-sm"
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
                      class="hover:text-fg-primary"
                    >
                      <svelte:component
                        this={headerIcons[header.column.id].icon}
                        class="text-fg-secondary"
                        size="11px"
                        strokeWidth={2}
                      />
                    </a>
                  {/if}
                  {#if header.column.getIsSorted()}
                    <span>
                      <ArrowDown
                        flip={header.column.getIsSorted().toString() === "asc"}
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
      {#if rows.length === 0}
        <tr>
          <td
            colspan={columns.length}
            class="px-4 py-4 text-center text-fg-secondary"
          >
            {emptyStateMessage}
          </td>
        </tr>
      {:else}
        {#each rows as row (row.id)}
          <tr>
            {#each row.getVisibleCells() as cell (cell.id)}
              <td
                class={`px-4 py-2 max-w-[200px] truncate ${cell.column.id === "actions" ? "w-1" : ""}`}
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
        {#if hasNextPage}
          <tr class="h-0">
            <td colspan={columns.length} class="p-0 border-0">
              <div bind:this={sentinelEl} />
            </td>
          </tr>
        {/if}
      {/if}
    </tbody>
  </table>
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
    @apply sticky top-0 z-30 bg-surface-subtle;
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
