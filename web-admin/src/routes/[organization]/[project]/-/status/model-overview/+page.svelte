<script lang="ts">
  import {
    useTablesList,
    useTableMetadata,
  } from "@rilldata/web-admin/features/projects/status/selectors";
  import ProjectTables from "@rilldata/web-admin/features/projects/status/ProjectTables.svelte";
  import {
    createRuntimeServiceGetInstance,
    type V1OlapTableInfo,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  // Get instance info for OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(
    instanceId,
    { sensitive: true },
    { query: { enabled: !!instanceId } },
  );
  $: olapEngine = $instanceQuery.data?.instance?.olapConnector || "-";

  // Get tables list
  $: tablesList = useTablesList(instanceId, "");
  $: filteredTables =
    $tablesList.data?.tables?.filter(
      (t): t is V1OlapTableInfo =>
        !!t.name && !t.name.startsWith("__rill_tmp_"),
    ) ?? [];

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
    <h2 class="text-lg font-medium">Model Details</h2>
    <div class="grid grid-cols-3 gap-4">
    <div class="flex flex-col gap-y-1 p-4 border rounded-md">
      <div class="flex items-center gap-x-1">
        <span class="text-sm text-gray-500">Tables (Materialized Models)</span>
        <a
          href="https://docs.rilldata.com/build/models/performance#consider-which-models-to-materialize"
          target="_blank"
          rel="noopener noreferrer"
          class="text-gray-400 hover:text-gray-600"
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
    <div class="flex flex-col gap-y-1 p-4 border rounded-md">
      <span class="text-sm text-gray-500">Views</span>
      <span class="text-2xl font-semibold tabular-nums">
        {$tableMetadata?.isLoading ? "-" : viewCount}
      </span>
    </div>
    <div class="flex flex-col gap-y-1 p-4 border rounded-md">
      <div class="flex items-center gap-x-1">
        <span class="text-sm text-gray-500">OLAP Engine</span>
        <a
          href="https://docs.rilldata.com/reference/olap-engines"
          target="_blank"
          rel="noopener noreferrer"
          class="text-gray-400 hover:text-gray-600"
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
      <span class="text-2xl font-semibold">
        {$instanceQuery.isLoading ? "-" : olapEngine}
      </span>
    </div>
  </div>

  <ProjectTables />
</div>
