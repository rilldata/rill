<script lang="ts">
  import { createInfiniteQuery } from "@tanstack/svelte-query";
  import {
    type ColumnDef,
    type Row,
    type TableOptions,
    createSvelteTable,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
  } from "@tanstack/svelte-table";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { writable } from "svelte/store";
  import {
    type RpcStatus,
    type V1GetModelPartitionsResponse,
    type V1ModelPartition,
    type V1Resource,
    getRuntimeServiceGetModelPartitionsQueryKey,
    runtimeServiceGetModelPartitions,
  } from "../../../runtime-client";
  import type { ErrorType } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import DataCell from "./DataCell.svelte";
  import ErrorCell from "./ErrorCell.svelte";
  import TriggerPartition from "./TriggerPartition.svelte";

  export let resource: V1Resource;
  export let whereErrored: boolean;
  export let wherePending: boolean;

  $: modelName = resource?.meta?.name?.name as string;
  $: ({ instanceId } = $runtime);

  // ==========================
  // Infinite Query
  // ==========================
  $: baseParams = {
    ...(whereErrored ? { errored: true } : {}),
    ...(wherePending ? { pending: true } : {}),
  };
  $: query = createInfiniteQuery<
    V1GetModelPartitionsResponse,
    ErrorType<RpcStatus>
  >({
    queryKey: getRuntimeServiceGetModelPartitionsQueryKey(
      instanceId,
      modelName,
      baseParams,
    ),
    queryFn: ({ pageParam }) => {
      const getModelPartitionsParams = {
        ...baseParams,
        ...(pageParam
          ? {
              pageToken: pageParam as string,
            }
          : {}),
      };
      return runtimeServiceGetModelPartitions(
        instanceId,
        modelName,
        getModelPartitionsParams,
      );
    },
    enabled: !!modelName,
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken || lastPage.nextPageToken === "") {
        return undefined;
      }
      return lastPage.nextPageToken;
    },
    refetchOnMount: true,
  });
  $: ({ error } = $query);

  // ==========================
  // Table Options
  // ==========================
  const isIncremental = resource.model?.spec?.incremental;

  const columns: ColumnDef<V1ModelPartition>[] = [
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
              flexRender(TriggerPartition, {
                partitionKey: (row as Row<V1ModelPartition>).original
                  .key as string,
              }),
          },
        ]
      : []),
  ];

  const options = writable<TableOptions<V1ModelPartition>>({
    data: [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });
  const table = createSvelteTable(options);
  $: ({ getHeaderGroups } = $table);

  // Sync table data with query data
  let allRows: V1ModelPartition[] = [];
  $: {
    allRows =
      ($query.data &&
        $query.data.pages.flatMap(
          (page) => page.partitions as V1ModelPartition[],
        )) ||
      [];

    options.update((old) => ({
      ...old,
      data: allRows,
    }));
  }
  $: rows = $table.getRowModel().rows;

  // ==========================
  // Virtualizer
  // ==========================
  const ROW_HEIGHT = 71;
  const OVERSCAN = 10;

  let virtualListEl: HTMLDivElement;

  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => virtualListEl,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });
  $: ({
    getVirtualItems,
    getTotalSize,
    setOptions,
    options: virtualizerOptions,
  } = $virtualizer);
  $: virtualRows = getVirtualItems();

  // Logic for infinite scroll
  $: {
    setOptions({
      count: $query.hasNextPage ? allRows.length + 1 : allRows.length,
    });
    const [lastItem] = [...virtualRows].reverse();
    if (
      lastItem &&
      lastItem.index > allRows.length - 1 &&
      $query.hasNextPage &&
      !$query.isFetchingNextPage
    ) {
      void $query.fetchNextPage();
    }
  }

  // Positioning strategy from https://github.com/TanStack/virtual/issues/585#issuecomment-1716173260
  // (Required for sticky header)
  $: [paddingTop, paddingBottom] =
    virtualRows.length > 0
      ? [
          Math.max(0, virtualRows[0].start - virtualizerOptions.scrollMargin),
          Math.max(0, getTotalSize() - virtualRows[virtualRows.length - 1].end),
        ]
      : [0, 0];

  // Scroll to top when filter changes
  $: {
    if (virtualListEl && baseParams) {
      virtualListEl.scrollTo({ top: 0 });
    }
  }
</script>

<div class="scroll-container" bind:this={virtualListEl}>
  <div class="table-wrapper">
    <table>
      <thead>
        {#each getHeaderGroups() as headerGroup (headerGroup.id)}
          <tr>
            {#each headerGroup.headers as header (header.id)}
              {@const widthPercent = header.column.columnDef.meta?.widthPercent}
              <th colSpan={header.colSpan} style={`width: ${widthPercent}%;`}>
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
        {#if error}
          <tr>
            <td class="text-center h-16" colspan={columns.length}>
              <span class="text-red-500 font-semibold"
                >Error: {error.message}</span
              >
            </td>
          </tr>
        {:else if allRows.length === 0}
          <tr>
            <td class="text-center h-16" colspan={columns.length}>
              <span class="text-gray-500">None</span>
            </td>
          </tr>
        {:else}
          <tr style:height="{paddingTop}px" />
          {#each getVirtualItems() as virtualRow (virtualRow.index)}
            <tr>
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
          <tr style:height="{paddingBottom}px" />
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
    @apply sticky top-0 z-10 bg-white;
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
