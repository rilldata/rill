<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceGetInstance,
    getRuntimeServiceListResourcesQueryKey,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";
  import ModelsTable from "./ModelsTable.svelte";
  import ExternalTablesTable from "./ExternalTablesTable.svelte";
  import { useInfiniteTablesList, useModelResources } from "../selectors";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    filterTemporaryTables,
    applyTableFilters,
    splitTablesByModel,
  } from "./utils";
  import ResourceSpecDialog from "../resource-table/ResourceSpecDialog.svelte";
  import ModelPartitionsDialog from "./ModelPartitionsDialog.svelte";
  import RefreshErroredPartitionsDialog from "./RefreshErroredPartitionsDialog.svelte";
  import RefreshResourceConfirmDialog from "../resource-table/RefreshResourceConfirmDialog.svelte";
  import {
    createUrlFilterSync,
    parseEnumParam,
    parseStringParam,
  } from "../url-filter-sync";
  import { onMount } from "svelte";

  $: ({ instanceId } = $runtime);

  // OLAP connector info
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: connectorName = instance?.olapConnector ?? "";

  // Filters — initialized from URL params
  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "type", type: "enum", defaultValue: "all" },
  ]);
  filterSync.init($page.url);

  const typeValues = ["all", "table", "view"] as const;
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
    instanceId: "",
    connector: "",
    searchPattern: undefined as string | undefined,
  });
  $: tablesParams.set({
    instanceId,
    connector: connectorName,
    searchPattern,
  });
  const tablesList = useInfiniteTablesList(tablesParams);

  // Filter out temporary tables (e.g., __rill_tmp_ prefixed tables)
  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);

  // TODO: populate from OLAPGetTable responses when per-table metadata is available
  let isViewMap = new Map<string, boolean>();
  // createQuery (unlike createInfiniteQuery) handles re-creation in $: blocks safely
  $: modelResourcesQuery = useModelResources(instanceId);
  $: modelResources = $modelResourcesQuery.data ?? new Map();
  let typeFilter: (typeof typeValues)[number] = parseEnumParam(
    $page.url.searchParams.get("type"),
    typeValues,
    "all",
  );
  let typeDropdownOpen = false;
  let mounted = false;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    searchText = parseStringParam($page.url.searchParams.get("q"));
    typeFilter = parseEnumParam(
      $page.url.searchParams.get("type"),
      typeValues,
      "all",
    );
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({ q: searchText, type: typeFilter });
  }

  onMount(() => {
    mounted = true;
  });

  type TypeOption = { label: string; value: "all" | "table" | "view" };
  const typeOptions: TypeOption[] = [
    { label: "All", value: "all" },
    { label: "Table", value: "table" },
    { label: "View", value: "view" },
  ];

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

  const createTrigger = createRuntimeServiceCreateTrigger();
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
        instanceId,
        data: {
          models: [{ model: selectedModelName, ...opts }],
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
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

  <div class="flex flex-row items-center gap-x-4 min-h-9">
    <div class="flex-1 min-w-0 min-h-9">
      <Search
        bind:value={searchText}
        placeholder="Search"
        large
        autofocus={false}
        showBorderOnFocus={false}
      />
    </div>

    <DropdownMenu.Root bind:open={typeDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {typeDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {typeOptions.find((o) => o.value === typeFilter)?.label ?? "All"}
        </span>
        {#if typeDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-32">
        {#each typeOptions as option}
          <DropdownMenu.Item
            on:click={() => {
              typeFilter = option.value;
            }}
          >
            {option.label}
          </DropdownMenu.Item>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    {#if typeFilter !== "all" || searchText}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
        on:click={() => {
          typeFilter = "all";
          searchText = "";
        }}
      >
        Clear filters
      </button>
    {/if}
  </div>

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
