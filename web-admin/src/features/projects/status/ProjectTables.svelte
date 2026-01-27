<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    type V1OlapTableInfo,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ProjectTablesTable from "./ProjectTablesTable.svelte";
  import {
    useTablesList,
    useTableMetadata,
    useTableCardinality,
    useModelResources,
  } from "./selectors";
  import ModelInfoDialog from "./ModelInfoDialog.svelte";
  import ModelLogsPanel from "./ModelLogsPanel.svelte";
  import ModelPartitionsDialog from "./ModelPartitionsDialog.svelte";
  import RefreshErroredPartitionsDialog from "./RefreshErroredPartitionsDialog.svelte";
  import RefreshResourceConfirmDialog from "./RefreshResourceConfirmDialog.svelte";

  $: ({ instanceId } = $runtime);

  $: tablesList = useTablesList(instanceId, "");

  // Filter out temporary tables (e.g., __rill_tmp_ prefixed tables)
  $: filteredTables =
    $tablesList.data?.tables?.filter(
      (t): t is V1OlapTableInfo =>
        !!t.name && !t.name.startsWith("__rill_tmp_"),
    ) ?? [];

  $: tableMetadata = useTableMetadata(instanceId, "", filteredTables);
  $: tableCardinality = useTableCardinality(instanceId, filteredTables);
  $: modelResourcesQuery = useModelResources(instanceId);
  $: modelResources = $modelResourcesQuery.data ?? new Map();

  // Dialog states
  let modelInfoDialogOpen = false;
  let partitionsDialogOpen = false;
  let erroredPartitionsDialogOpen = false;
  let incrementalRefreshDialogOpen = false;
  let fullRefreshDialogOpen = false;

  let selectedResource: V1Resource | null = null;
  let selectedModelName = "";

  const createTrigger = createRuntimeServiceCreateTrigger();
  const queryClient = useQueryClient();

  // Handlers
  function handleModelInfoClick(resource: V1Resource) {
    selectedResource = resource;
    modelInfoDialogOpen = true;
  }

  function handleViewPartitionsClick(resource: V1Resource) {
    selectedResource = resource;
    partitionsDialogOpen = true;
  }

  function handleRefreshErroredClick(resource: V1Resource) {
    selectedResource = resource;
    selectedModelName = resource.meta?.name?.name ?? "";
    erroredPartitionsDialogOpen = true;
  }

  function handleIncrementalRefreshClick(resource: V1Resource) {
    selectedModelName = resource.meta?.name?.name ?? "";
    incrementalRefreshDialogOpen = true;
  }

  function handleFullRefreshClick(resource: V1Resource) {
    selectedModelName = resource.meta?.name?.name ?? "";
    fullRefreshDialogOpen = true;
  }

  async function handleRefreshErrored() {
    if (!selectedModelName) return;

    await $createTrigger.mutateAsync({
      instanceId,
      data: {
        models: [{ model: selectedModelName, allErroredPartitions: true }],
      },
    });

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    });
  }

  async function handleIncrementalRefresh() {
    if (!selectedModelName) return;

    await $createTrigger.mutateAsync({
      instanceId,
      data: {
        models: [{ model: selectedModelName }],
      },
    });

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    });
  }

  async function handleFullRefresh() {
    if (!selectedModelName) return;

    await $createTrigger.mutateAsync({
      instanceId,
      data: {
        models: [{ model: selectedModelName, full: true }],
      },
    });

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    });
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Model Details</h2>
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
      columnCount={$tableMetadata?.data?.columnCount ?? new Map()}
      rowCount={$tableCardinality?.data?.rowCount ?? new Map()}
      {modelResources}
      onModelInfoClick={handleModelInfoClick}
      onViewPartitionsClick={handleViewPartitionsClick}
      onRefreshErroredClick={handleRefreshErroredClick}
      onIncrementalRefreshClick={handleIncrementalRefreshClick}
      onFullRefreshClick={handleFullRefreshClick}
    />
    {#if $tableMetadata?.isLoading || $tableCardinality?.isLoading}
      <div class="mt-2 text-xs text-gray-500">Loading table metadata...</div>
    {/if}
  {:else}
    <div class="text-gray-500 text-sm">No tables found</div>
  {/if}
</section>

<ModelLogsPanel />

<ModelInfoDialog
  bind:open={modelInfoDialogOpen}
  resource={selectedResource}
  onClose={() => {
    modelInfoDialogOpen = false;
  }}
/>

<ModelPartitionsDialog
  bind:open={partitionsDialogOpen}
  resource={selectedResource}
  onClose={() => {
    partitionsDialogOpen = false;
  }}
/>

<RefreshErroredPartitionsDialog
  bind:open={erroredPartitionsDialogOpen}
  modelName={selectedModelName}
  onRefresh={handleRefreshErrored}
/>

<RefreshResourceConfirmDialog
  bind:open={incrementalRefreshDialogOpen}
  name={selectedModelName}
  refreshType="incremental"
  onRefresh={handleIncrementalRefresh}
/>

<RefreshResourceConfirmDialog
  bind:open={fullRefreshDialogOpen}
  name={selectedModelName}
  refreshType="full"
  onRefresh={handleFullRefresh}
/>
