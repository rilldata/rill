<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { FilterGroup } from "@rilldata/web-common/components/table-toolbar/types";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetInstance,
    getRuntimeServiceListResourcesQueryKey,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";
  import ModelsTable from "@rilldata/web-common/features/projects/status/tables/ModelsTable.svelte";
  import ExternalTablesTable from "@rilldata/web-common/features/projects/status/tables/ExternalTablesTable.svelte";
  import { useInfiniteTablesList, useModelResources } from "../selectors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    filterTemporaryTables,
    applyTableFilters,
    splitTablesByModel,
  } from "@rilldata/web-common/features/projects/status/tables/utils";
  import ResourceSpecDialog from "@rilldata/web-common/features/projects/status/ResourceSpecDialog.svelte";
  import ModelPartitionsDialog from "@rilldata/web-common/features/projects/status/tables/ModelPartitionsDialog.svelte";
  import RefreshErroredPartitionsDialog from "@rilldata/web-common/features/projects/status/tables/RefreshErroredPartitionsDialog.svelte";
  import RefreshResourceConfirmDialog from "@rilldata/web-common/features/projects/status/RefreshResourceConfirmDialog.svelte";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { onMount } from "svelte";

  const runtimeClient = useRuntimeClient();

  // OLAP connector info
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: connectorName = instance?.olapConnector ?? "";

  // Filters — initialized from URL params (type is multi-select array)
  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "type", type: "array" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));

  // Debounce search for server-side filtering
  let debouncedSearch = searchText;
  const updateDebouncedSearch = debounce((text: string) => {
    debouncedSearch = text;
  }, 300);
  $: updateDebouncedSearch(searchText);

  $: searchPattern = debouncedSearch ? `%${debouncedSearch}%` : undefined;

  // Use a writable store so createInfiniteQuery is called once during init;
  // parameter changes flow reactively through the store.
  const tablesParams = writable({
    client: runtimeClient,
    connector: "",
    searchPattern: undefined as string | undefined,
  });
  $: tablesParams.set({
    client: runtimeClient,
    connector: connectorName,
    searchPattern,
  });
  const tablesList = useInfiniteTablesList(tablesParams);

  // Filter out temporary tables (e.g., __rill_tmp_ prefixed tables)
  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);

  // TODO: populate from OLAPGetTable responses when per-table metadata is available
  let isViewMap = new Map<string, boolean>();
  // createQuery (unlike createInfiniteQuery) handles re-creation in $: blocks safely
  $: modelResourcesQuery = useModelResources(runtimeClient);
  $: modelResources = $modelResourcesQuery.data ?? new Map();
  let typeFilter: string[] = parseArrayParam(
    $page.url.searchParams.get("type"),
  );
  let mounted = false;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    searchText = parseStringParam($page.url.searchParams.get("q"));
    typeFilter = parseArrayParam($page.url.searchParams.get("type"));
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({ q: searchText, type: typeFilter });
  }

  onMount(() => {
    mounted = true;
  });

  $: filterGroups = [
    {
      label: "Type",
      key: "type",
      options: [
        { label: "Table", value: "table" },
        { label: "View", value: "view" },
      ],
      selected: typeFilter,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  // Split once on unfiltered tables, then apply type filter per section
  $: ({ modelTables: allModelTables, externalTables: allExternalTables } =
    splitTablesByModel(filteredTables, modelResources));
  $: modelTables = applyTableFilters(allModelTables, typeFilter, isViewMap);
  $: externalTables = applyTableFilters(
    allExternalTables,
    typeFilter,
    isViewMap,
  );

  // Dialog states
  let specDialogOpen = false;
  let specResourceName = "";
  let specResourceKind = "";
  let specResource: V1Resource | undefined = undefined;

  let partitionsDialogOpen = false;
  let erroredPartitionsDialogOpen = false;
  let incrementalRefreshDialogOpen = false;
  let fullRefreshDialogOpen = false;

  let selectedResource: V1Resource | null = null;
  let selectedModelName = "";

  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);
  const queryClient = useQueryClient();

  // Handlers
  function handleModelInfoClick(resource: V1Resource) {
    specResourceName = resource.meta?.name?.name ?? "";
    specResourceKind = resource.meta?.name?.kind ?? "";
    specResource = resource;
    specDialogOpen = true;
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

  function handleViewLogsClick(name: string) {
    const basePath = $page.url.pathname.replace(/\/tables\/?$/, "");
    void goto(`${basePath}/logs?q=${encodeURIComponent(name)}`);
  }

  async function refreshModel(opts: {
    full?: boolean;
    allErroredPartitions?: boolean;
  }) {
    if (!selectedModelName) return;

    try {
      await $createTrigger.mutateAsync({
        models: [{ model: selectedModelName, ...opts }],
      });

      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    } catch (error) {
      console.error("Failed to refresh model:", error);
      throw error;
    }
  }

  const handleRefreshErrored = () =>
    refreshModel({ allErroredPartitions: true });
  const handleIncrementalRefresh = () => refreshModel({});
  const handleFullRefresh = () => refreshModel({ full: true });
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Tables</h2>
  </div>

  <TableToolbar
    {searchText}
    onSearchChange={(text) => {
      searchText = text;
    }}
    {filterGroups}
    onFilterChange={(key, value) => {
      if (key === "type") {
        typeFilter = typeFilter.includes(value)
          ? typeFilter.filter((v) => v !== value)
          : [...typeFilter, value];
      }
    }}
    onClearAllFilters={() => {
      typeFilter = [];
      searchText = "";
    }}
    showSort={false}
  />

  {#if $tablesList.isError}
    <div class="text-red-500">
      Error loading tables: {$tablesList.error?.message}
    </div>
  {:else}
    {@const isLoading =
      $instanceQuery.isLoading || !connectorName || $tablesList.isLoading}

    <!-- Models section -->
    <section class="flex flex-col gap-y-2">
      <h3 class="text-sm font-semibold text-fg-primary">
        Models{isLoading
          ? ""
          : ` (${modelTables.length}${$tablesList.hasNextPage ? "+" : ""})`}
      </h3>
      {#if isLoading}
        <div
          class="border border-border rounded-sm py-10 flex flex-col items-center gap-y-2 text-fg-secondary"
        >
          <DelayedSpinner isLoading={true} size="20px" />
          <span class="text-sm">Loading models</span>
        </div>
      {:else if modelTables.length > 0}
        <ModelsTable
          tables={modelTables}
          isView={isViewMap}
          {modelResources}
          onModelInfoClick={handleModelInfoClick}
          onViewPartitionsClick={handleViewPartitionsClick}
          onRefreshErroredClick={handleRefreshErroredClick}
          onIncrementalRefreshClick={handleIncrementalRefreshClick}
          onFullRefreshClick={handleFullRefreshClick}
          onViewLogsClick={handleViewLogsClick}
        />
      {:else}
        <div
          class="border border-border rounded-sm py-10 flex flex-col items-center gap-y-1"
        >
          {#if allModelTables.length > 0}
            <span class="text-fg-secondary font-semibold text-sm">
              No models match the current filters
            </span>
          {:else}
            <span class="text-fg-secondary font-semibold text-sm">
              No models
            </span>
            <span class="text-fg-muted text-sm">
              Models are created in Rill Developer.
              <a
                href="https://docs.rilldata.com/build/models/"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary-500 hover:text-primary-600"
              >
                Learn more
              </a>
            </span>
          {/if}
        </div>
      {/if}
    </section>

    <!-- External Tables section -->
    <section class="flex flex-col gap-y-2">
      <h3 class="text-sm font-semibold text-fg-primary">
        External Tables{isLoading
          ? ""
          : ` (${externalTables.length}${$tablesList.hasNextPage ? "+" : ""})`}
      </h3>
      {#if isLoading}
        <div
          class="border border-border rounded-sm py-10 flex flex-col items-center gap-y-2 text-fg-secondary"
        >
          <DelayedSpinner isLoading={true} size="20px" />
          <span class="text-sm">Loading tables</span>
        </div>
      {:else if externalTables.length > 0}
        <ExternalTablesTable tables={externalTables} isView={isViewMap} />
      {:else}
        <div
          class="border border-border rounded-sm py-10 flex flex-col items-center gap-y-1"
        >
          {#if allExternalTables.length > 0}
            <span class="text-fg-secondary font-semibold text-sm">
              No external tables match the current filters
            </span>
          {:else}
            <span class="text-fg-secondary font-semibold text-sm">
              No external tables
            </span>
            <span class="text-fg-muted text-sm">
              <a
                href="https://docs.rilldata.com/developers/build/connectors/olap"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary-500 hover:text-primary-600"
              >
                Learn about connecting external OLAP engines
              </a>
            </span>
          {/if}
        </div>
      {/if}
    </section>

    {#if $tablesList.hasNextPage}
      <div class="flex justify-center">
        <Button
          type="tertiary"
          onClick={() => $tablesList.fetchNextPage()}
          disabled={$tablesList.isFetchingNextPage}
          loading={$tablesList.isFetchingNextPage}
          loadingCopy="Loading..."
        >
          Load more tables
        </Button>
      </div>
    {/if}
  {/if}
</section>

<ResourceSpecDialog
  bind:open={specDialogOpen}
  resourceName={specResourceName}
  resourceKind={specResourceKind}
  resource={specResource}
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
