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
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { V1MagicAuthToken } from "@rilldata/web-admin/client";

  export let magicAuthTokens: V1MagicAuthToken[];
  export let onDelete: (deletedTokenId: string) => void;

  function formatDate(value: string | null) {
    if (!value) return "-";
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
      accessorFn: (row) => row.title || row.metricsView,
      header: "Dashboard name",
    },
    {
      accessorKey: "expiresOn",
      header: "Expires on",
      cell: (info) => {
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

  let sorting: SortingState = [];
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

  const options = writable<TableOptions<V1MagicAuthToken>>({
    data: magicAuthTokens,
    columns: columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);

  // TODO: to be removed
  // function rerender() {
  //   options.update((options) => ({
  //     ...options,
  //     data: magicAuthTokens,
  //   }));
  // }
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
