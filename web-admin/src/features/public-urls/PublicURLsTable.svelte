<script lang="ts">
  import { writable } from "svelte/store";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import type {
    ColumnDef,
    OnChangeFn,
    SortingState,
    TableOptions,
  } from "@tanstack/svelte-table";
  import PublicURLsActionsRow from "./PublicURLsActionsRow.svelte";
  import DashboardLink from "./DashboardLink.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
  import { createVirtualizer } from "@tanstack/svelte-virtual";

  interface MagicAuthTokenProps extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  const ROW_HEIGHT = 40;
  const OVERSCAN = 5;

  export let data: MagicAuthTokenProps[];
  export let query: any;
  export let onDelete: (deletedTokenId: string) => void;

  let virtualListEl: HTMLDivElement;
  let sorting: SortingState = [];

  function formatDate(value: string) {
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  // Re-render table when data changes
  $: safeData = Array.isArray(data) ? data : [];
  $: {
    if (safeData) {
      options.update((old) => ({
        ...old,
        data: safeData,
      }));
    }
  }

  const columns: ColumnDef<MagicAuthTokenProps, any>[] = [
    {
      accessorKey: "title",
      header: "Label",
      cell: ({ row }) =>
        flexRender(DashboardLink, {
          href: row.original.url,
          title: row.original.displayName,
        }),
    },
    {
      accessorFn: (row) => row.dashboardTitle,
      header: "Dashboard title",
    },
    {
      accessorKey: "expiresOn",
      header: "Expires on",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
    {
      accessorFn: (row) => row.attributes.name,
      header: "Created by",
    },
    {
      accessorKey: "usedOn",
      header: "Last acccesed",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(PublicURLsActionsRow, {
          id: row.original.id,
          url: row.original.url,
          onDelete,
        }),
    },
  ];

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

  const options = writable<TableOptions<MagicAuthTokenProps>>({
    data: safeData,
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);

  $: rows = $table.getRowModel().rows;

  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });

  $: {
    $virtualizer.setOptions({
      count: query.hasNextPage ? safeData.length + 1 : safeData.length,
    });

    const [lastItem] = [...$virtualizer.getVirtualItems()].reverse();

    if (
      lastItem &&
      lastItem.index > safeData.length - 1 &&
      query.hasNextPage &&
      !query.isFetchingNextPage
    ) {
      query.fetchNextPage();
    }
  }
</script>

<div class="list scroll-container" bind:this={virtualListEl}>
  <!-- FIXME: hide the bleeding corner in the sticky header -->
  <div
    class="table-wrapper"
    style="position: relative; height: {$virtualizer.getTotalSize()}px;"
  >
    <table>
      <thead>
        {#each $table.getHeaderGroups() as headerGroup}
          <tr class="h-10">
            {#each headerGroup.headers as header}
              <th
                colSpan={header.colSpan}
                class="px-4 py-2 text-left"
                on:click={header.column.getToggleSortingHandler()}
              >
                {#if !header.isPlaceholder}
                  <div
                    class:cursor-pointer={header.column.getCanSort()}
                    class:select-none={header.column.getCanSort()}
                    class="font-semibold text-gray-500 flex flex-row items-center gap-x-1"
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
    height: 680px;
    width: 100%;
    overflow-y: auto;
  }
</style>
