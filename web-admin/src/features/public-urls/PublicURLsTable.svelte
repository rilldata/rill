<script lang="ts">
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import PublicURLsActionsRow from "./PublicURLsActionsRow.svelte";
  import DashboardLink from "./DashboardLink.svelte";
  import type {
    V1MagicAuthToken,
    RpcStatus,
    V1ListMagicAuthTokensResponse,
  } from "@rilldata/web-admin/client";
  import InfiniteScrollTable from "@rilldata/web-common/components/table/InfiniteScrollTable.svelte";
  import type {
    InfiniteData,
    InfiniteQueryObserverResult,
  } from "@tanstack/svelte-query";

  interface MagicAuthTokenProps extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  export let data: MagicAuthTokenProps[];
  export let query: InfiniteQueryObserverResult<
    InfiniteData<V1ListMagicAuthTokensResponse, unknown>,
    RpcStatus
  >;
  export let onDelete: (deletedTokenId: string) => void;

  $: safeData = Array.isArray(data) ? data : [];

  $: dynamicTableMaxHeight =
    safeData.length > 12 ? `calc(100dvh - 300px)` : "auto";

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
      sortDescFirst: true,
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

  function formatDate(value: string) {
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }
</script>

<InfiniteScrollTable
  data={safeData}
  {columns}
  hasNextPage={query.hasNextPage}
  isFetchingNextPage={query.isFetchingNextPage}
  onLoadMore={() => query.fetchNextPage()}
  maxHeight={dynamicTableMaxHeight}
  rowHeight={40}
/>
