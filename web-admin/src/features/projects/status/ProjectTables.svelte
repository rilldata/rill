<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import ProjectTablesTable from "./ProjectTablesTable.svelte";
  import { useTablesList, useTableMetadata } from "./selectors";

  $: ({ instanceId } = $runtime);

  $: tablesList = useTablesList(instanceId, "");

  // Filter out temporary tables (e.g., __rill_tmp_ prefixed tables)
  $: filteredTables =
    $tablesList.data?.tables?.filter(
      (t): t is V1OlapTableInfo =>
        !!t.name && !t.name.startsWith("__rill_tmp_"),
    ) ?? [];

  $: tableMetadata = useTableMetadata(instanceId, "", filteredTables);
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Tables</h2>
  </div>

  {#if $tablesList.isLoading}
    <div class="flex items-center gap-x-2 text-gray-500">
      <DelayedSpinner isLoading={true} size="16px" />
      <span class="text-sm">Loading tables...</span>
    </div>
  {:else if $tablesList.isError}
    <div class="text-red-500">
      Error loading tables: {$tablesList.error?.message}
    </div>
  {:else if filteredTables.length > 0}
    <ProjectTablesTable
      tables={filteredTables}
      isView={$tableMetadata?.data?.isView ?? new Map()}
    />
    {#if $tableMetadata?.isLoading}
      <div class="mt-2 text-xs text-gray-500">Loading table metadata...</div>
    {/if}
  {:else}
    <div class="text-gray-500 text-sm">No tables found</div>
  {/if}
</section>
