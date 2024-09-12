<script lang="ts">
  import { writable } from "svelte/store";
  import {
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getPaginationRowModel,
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

  export let tableData: V1MagicAuthToken[];
  export let pageSize: number;
  export let onDelete: (deletedTokenId: string) => void;
  export let onLoadMore: () => void;
  export let hasNextPage: boolean;

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

  const columns: ColumnDef<V1MagicAuthToken>[] = [
    {
      accessorFn: (row) => row.token,
      header: "Public URL name",
      cell: ({ row }) =>
        flexRender(DashboardLink, {
          href: row.original.url,
          title: row.original.title,
        }),
    },
    {
      // TODO: use dashboard title instead of metricsView
      accessorFn: (row) => row.metricsView,
      header: "Dashboard name",
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
      enableSorting: false,
    },
    {
      accessorKey: "usedOn",
      header: "Last used",
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
    updateTableOptions();
  };

  const options = writable<TableOptions<V1MagicAuthToken>>({
    data: tableData,
    columns: columns,
    state: {
      sorting,
      pagination: {
        pageSize,
        pageIndex: 0, // Always 0 since we're using cursor-based pagination
      },
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    manualPagination: true,
  });

  const table = createSvelteTable(options);

  function updateTableOptions() {
    options.update((old) => ({
      ...old,
      state: {
        ...old.state,
        sorting,
      },
    }));
  }

  function updateTable(data: V1MagicAuthToken[]) {
    options.update((old) => ({
      ...old,
      data: data,
    }));
  }

  $: {
    // Update table when magicAuthTokens changes
    if (tableData) {
      updateTable(tableData);
    }
  }
</script>

<table class="w-full">
  <thead>
    {#each $table.getHeaderGroups() as headerGroup}
      <tr>
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
            <!-- hover:bg-slate-50  -->
            <td class="px-4 py-2" data-label={cell.column.columnDef.header}>
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

{#if hasNextPage}
  <div class="flex items-center gap-2 mt-2 justify-between">
    <button
      class="border rounded px-3 py-1 text-xs font-medium disabled:opacity-50 disabled:pointer-events-none w-full"
      on:click={onLoadMore}
      disabled={!hasNextPage}
    >
      Load More
    </button>
    <!-- <span class="flex items-center gap-1">
    <p class="text-sm font-medium">Rows per page</p>
    <select
      bind:value={pageSize}
      on:change={(e) => onPageSizeChange(Number(e.target.value))}
      class="border p-1 rounded"
    >
      {#each [10, 20, 30, 40, 50] as size}
        <option value={size}>{size}</option>
      {/each}
    </select>
  </span> -->
  </div>
{/if}

<style lang="postcss">
  table {
    @apply border-separate border-spacing-0;
  }
  table th,
  table td {
    @apply border-b border-gray-200;
  }

  thead tr th {
    @apply border-t border-gray-200;
  }
  thead tr th:first-child {
    @apply border-l rounded-tl-sm;
  }
  thead tr th:last-child {
    @apply border-r rounded-tr-sm;
  }
  thead tr:last-child th {
    @apply border-b;
  }
  tbody tr {
    @apply border-t border-gray-200;
  }
  tbody tr:first-child {
    @apply border-t-0;
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
