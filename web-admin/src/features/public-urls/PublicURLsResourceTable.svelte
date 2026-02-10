<script lang="ts">
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import { ExternalLinkIcon } from "lucide-svelte";
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import PublicURLsCompositeCell from "./PublicURLsCompositeCell.svelte";

  interface PublicURLRow extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  export let data: PublicURLRow[];
  export let onDelete: (deletedTokenId: string) => void;

  const columns: ColumnDef<PublicURLRow, string>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(PublicURLsCompositeCell, {
          url: info.row.original.url,
          displayName: info.row.original.displayName,
          dashboardTitle: info.row.original.dashboardTitle,
          createdBy: info.row.original.attributes?.name ?? "",
          expiresOn: info.row.original.expiresOn,
          id: info.row.original.id,
          metricsViewFilters: info.row.original.metricsViewFilters,
          onDelete,
        }),
    },
    {
      id: "name",
      accessorFn: (row) => row.displayName || row.dashboardTitle || "",
    },
  ];

  const columnVisibility = {
    name: false,
  };
</script>

<ResourceList {columns} {data} {columnVisibility} kind="public URL" fixedRowHeight={false}>
  <ResourceListEmptyState
    slot="empty"
    icon={ExternalLinkIcon}
    message="You don't have any public URLs yet"
  >
    <span slot="action">
      To create a public URL, click the Share button in a dashboard.
    </span>
  </ResourceListEmptyState>
</ResourceList>
