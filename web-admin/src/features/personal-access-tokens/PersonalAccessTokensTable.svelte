<script lang="ts">
  import type { V1UserAuthToken } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import RevokeTokenCell from "./cells/RevokeTokenCell.svelte";
  import TokenDateCell from "./cells/TokenDateCell.svelte";
  import TokenExpiresCell from "./cells/TokenExpiresCell.svelte";
  import TokenPrefixCell from "./cells/TokenPrefixCell.svelte";
  import { getTokenLastUsedOn } from "./personal-access-tokens";

  let {
    tokens,
    onRevoke,
  }: {
    tokens: V1UserAuthToken[];
    onRevoke: (token: V1UserAuthToken) => void;
  } = $props();

  const columns: ColumnDef<V1UserAuthToken, any>[] = [
    {
      accessorKey: "displayName",
      header: m.common_name(),
    },
    {
      accessorKey: "prefix",
      header: m.access_tokens_column_token(),
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(TokenPrefixCell, { prefix: row.original.prefix }),
    },
    // Date columns sort by timestamp.
    {
      id: "createdOn",
      accessorFn: (token) => Date.parse(token.createdOn ?? ""),
      header: m.access_tokens_column_created(),
      cell: ({ row }) =>
        renderComponent(TokenDateCell, {
          value: row.original.createdOn,
          emptyLabel: "",
        }),
    },
    {
      id: "usedOn",
      // A token that was never used sorts as the least recently used.
      accessorFn: (token) => {
        const usedOn = getTokenLastUsedOn(token);
        return usedOn ? Date.parse(usedOn) : 0;
      },
      header: m.access_tokens_column_last_used(),
      cell: ({ row }) =>
        renderComponent(TokenDateCell, {
          value: getTokenLastUsedOn(row.original),
          emptyLabel: m.access_tokens_never_used(),
        }),
    },
    {
      id: "expiresOn",
      // A token that never expires sorts after every expiry date.
      accessorFn: (token) =>
        token.expiresOn ? Date.parse(token.expiresOn) : Infinity,
      header: m.access_tokens_column_expires(),
      cell: ({ row }) =>
        renderComponent(TokenExpiresCell, {
          expiresOn: row.original.expiresOn,
        }),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(RevokeTokenCell, {
          onRevoke: () => onRevoke(row.original),
        }),
    },
  ];
</script>

<BasicTable
  data={tokens}
  {columns}
  columnLayout="minmax(180px, 2fr) minmax(170px, 1.2fr) minmax(110px, 1fr) minmax(110px, 1fr) minmax(120px, 1fr) 96px"
/>
