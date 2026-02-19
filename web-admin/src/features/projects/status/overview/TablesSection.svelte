<script lang="ts">
  import { page } from "$app/stores";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useInfiniteTablesList, useTableMetadata } from "../selectors";
  import { filterTemporaryTables } from "../tables/utils";
  import { writable } from "svelte/store";

  const queryClient = useQueryClient();

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Get instance info for OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;

  // Get tables list (first page only; show "+" suffix when more pages exist)
  $: connectorName = instance?.olapConnector ?? "";
  const tablesParams = writable({
    instanceId: "",
    connector: "",
    searchPattern: undefined as string | undefined,
  });
  $: tablesParams.set({ instanceId, connector: connectorName });
  const tablesList = useInfiniteTablesList(tablesParams);

  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);
  $: hasMore = $tablesList.hasNextPage;
  $: tableMetadata = useTableMetadata(
    instanceId,
    connectorName,
    filteredTables,
    queryClient,
  );

  // Count tables vs views
  $: viewCount = Array.from(
    $tableMetadata?.data?.isView?.values() ?? [],
  ).filter(Boolean).length;
  $: tableCount = filteredTables.length - viewCount;
  $: isLoading = $tablesList.isLoading || $tableMetadata?.isLoading;
</script>

<section class="section">
  <div class="section-header">
    <h3 class="section-title">Tables</h3>
    <a href="{basePage}/tables" class="view-all">View all</a>
  </div>
  {#if isLoading}
    <p class="text-sm text-fg-secondary">Loading tables...</p>
  {:else if filteredTables.length > 0}
    <div class="table-chips">
      <a href="{basePage}/tables?type=table" class="table-chip">
        <span class="font-medium">{tableCount}{hasMore ? "+" : ""}</span>
        <span class="text-fg-secondary"
          >{tableCount === 1 ? "Table" : "Tables"}</span
        >
      </a>
      <a href="{basePage}/tables?type=view" class="table-chip">
        <span class="font-medium">{viewCount}{hasMore ? "+" : ""}</span>
        <span class="text-fg-secondary"
          >{viewCount === 1 ? "View" : "Views"}</span
        >
      </a>
    </div>
  {:else}
    <p class="text-sm text-fg-secondary">No tables found.</p>
  {/if}
</section>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .view-all {
    @apply text-xs text-primary-500 no-underline;
  }
  .view-all:hover {
    @apply text-primary-600;
  }
  .table-chips {
    @apply flex flex-wrap gap-2;
  }
  .table-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit;
  }
  .table-chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
