<script lang="ts">
  import { page } from "$app/stores";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useTablesList, useTableMetadata } from "../selectors";
  import { filterTemporaryTables } from "../tables/utils";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Get instance info for OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: isDuckDB = !olapConnector || olapConnector.type === "duckdb";

  // Get tables list and metadata
  $: tablesList = useTablesList(instanceId, "");
  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);
  $: tableMetadata = useTableMetadata(instanceId, "", filteredTables);

  // Count tables vs views
  $: viewCount = Array.from(
    $tableMetadata?.data?.isView?.values() ?? [],
  ).filter(Boolean).length;
  $: tableCount = filteredTables.length - viewCount;
  $: isLoading = $tablesList.isLoading || $tableMetadata?.isLoading;
</script>

{#if !isLoading && filteredTables.length > 0}
  <section class="section">
    <div class="section-header">
      <h3 class="section-title">Tables</h3>
      <a href="{basePage}/tables" class="view-all">View all</a>
    </div>
    <div class="table-chips">
      <a href="{basePage}/tables" class="table-chip">
        <span class="font-medium">{tableCount}</span>
        <span class="text-fg-secondary"
          >{tableCount === 1 ? "Table" : "Tables"}</span
        >
      </a>
      {#if isDuckDB}
        <a href="{basePage}/tables" class="table-chip">
          <span class="font-medium">{viewCount}</span>
          <span class="text-fg-secondary"
            >{viewCount === 1 ? "View" : "Views"}</span
          >
        </a>
      {/if}
    </div>
  </section>
{/if}

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
