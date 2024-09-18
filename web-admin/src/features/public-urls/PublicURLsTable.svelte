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

  interface MagicAuthTokenProps extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  export let tableData: MagicAuthTokenProps[];
  export let pageSize: number;
  export let onDelete: (deletedTokenId: string) => void;

  let sorting: SortingState = [
    {
      id: "createdOn",
      desc: true,
    },
  ];

  function formatDate(value: string) {
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  const columns: ColumnDef<MagicAuthTokenProps, any>[] = [
    {
      accessorKey: "title",
      header: "Label",
      cell: ({ row }) =>
        flexRender(DashboardLink, {
          href: row.original.url,
          title: row.original.title,
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
      enableSorting: false,
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
      accessorKey: "createdOn",
      header: "Created on",
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

  const options = writable<TableOptions<MagicAuthTokenProps>>({
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

  function updateTable(data: MagicAuthTokenProps[]) {
    options.update((old) => ({
      ...old,
      data: data,
    }));
  }

  // Update table when magicAuthTokens changes
  $: {
    if (tableData) {
      updateTable(tableData);
    }
  }
</script>

<div class="list scroll-container">
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
      {#each $table.getRowModel().rows as row}
        <tr>
          {#each row.getVisibleCells() as cell}
            <td
              class={`px-4 py-2 ${cell.column.id === "actions" ? "w-1" : ""}`}
              data-label={cell.column.columnDef.header}
            >
              <svelte:component
                this={flexRender(cell.column.columnDef.cell, cell.getContext())}
              />
            </td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
</div>

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

  .scroll-container {
    height: 800px;
    width: 100%;
    overflow: auto;
  }
</style>
