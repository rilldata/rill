<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectTablesTable from "./ProjectTablesTable.svelte";
  import ProjectTablesRowCounts from "./ProjectTablesRowCounts.svelte";
  import { useTablesList, useTableMetadata } from "./selectors";

  let tablesList: any;
  let tableMetadata: any;
  let rowCounts: Map<string, number | "loading" | "error"> = new Map();

  $: ({ instanceId } = $runtime);

  $: tablesList = useTablesList(instanceId, "");
  $: tableMetadata = useTableMetadata(instanceId, "", $tablesList.data?.tables);
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Tables</h2>
  </div>

  {#if $tablesList.isLoading}
    <DelayedSpinner isLoading={$tablesList.isLoading} size="16px" />
  {:else if $tablesList.isError}
    <div class="text-red-500">
      Error loading tables: {$tablesList.error?.message}
    </div>
  {:else if $tablesList.data}
    <ProjectTablesRowCounts
      {instanceId}
      tables={$tablesList?.data?.tables ?? []}
      bind:rowCounts
    />
    <ProjectTablesTable
      tables={$tablesList?.data?.tables ?? []}
      columnCounts={$tableMetadata?.data?.columnCounts ?? new Map()}
      {rowCounts}
      isView={$tableMetadata?.data?.isView ?? new Map()}
    />
    {#if $tableMetadata?.isLoading}
      <div class="mt-2 text-xs text-gray-500">Loading table metadata...</div>
    {/if}
  {/if}
</section>
