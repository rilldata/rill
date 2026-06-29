<script lang="ts">
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

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

  $: columns = [
    {
      accessorKey: "title",
      header: m.public_url_table_label_header(),
      cell: ({ row }) =>
        renderComponent(DashboardLink, {
          href: row.original.url,
          title: row.original.displayName,
        }),
    },
    {
      accessorFn: (row) => row.dashboardTitle,
      header: m.public_url_table_dashboard_title_header(),
    },
    {
      accessorKey: "expiresOn",
      header: m.public_url_table_expires_header(),
      cell: (info) => {
        if (!info.getValue()) return "-";
        const date = formatDate(info.getValue() as string);
        return date;
      },
    },
    {
      accessorFn: (row) => row.attributes.name,
      header: m.public_url_table_created_by_header(),
    },
    {
      accessorKey: "usedOn",
      header: m.public_url_table_last_accessed_header(),
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
        renderComponent(PublicURLsActionsRow, {
          id: row.original.id,
          url: row.original.url,
          onDelete,
        }),
    },
  ] as ColumnDef<MagicAuthTokenProps, any>[];

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
/>
