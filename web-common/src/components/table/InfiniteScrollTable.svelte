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
  export let onLoadMore: () => void;
  export let emptyStateMessage = "No items found";
  export let rowHeight = 69;
  export let overscan = 5;
  export let maxHeight = "auto";
  export let headerIcons: Record<string, { icon: any; href: string }> = {};
  export let scrollToTopTrigger: any = null;

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

  const isSafari =
    typeof window !== "undefined" &&
    /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

  // Constants for table sizing
  const EMPTY_TABLE_MIN_HEIGHT = 100;
  const SAFARI_EXTRA_PADDING = 50;

  let totalSize: number;

  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => rowHeight,
    overscan,
    measureElement: (el) => el?.getBoundingClientRect()?.height ?? rowHeight,
  });

  $: {
    $virtualizer.setOptions({
      count: hasNextPage ? safeData.length + 1 : safeData.length,
    });

    const [lastItem] = [...$virtualizer.getVirtualItems()].reverse();

    if (
      lastItem &&
      lastItem.index > safeData.length - 1 &&
      hasNextPage &&
      !isFetchingNextPage
    ) {
      onLoadMore();
    }
  }

  // Calculate total size, ensuring it updates when data changes
  $: {
    if (safeData.length === 0) {
      totalSize = EMPTY_TABLE_MIN_HEIGHT;
    } else {
      const virtualizerSize = $virtualizer.getTotalSize();
      const minContentSize = safeData.length * rowHeight;

      if (isSafari) {
        // Safari needs extra padding to prevent scrolling issues
        totalSize = Math.max(
          virtualizerSize + SAFARI_EXTRA_PADDING,
          minContentSize,
        );
      } else {
        totalSize = Math.max(virtualizerSize, minContentSize);
      }
    }
  }

  // Auto scroll to top when scrollToTopTrigger changes
  $: if (scrollToTopTrigger !== null && virtualListEl) {
    virtualListEl.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  }
</script>

<div
  class={`list scroll-container`}
  bind:this={virtualListEl}
  style:max-height={maxHeight}
>
  <div
    class="table-wrapper"
    style="min-height: {safeData.length === 0
      ? '100'
      : totalSize}px; width: 100%; position: relative;"
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
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  <div
                    style={`margin-left: ${marginLeft};`}
                    class:cursor-pointer={header.column.getCanSort()}
                    class:select-none={header.column.getCanSort()}
                    class="font-semibold text-muted-foreground flex flex-row items-center gap-x-1 text-sm"
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
                        class="hover:text-surface-foreground"
                      >
                        <svelte:component
                          this={headerIcons[header.column.id].icon}
                          class="text-muted-foreground"
                          size="11px"
                          strokeWidth={2}
                        />
                      </a>
                    {/if}
                    {#if header.column.getIsSorted()}
                      <span>
                        <ArrowDown
                          flip={header.column.getIsSorted().toString() ===
                            "asc"}
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
        {#if $table.getRowModel().rows.length === 0}
          <tr>
            <td
              colspan={columns.length}
              class="px-4 py-4 text-center text-muted-foreground"
            >
              {emptyStateMessage}
            </td>
          </tr>
        {:else}
          {#each $virtualizer.getVirtualItems() as virtualRow, idx (virtualRow.index)}
            <tr
              style="height: {virtualRow.size}px; transform: translateY({virtualRow.start -
                idx * virtualRow.size}px);"
            >
              {#each rows[virtualRow.index]?.getVisibleCells() ?? [] as cell (cell.id)}
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
    @apply sticky top-0 z-30 bg-surface;
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
