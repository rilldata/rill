<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { prettyResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  import { writable } from "svelte/store";
  import ResourceErrorMessage from "./ResourceErrorMessage.svelte";
  import {
    getResourceKindTagColor,
    prettyReconcileStatus,
  } from "./display-utils";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
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

  export let resources: V1Resource[];

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

  const columns: ColumnDef<V1Resource, any>[] = [
    {
      accessorKey: "title",
      header: "Type",
      enableSorting: false,
      cell: ({ row }) => {
        const prettyKind = prettyResourceKind(row.original.meta.name.kind);
        const color = getResourceKindTagColor(row.original.meta.name.kind);
        return flexRender(Tag, {
          color,
          text: prettyKind,
        });
      },
    },
    {
      accessorFn: (row) => row.meta.name.name,
      header: "Name",
    },
    {
      accessorFn: (row) => row.meta.reconcileStatus,
      header: "Execution status",
      cell: ({ row }) =>
        prettyReconcileStatus(row.original.meta.reconcileStatus),
    },
    {
      accessorFn: (row) => row.meta.reconcileError,
      header: "Error",
      cell: ({ row }) =>
        flexRender(ResourceErrorMessage, {
          message: row.original.meta.reconcileError,
        }),
    },
    {
      accessorFn: (row) => row.meta.stateUpdatedOn,
      header: "Last refresh",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
    {
      accessorFn: (row) => row.meta.reconcileOn,
      header: "Next refresh",
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
  ];

  const options = writable<TableOptions<V1Resource>>({
    data: resources,
    columns: columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const table = createSvelteTable(options);
</script>

<div class="list">
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
</style>
