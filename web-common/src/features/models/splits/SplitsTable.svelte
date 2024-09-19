<script lang="ts">
  import { createInfiniteQuery } from "@tanstack/svelte-query";
  import {
    ColumnDef,
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
    ...(whereErrored ? { whereErrored: true } : {}),
    ...(wherePending ? { wherePending: true } : {}),
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
      cell: (info) => (info.getValue() ? JSON.stringify(info.getValue()) : "-"),
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
    },
    {
      accessorKey: "elapsedMs",
      header: "Elapsed time",
      cell: ({ row }) => row.original.elapsedMs + "ms",
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
    },
    {
      accessorKey: "error",
      header: "Error",
      cell: ({ row }) => flexRender(ErrorCell, { error: row.original.error }),
    },
    ...(isIncremental
      ? [
          {
            accessorKey: "split",
            header: "",
            cell: ({ row }) =>
              flexRender(TriggerSplit, {
                resource,
                split: row.original,
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
  const ROW_HEIGHT = 50;
  const OVERSCAN = 5;
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

<div class="list scroll-container" bind:this={virtualListEl}>
  <div style="position: relative; height: {getTotalSize()}px;">
    <table class="w-full">
      <thead>
        {#each getHeaderGroups() as headerGroup (headerGroup.id)}
          <tr>
            {#each headerGroup.headers as header (header.id)}
              <th colSpan={header.colSpan} class="px-4 py-2 text-left">
                {#if !header.isPlaceholder}
                  <div
                    class="font-semibold text-gray-500 flex flex-row items-center gap-x-1"
                  >
                    <svelte:component
                      this={flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                    />
                  </div>
                {/if}
              </th>
            {/each}
          </tr>
        {/each}
      </thead>
      <tbody>
        {#if allRows.length === 0}
          <tr>
            <td class="text-center py-4" colspan={columns.length}>
              <div class="text-gray-500">None</div>
            </td>
          </tr>
        {:else}
          {#each getVirtualItems() as virtualRow, idx (virtualRow.index)}
            <tr
              style="height: {virtualRow.size}px; transform: translateY({virtualRow.start -
                idx * virtualRow.size}px);"
            >
              {#each rows[virtualRow.index]?.getVisibleCells() ?? [] as cell (cell.id)}
                <td
                  class={`px-4 py-2 ${cell.column.id === "actions" ? "w-1" : ""}`}
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
    height: 600px;
    width: 100%;
    overflow: auto;
  }
</style>
