<script lang="ts">
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { SortDirection } from "@rilldata/web-common/components/table-toolbar/types";
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

  interface PublicURLRow extends V1MagicAuthToken {
    dashboardTitle: string;
  }

  export let data: PublicURLRow[];
  export let onDelete: (deletedTokenId: string) => void;

  let searchText = "";
  let sortDirection: SortDirection = "newest";

  $: filteredData = data.filter((row) => {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    const label = (row.displayName || row.dashboardTitle || "").toLowerCase();
    const dashboard = (row.dashboardTitle || "").toLowerCase();
    const creator = String(row.attributes?.name || "").toLowerCase();
    return label.includes(q) || dashboard.includes(q) || creator.includes(q);
  });

  $: sortedData = [...filteredData].sort((a, b) => {
    const aTime = new Date(a.createdOn ?? 0).getTime();
    const bTime = new Date(b.createdOn ?? 0).getTime();
    return sortDirection === "newest" ? bTime - aTime : aTime - bTime;
  });

  function handleSortToggle() {
    sortDirection = sortDirection === "newest" ? "oldest" : "newest";
  }

  const columns: ColumnDef<PublicURLRow, any>[] = [
    {
      accessorKey: "displayName",
      header: "Label",
      cell: ({ row }) =>
        renderComponent(LabelCell, {
          displayName: row.original.displayName ?? "",
          dashboardTitle: row.original.dashboardTitle,
          url: row.original.url ?? "",
        }),
    },
    {
      accessorKey: "dashboardTitle",
      header: "Dashboard",
      enableSorting: false,
    },
    {
      accessorKey: "metricsViewFilters",
      header: "Filters",
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(FiltersCell, {
          metricsViewFilters: row.original.metricsViewFilters,
        }),
    },
    {
      accessorKey: "expiresOn",
      header: "Expires on",
      cell: ({ row }) =>
        renderComponent(DateCell, { value: row.original.expiresOn }),
    },
    {
      id: "createdBy",
      header: "Created by",
      accessorFn: (row) => row.attributes?.name || "—",
      enableSorting: false,
    },
    {
      accessorKey: "usedOn",
      header: "Last accessed",
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
  ];
</script>

<div class="flex flex-col gap-y-3 w-full">
  <TableToolbar
    {searchText}
    onSearchChange={(text) => (searchText = text)}
    searchDisabled={data.length === 0}
    {sortDirection}
    onSortToggle={handleSortToggle}
    showSort={false}
  />

  {#if sortedData.length === 0 && data.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16">
        <ResourceListEmptyState
          icon={ExternalLinkIcon}
          message="You don't have any public URLs yet"
        >
          <span slot="action">
            To create a public URL, click the Share button in a dashboard.
          </span>
        </ResourceListEmptyState>
      </div>
    </div>
  {:else if sortedData.length === 0}
    <div class="border rounded-lg bg-surface-background">
      <div class="text-center py-16 text-fg-secondary text-sm font-semibold">
        No public URLs match your search
      </div>
    </div>
  {:else}
    <BasicTable
      data={sortedData}
      {columns}
      columnLayout="minmax(150px, 1.5fr) minmax(120px, 1fr) minmax(120px, 1.5fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) 56px"
    />
  {/if}
</div>
