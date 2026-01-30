<script lang="ts">
  import {
    useTablesList,
    useTableMetadata,
  } from "@rilldata/web-admin/features/projects/status/selectors";
  import { filterTemporaryTables } from "@rilldata/web-admin/features/projects/status/model-overview/utils";
  import ProjectTables from "@rilldata/web-admin/features/projects/status/model-overview/ProjectTables.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { formatConnectorName } from "@rilldata/web-admin/features/projects/status/display-utils";

  $: ({ instanceId } = $runtime);

  // Get instance info for OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(
    instanceId,
    { sensitive: true },
    { query: { enabled: !!instanceId } },
  );
  $: instance = $instanceQuery.data?.instance;
  $: olapConnectorName = instance?.olapConnector;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === olapConnectorName,
  );

  // Get tables list
  $: tablesList = useTablesList(instanceId, "");
  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);

  // Get table metadata to determine views vs tables
  $: tableMetadata = useTableMetadata(instanceId, "", filteredTables);

  // Count tables and views
  $: viewCount = Array.from(
    $tableMetadata?.data?.isView?.values() ?? [],
  ).filter(Boolean).length;
  $: tableCount = filteredTables.length - viewCount;
</script>

<div class="flex flex-col gap-y-6 size-full">
  <section class="flex flex-col gap-y-4">
    <h2 class="text-lg font-medium">Model Overview</h2>
    <div class="grid grid-cols-3 gap-4">
      <div class="flex flex-col gap-y-1 p-4 border rounded-md bg-surface">
        <div class="flex items-center gap-x-1">
          <span class="text-sm text-fg-secondary"
            >Tables (Materialized Models)</span
          >
          <a
            href="https://docs.rilldata.com/build/models/performance#consider-which-models-to-materialize"
            target="_blank"
            rel="noopener noreferrer"
            class="text-fg-muted hover:text-fg-secondary"
            title="Learn about materialized models"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 16 16"
              fill="currentColor"
              class="w-3.5 h-3.5"
            >
              <path
                fill-rule="evenodd"
                d="M15 8A7 7 0 1 1 1 8a7 7 0 0 1 14 0ZM9 5a1 1 0 1 1-2 0 1 1 0 0 1 2 0ZM6.75 8a.75.75 0 0 0 0 1.5h.75v1.75a.75.75 0 0 0 1.5 0v-2.5A.75.75 0 0 0 8.25 8h-1.5Z"
                clip-rule="evenodd"
              />
            </svg>
          </a>
        </div>
        <span class="text-2xl font-semibold tabular-nums">
          {$tableMetadata?.isLoading ? "-" : tableCount}
        </span>
      </div>
      <div class="flex flex-col gap-y-1 p-4 border rounded-md bg-surface">
        <span class="text-sm text-fg-secondary">Views</span>
        <span class="text-2xl font-semibold tabular-nums">
          {$tableMetadata?.isLoading ? "-" : viewCount}
        </span>
      </div>
      <div class="flex flex-col gap-y-1 p-4 border rounded-md bg-surface">
        <div class="flex items-center gap-x-1">
          <span class="text-sm text-fg-secondary">OLAP Engine</span>
          <a
            href="https://docs.rilldata.com/reference/olap-engines"
            target="_blank"
            rel="noopener noreferrer"
            class="text-fg-muted hover:text-fg-secondary"
            title="Learn about OLAP engines"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 16 16"
              fill="currentColor"
              class="w-3.5 h-3.5"
            >
              <path
                fill-rule="evenodd"
                d="M15 8A7 7 0 1 1 1 8a7 7 0 0 1 14 0ZM9 5a1 1 0 1 1-2 0 1 1 0 0 1 2 0ZM6.75 8a.75.75 0 0 0 0 1.5h.75v1.75a.75.75 0 0 0 1.5 0v-2.5A.75.75 0 0 0 8.25 8h-1.5Z"
                clip-rule="evenodd"
              />
            </svg>
          </a>
        </div>
        {#if $instanceQuery.isLoading}
          <span class="text-2xl font-semibold">-</span>
        {:else if olapConnector}
          <span class="text-sm font-medium text-gray-900">
            {formatConnectorName(olapConnector.type)}
          </span>
          <div class="flex items-center gap-2 text-xs text-gray-600">
            <span class="text-xs text-gray-600">
              {olapConnector.provision ? "Rill-Managed" : "Self-Managed"}
            </span>
          </div>
        {:else}
          <span class="text-sm font-medium text-gray-900">-</span>
        {/if}
      </div>
    </div>

    <ProjectTables />
  </section>
</div>
