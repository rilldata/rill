<script lang="ts">
  import { createInfiniteQuery } from "@tanstack/svelte-query";
  import {
    ColumnDef,
    Row,
    TableOptions,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { writable } from "svelte/store";
  import {
    V1ModelSplit,
    V1Resource,
    getRuntimeServiceGetModelSplitsQueryKey,
    runtimeServiceGetModelSplits,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import DataCell from "./DataCell.svelte";
  import ErrorCell from "./ErrorCell.svelte";
  import TriggerSplit from "./TriggerSplit.svelte";

  export let resource: V1Resource;
  export let whereErrored: boolean;
  export let wherePending: boolean;

  $: modelName = resource?.meta?.name?.name as string;

  /**
   * Infinite Query
   */
  $: baseParams = {
    ...(whereErrored ? { errored: true } : {}),
    ...(wherePending ? { pending: true } : {}),
  };
  $: query = createInfiniteQuery({
    queryKey: getRuntimeServiceGetModelSplitsQueryKey(
      $runtime.instanceId,
      modelName,
      baseParams,
    ),
    queryFn: ({ pageParam }) => {
      const getModelSplitsParams = {
        ...baseParams,
        ...(pageParam
          ? {
              pageToken: pageParam as string,
            }
          : {}),
      };
      return runtimeServiceGetModelSplits(
        $runtime.instanceId,
        modelName,
        getModelSplitsParams,
      );
    },
    enabled: !!modelName,
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken || lastPage.nextPageToken === "") {
        return undefined;
      }
      return lastPage.nextPageToken;
    },
  });

  /**
   * Table Options
   */
  const isIncremental = resource.model?.spec?.incremental;
  const columns: ColumnDef<V1ModelSplit>[] = [
    {
      accessorKey: "data",
      header: "Data",
      cell: ({ row }) => flexRender(DataCell, { data: row.original.data }),
      meta: {
        widthPercent: 30,
      },
    },
    {
      accessorKey: "executedOn",
      header: "Executed on",
      cell: ({ row }) =>
        row.original.executedOn
          ? new Date(row.original.executedOn).toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
              second: "numeric",
              fractionalSecondDigits: 3,
            })
          : "-",
      meta: {
        widthPercent: 10,
      },
    },
    {
      accessorKey: "elapsedMs",
      header: "Elapsed time",
      cell: ({ row }) => row.original.elapsedMs + "ms",
      meta: {
        widthPercent: 10,
      },
    },
    {
      accessorKey: "watermark",
      header: "Watermark",
      cell: ({ row }) =>
        row.original.watermark
          ? new Date(row.original.watermark).toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
              second: "numeric",
              fractionalSecondDigits: 3,
            })
          : "-",
      meta: {
        widthPercent: 10,
      },
    },
    {
      accessorKey: "error",
      header: "Error",
      cell: ({ row }) => flexRender(ErrorCell, { error: row.original.error }),
      meta: {
        widthPercent: 25,
      },
    },
    ...(isIncremental
      ? [
          {
            accessorKey: "key",
            header: "",
            id: "actions",
            meta: {
              widthPercent: 10,
            },
            cell: ({ row }) =>
              flexRender(TriggerSplit, {
                splitKey: (row as Row<V1ModelSplit>).original.key as string,
              }),
          },
        ]
      : []),
  ];

  const options = writable<TableOptions<V1ModelSplit>>({
    data: [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });
  const table = createSvelteTable(options);
  $: ({ getHeaderGroups } = $table);

  // Update table when data changes
  let allRows: V1ModelSplit[] = [];
  $: {
    allRows =
      ($query.data &&
        $query.data.pages.flatMap((page) => page.splits as V1ModelSplit[])) ||
      [];

    options.update((old) => ({
      ...old,
      data: allRows,
    }));
  }

  /**
   * Virtualizer
   */
  const ROW_HEIGHT = 71;
  const OVERSCAN = 10;
  let virtualListEl: HTMLDivElement;
  $: rows = $table.getRowModel().rows;
  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });
  $: {
    $virtualizer.setOptions({
      count: $query.hasNextPage ? allRows.length + 1 : allRows.length,
    });
    const [lastItem] = [...$virtualizer.getVirtualItems()].reverse();
    if (
      lastItem &&
      lastItem.index > allRows.length - 1 &&
      $query.hasNextPage &&
      !$query.isFetchingNextPage
    ) {
      void $query.fetchNextPage();
    }
  }
  $: ({ getVirtualItems, getTotalSize } = $virtualizer);
</script>

<div class="scroll-container" bind:this={virtualListEl}>
  <div class="table-wrapper" style="height: {getTotalSize()}px;">
    <table>
      <thead>
        {#each getHeaderGroups() as headerGroup (headerGroup.id)}
          <tr>
            {#each headerGroup.headers as header (header.id)}
              <th
                colSpan={header.colSpan}
                style={`width: ${header.column.columnDef.meta?.widthPercent}%;`}
              >
                <svelte:component
                  this={flexRender(
                    header.column.columnDef.header,
                    header.getContext(),
                  )}
                />
              </th>
            {/each}
          </tr>
        {/each}
      </thead>
      <tbody>
        {#if allRows.length === 0}
          <tr>
            <td class="text-center h-16" colspan={columns.length}>
              <span class="text-gray-500">None</span>
            </td>
          </tr>
        {:else}
          {#each getVirtualItems() as virtualRow, idx (virtualRow.index)}
            <tr
              style="height: {virtualRow.size}px; transform: translateY({virtualRow.start -
                idx * virtualRow.size}px);"
            >
              {#each rows[virtualRow.index]?.getVisibleCells() ?? [] as cell (cell.id)}
                <td data-label={cell.column.columnDef.header}>
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
  .scroll-container {
    @apply h-[600px];
    @apply min-w-full;
    @apply overflow-auto;
  }

  .table-wrapper {
    @apply relative;
    @apply min-w-full;
  }

  table {
    @apply table-fixed min-w-full;
    @apply border-separate border-spacing-0;
  }
  table th,
  table td {
    @apply px-4 py-2;
    @apply border-b border-gray-200;
  }
  thead tr th {
    @apply border-t border-gray-200;
    @apply text-left font-semibold text-gray-500;
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
