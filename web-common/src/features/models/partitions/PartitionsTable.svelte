<script lang="ts">
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { type ColumnDef, renderComponent } from "tanstack-table-8-svelte-5";
  import {
    type V1ModelPartition,
    type V1Resource,
  } from "../../../runtime-client";
  import { createRuntimeServiceGetModelPartitionsInfinite } from "../../../runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import DataCell from "./DataCell.svelte";
  import ErrorCell from "./ErrorCell.svelte";
  import TriggerPartition from "./TriggerPartition.svelte";

  export let resource: V1Resource;
  export let whereErrored: boolean;
  export let wherePending: boolean;
  export let searchText = "";

  const runtimeClient = useRuntimeClient();
  const isIncremental = !!resource.model?.spec?.incremental;

  $: modelName = resource?.meta?.name?.name as string;

  $: baseParams = {
    ...(whereErrored ? { errored: true } : {}),
    ...(wherePending ? { pending: true } : {}),
  };
  $: query = createRuntimeServiceGetModelPartitionsInfinite(
    runtimeClient,
    { model: modelName, ...baseParams },
    {
      query: {
        enabled: !!modelName,
        refetchOnMount: true,
      },
    },
  );

  // Auto-fetch remaining pages so client-side search and sort cover the full set.
  $: if ($query.hasNextPage && !$query.isFetchingNextPage) {
    void $query.fetchNextPage();
  }

  $: allRows =
    $query.data?.pages.flatMap((p) => p.partitions as V1ModelPartition[]) ?? [];

  $: filteredRows = (() => {
    if (!searchText) return allRows;
    const q = searchText.toLowerCase();
    return allRows.filter((row) => {
      const key = (row.key ?? "").toLowerCase();
      const data = JSON.stringify(row.data ?? {}).toLowerCase();
      const error = (row.error ?? "").toLowerCase();
      return key.includes(q) || data.includes(q) || error.includes(q);
    });
  })();

  function formatTimestamp(ts: string | undefined) {
    return ts
      ? new Date(ts).toLocaleString(undefined, {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
          second: "numeric",
          fractionalSecondDigits: 3,
        })
      : "-";
  }

  const columns: ColumnDef<V1ModelPartition, any>[] = [
    {
      id: "data",
      header: "Data",
      accessorFn: (row) =>
        ((row.data?.uri as string) ?? row.key ?? "").toLowerCase(),
      cell: ({ row }) => renderComponent(DataCell, { data: row.original.data }),
    },
    {
      accessorKey: "executedOn",
      header: "Executed on",
      cell: ({ row }) => formatTimestamp(row.original.executedOn),
      sortDescFirst: true,
    },
    {
      accessorKey: "elapsedMs",
      header: "Elapsed time",
      cell: ({ row }) => (row.original.elapsedMs ?? 0) + "ms",
    },
    {
      accessorKey: "watermark",
      header: "Watermark",
      cell: ({ row }) => formatTimestamp(row.original.watermark),
    },
    {
      accessorKey: "error",
      header: "Error",
      cell: ({ row }) =>
        renderComponent(ErrorCell, { error: row.original.error }),
      enableSorting: false,
    },
    ...(isIncremental
      ? [
          {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row }) =>
              renderComponent(TriggerPartition, {
                partitionKey: row.original.key as string,
                resource,
              }),
          } satisfies ColumnDef<V1ModelPartition, any>,
        ]
      : []),
  ];

  const columnLayout = isIncremental
    ? "minmax(0, 3fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 2.5fr) minmax(180px, auto)"
    : "minmax(0, 3fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 2.5fr)";

  $: emptyText = $query.isLoading
    ? "Loading partitions…"
    : searchText
      ? "No partitions match your search"
      : "No partitions";
</script>

{#if $query.error}
  <div class="flex items-center justify-center h-16">
    <span class="text-red-500 font-semibold">
      Error: {$query.error.message}
    </span>
  </div>
{:else}
  <BasicTable data={filteredRows} {columns} {columnLayout} {emptyText} />
  {#if $query.isFetchingNextPage}
    <div class="flex items-center justify-center py-2">
      <span class="text-fg-secondary text-sm">Loading more…</span>
    </div>
  {/if}
{/if}
