<script lang="ts">
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import { ExternalLinkIcon } from "lucide-svelte";
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import LabelCell from "./cells/LabelCell.svelte";
  import FiltersCell from "./cells/FiltersCell.svelte";
  import DateCell from "./cells/DateCell.svelte";
  import PublicURLsActionsRow from "./PublicURLsActionsRow.svelte";
  import { InMemoryRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  interface PublicURLRow extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  let {
    data,
    onDelete,
  }: { data: PublicURLRow[]; onDelete: (deletedTokenId: string) => void } =
    $props();

  const searchTextStore = new InMemoryRuneStore<string>("");

  let filteredData = $derived(
    data.filter((row) => {
      if (!searchTextStore.value) return true;
      const q = searchTextStore.value.toLowerCase();
      const label = (row.displayName || row.dashboardTitle || "").toLowerCase();
      const dashboard = (row.dashboardTitle || "").toLowerCase();
      const creator = String(row.attributes?.name || "").toLowerCase();
      return label.includes(q) || dashboard.includes(q) || creator.includes(q);
    }),
  );

  let columns = $derived([
    {
      accessorKey: "displayName",
      header: m.public_url_table_label_header(),
      cell: ({ row }) =>
        renderComponent(LabelCell, {
          displayName: row.original.displayName ?? "",
          dashboardTitle: row.original.dashboardTitle,
          url: row.original.url ?? "",
        }),
    },
    {
      accessorKey: "dashboardTitle",
      header: m.public_url_table_dashboard_header(),
      enableSorting: false,
    },
    {
      accessorKey: "metricsViewFilters",
      header: m.public_url_table_filters_header(),
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(FiltersCell, {
          metricsViewFilters: row.original.metricsViewFilters,
        }),
    },
    {
      accessorKey: "expiresOn",
      header: m.public_url_table_expires_header(),
      cell: ({ row }) =>
        renderComponent(DateCell, { value: row.original.expiresOn }),
    },
    {
      id: "createdBy",
      header: m.public_url_table_created_by_header(),
      accessorFn: (row) => row.attributes?.name || "—",
      enableSorting: false,
    },
    {
      accessorKey: "usedOn",
      header: m.public_url_table_last_accessed_header(),
      cell: ({ row }) =>
        renderComponent(DateCell, { value: row.original.usedOn }),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(PublicURLsActionsRow, {
          id: row.original.id ?? "",
          url: row.original.url ?? "",
          onDelete,
        }),
    },
  ] as ColumnDef<PublicURLRow, any>[]);
</script>

<div class="flex flex-col gap-y-3 w-full">
  <TableToolbar {searchTextStore} />

  <BasicTable
    data={filteredData}
    {columns}
    columnLayout="minmax(150px, 1.5fr) minmax(120px, 1fr) minmax(120px, 1.5fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) 56px"
  >
    <div slot="empty" class="text-center py-16">
      {#if data.length === 0}
        <ResourceListEmptyState
          icon={ExternalLinkIcon}
          message={m.public_url_no_urls_title()}
        >
          <span slot="action">
            {m.public_url_no_urls_empty()}
          </span>
        </ResourceListEmptyState>
      {:else}
        <span class="text-fg-secondary text-sm font-semibold">
          {m.public_url_no_match_search()}
        </span>
      {/if}
    </div>
  </BasicTable>
</div>
