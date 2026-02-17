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
    type V1OlapTableInfo,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ProjectTablesTable from "./ProjectTablesTable.svelte";
  import {
    useTablesList,
    useTableMetadata,
    useModelResources,
  } from "../selectors";
  import { filterTemporaryTables } from "./utils";
  import ResourceSpecDialog from "../resource-table/ResourceSpecDialog.svelte";
  import ModelPartitionsDialog from "./ModelPartitionsDialog.svelte";
  import RefreshErroredPartitionsDialog from "./RefreshErroredPartitionsDialog.svelte";
  import RefreshResourceConfirmDialog from "../resource-table/RefreshResourceConfirmDialog.svelte";
  import { onMount } from "svelte";

  $: ({ instanceId } = $runtime);

  // OLAP connector info
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: isDuckDB = !olapConnector || olapConnector.type === "duckdb";

  $: tablesList = useTablesList(instanceId, "");

  // Filter out temporary tables (e.g., __rill_tmp_ prefixed tables)
  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);

  $: tableMetadata = useTableMetadata(instanceId, "", filteredTables);
  $: isViewMap = new Map($tableMetadata?.data?.isView ?? []);
  $: modelResourcesQuery = useModelResources(instanceId);
  $: modelResources = $modelResourcesQuery.data ?? new Map();

  // Filters — initialized from URL params
  $: qParam = $page.url.searchParams.get("q");
  $: typeParam = $page.url.searchParams.get("type");

  let searchText = $page.url.searchParams.get("q") ?? "";
  let typeFilter: "all" | "table" | "view" = (() => {
    const t = $page.url.searchParams.get("type");
    return t === "table" || t === "view" ? t : "all";
  })();
  let typeDropdownOpen = false;
  let mounted = false;
  let lastSyncedSearch = $page.url.search;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && $page.url.search !== lastSyncedSearch) {
    lastSyncedSearch = $page.url.search;
    searchText = qParam ?? "";
    const t = typeParam;
    typeFilter = t === "table" || t === "view" ? t : "all";
  }

  // Sync filter state → URL
  $: if (mounted) {
    syncFiltersToUrl(searchText, typeFilter);
  }

  onMount(() => {
    mounted = true;
  });

  function syncFiltersToUrl(search: string, type: string) {
    const url = new URL($page.url);
    if (search) {
      url.searchParams.set("q", search);
    } else {
      url.searchParams.delete("q");
    }
    if (type !== "all") {
      url.searchParams.set("type", type);
    } else {
      url.searchParams.delete("type");
    }
    lastSyncedSearch = url.search;
    void goto(url.pathname + url.search, {
      replaceState: true,
      noScroll: true,
      keepFocus: true,
    });
  }

  type TypeOption = { label: string; value: "all" | "table" | "view" };
  const typeOptions: TypeOption[] = [
    { label: "All", value: "all" },
    { label: "Table", value: "table" },
    { label: "View", value: "view" },
  ];

  $: metadataLoaded = !$tableMetadata?.isLoading;
  $: displayedTables = applyFilters(
    filteredTables,
    searchText,
    typeFilter,
    isViewMap,
    metadataLoaded,
  );

  function applyFilters(
    tables: V1OlapTableInfo[],
    search: string,
    type: "all" | "table" | "view",
    viewMap: Map<string, boolean>,
    isMetadataLoaded: boolean,
  ): V1OlapTableInfo[] {
    return tables.filter((t) => {
      const name = t.name ?? "";
      const matchesSearch =
        !search || name.toLowerCase().includes(search.toLowerCase());
      if (type === "all" || !isMetadataLoaded) return matchesSearch;
      const isView = viewMap.get(name);
      // If we don't have metadata for this item yet, include it
      if (isView === undefined) return matchesSearch;
      const matchesType =
        (type === "view" && isView) || (type === "table" && !isView);
      return matchesSearch && matchesType;
    });
  }

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
    <h2 class="text-lg font-medium">Tables</h2>
  </div>
  <div class="flex items-center gap-x-3">
    <DropdownMenu.Root bind:open={typeDropdownOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <Button builders={[builder]} type="tertiary">
          <span class="flex items-center gap-x-1.5">
            {typeOptions.find((o) => o.value === typeFilter)?.label ?? "All"}
            {#if typeDropdownOpen}
              <CaretUpIcon size="12px" />
            {:else}
              <CaretDownIcon size="12px" />
            {/if}
          </span>
        </Button>
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
        class="text-sm text-primary-500 hover:text-primary-600"
        on:click={() => {
          typeFilter = "all";
          searchText = "";
        }}
      >
        Clear filters
      </button>
    {/if}

    <div class="flex-1" />

    <div class="w-64">
      <Search
        bind:value={searchText}
        placeholder="Search by name..."
        autofocus={false}
        retainValueOnMount={true}
      />
    </div>
  </div>

  {#if $tablesList.isLoading}
    <div
      class="flex-1 flex flex-col items-center justify-center gap-y-2 text-fg-secondary"
    >
      <DelayedSpinner isLoading={true} size="20px" />
      <span class="text-sm">Loading tables</span>
    </div>
  {:else if $tablesList.isError}
    <div class="text-red-500">
      Error loading tables: {$tablesList.error?.message}
    </div>
  {:else if displayedTables.length > 0}
    <ProjectTablesTable
      tables={displayedTables}
      isView={isViewMap}
      {modelResources}
      onModelInfoClick={handleModelInfoClick}
      onViewPartitionsClick={handleViewPartitionsClick}
      onRefreshErroredClick={handleRefreshErroredClick}
      onIncrementalRefreshClick={handleIncrementalRefreshClick}
      onFullRefreshClick={handleFullRefreshClick}
      onViewLogsClick={handleViewLogsClick}
    />
    {#if $tableMetadata?.isLoading}
      <div class="mt-2 text-xs text-fg-secondary">
        Loading table metadata...
      </div>
    {/if}
  {:else}
    <div class="text-fg-secondary text-sm">
      {#if searchText || typeFilter !== "all"}
        No tables match the current filters
      {:else}
        No tables found
      {/if}
    </div>
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
