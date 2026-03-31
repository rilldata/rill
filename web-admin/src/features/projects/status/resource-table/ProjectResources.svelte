<script lang="ts">
  import { page } from "$app/stores";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { SingletonProjectParserName } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import { useResources } from "../selectors";
  import { filterResources } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { onMount } from "svelte";
  import {
    FilterIcon,
    LayoutGridIcon,
    ListIcon,
    SearchIcon,
    XIcon,
  } from "lucide-svelte";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  $: basePath = `/${$page.params.organization}/${$page.params.project}/-/status/resources`;
  $: isGraphView = $page.route.id?.endsWith("/graph") ?? false;

  const filterSync = createUrlFilterSync([
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let searchExpanded = false;
  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
  let selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
  let mounted = false;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
    selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
    searchText = parseStringParam($page.url.searchParams.get("q"));
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({
      kind: selectedTypes,
      status: selectedStatuses,
      q: searchText,
    });
  }

  onMount(() => {
    mounted = true;
  });

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Warning", value: "warning" },
    { label: "Errored", value: "errored" },
  ];

  // Resource types grouped by category
  const filterSections: { label: string; types: string[] }[] = [
    {
      label: "Data",
      types: [ResourceKind.Source, ResourceKind.Model, ResourceKind.Connector],
    },
    {
      label: "Dashboards",
      types: [
        ResourceKind.MetricsView,
        ResourceKind.Explore,
        ResourceKind.Canvas,
        ResourceKind.Theme,
      ],
    },
    {
      label: "Automation",
      types: [ResourceKind.Report, ResourceKind.Alert, ResourceKind.API],
    },
  ];

  $: resources = useResources(runtimeClient);

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  // Filter resources by type, search text, and status
  $: filteredResources = filterResources(
    $resources.data?.resources,
    selectedTypes,
    searchText,
    selectedStatuses,
  );

  $: activeFilterCount = selectedTypes.length + selectedStatuses.length;
  $: hasActiveFilters =
    selectedTypes.length > 0 ||
    selectedStatuses.length > 0 ||
    searchText.length > 0;

  function toggleType(type: string) {
    if (selectedTypes.includes(type)) {
      selectedTypes = selectedTypes.filter((t) => t !== type);
    } else {
      selectedTypes = [...selectedTypes, type];
    }
  }

  function toggleStatus(status: string) {
    if (selectedStatuses.includes(status)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== status);
    } else {
      selectedStatuses = [...selectedStatuses, status];
    }
  }

  function clearFilters() {
    selectedTypes = [];
    selectedStatuses = [];
    searchText = "";
    searchExpanded = false;
  }

  function toggleSearchExpanded() {
    searchExpanded = !searchExpanded;
    if (!searchExpanded) {
      searchText = "";
    }
  }

  function refreshAllSourcesAndModels() {
    void $createTrigger.mutateAsync({ all: true }).then(() => {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    });
  }
</script>

<section class="flex flex-col gap-y-3">
  <!-- Row 1: Resources + Refresh -->
  <div class="flex items-center justify-between h-9">
    <h2 class="text-lg font-medium">Resources</h2>
    <Button
      type="secondary"
      large
      class="shrink-0 whitespace-nowrap"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={hasReconcilingResources}
    >
      <span class="hidden lg:inline">Refresh all sources and models</span>
      <span class="lg:hidden">Refresh all</span>
    </Button>
  </div>

  <!-- Row 2: [Filter button] ...spacer... [search] [Grid/List] -->
  <div class="flex items-center gap-x-2 min-h-8">
    <!-- Filter dropdown -->
    <DropdownMenu.Root bind:open={filterDropdownOpen}>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <button
            {...props}
            class="filter-trigger"
          >
            <FilterIcon size="14px" />
            <span>Filter</span>
            {#if activeFilterCount > 0}
              <span class="filter-badge">{activeFilterCount}</span>
            {/if}
          </button>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-52">
        <DropdownMenu.Group>
          <DropdownMenu.Label class="uppercase text-[10px] tracking-wide"
            >Status</DropdownMenu.Label
          >
          {#each statusFilters as status}
            <DropdownMenu.CheckboxItem
              closeOnSelect={false}
              checked={selectedStatuses.includes(status.value)}
              onCheckedChange={() => toggleStatus(status.value)}
            >
              {status.label}
            </DropdownMenu.CheckboxItem>
          {/each}
        </DropdownMenu.Group>
        <DropdownMenu.Separator />
        {#each filterSections as section, i}
          <DropdownMenu.Group>
            <DropdownMenu.Label class="uppercase text-[10px] tracking-wide"
              >{section.label}</DropdownMenu.Label
            >
            {#each section.types as type}
              <DropdownMenu.CheckboxItem
                closeOnSelect={false}
                checked={selectedTypes.includes(type)}
                onCheckedChange={() => toggleType(type)}
              >
                {prettyResourceKind(type)}
              </DropdownMenu.CheckboxItem>
            {/each}
          </DropdownMenu.Group>
          {#if i < filterSections.length - 1}
            <DropdownMenu.Separator />
          {/if}
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <div class="flex-1"></div>

    <!-- Search icon / expandable search -->
    {#if searchExpanded}
      <div class="flex items-center w-56 shrink-0">
        <Search
          bind:value={searchText}
          placeholder="Search resources..."
          autofocus={true}
          showBorderOnFocus={false}
          retainValueOnMount
        />
        <button
          class="ml-1 p-1 text-fg-primary hover:bg-surface-hover rounded-sm"
          onclick={toggleSearchExpanded}
        >
          <XIcon size="14px" />
        </button>
      </div>
    {:else}
      <button
        class="toolbar-icon-btn"
        onclick={toggleSearchExpanded}
      >
        <SearchIcon size="14px" />
      </button>
    {/if}

    <!-- Grid / List toggle -->
    <div class="view-toggle">
      <a href="{basePath}/graph" class="toggle-btn" class:active={isGraphView}>
        <LayoutGridIcon size="14px" />
      </a>
      <a href={basePath} class="toggle-btn" class:active={!isGraphView}>
        <ListIcon size="14px" />
      </a>
    </div>
  </div>

  <hr class="border-t border-gray-200 -mt-1 mb-1" />

  <!-- Row 3: Filter pills + Clear all (when any filter or search is active) -->
  {#if hasActiveFilters}
    <div class="filter-pills-row">
      <div class="filter-pills-scroll">
        {#if selectedStatuses.length > 0}
          <button
            class="filter-pill"
            onclick={() => (selectedStatuses = [])}
          >
            <span>Status = {selectedStatuses
              .map((s) => statusFilters.find((f) => f.value === s)?.label ?? s)
              .join(", ")}</span>
            <XIcon size="10px" />
          </button>
        {/if}
        {#if selectedTypes.length > 0}
          <button
            class="filter-pill"
            onclick={() => (selectedTypes = [])}
          >
            <span>Type = {selectedTypes.map(prettyResourceKind).join(", ")}</span>
            <XIcon size="10px" />
          </button>
        {/if}
      </div>
      <button
        class="filter-pills-clear"
        onclick={clearFilters}
      >
        Clear all
      </button>
    </div>
  {/if}

  <!-- Content -->
  {#if $resources.isLoading}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.data}
    <ProjectResourcesTable data={filteredResources} />
  {/if}

  <div class="parse-errors">
    <h3 class="parse-errors-header">
      Parse Errors
      {#if parseErrors.length > 0}
        <span class="parse-errors-badge">{parseErrors.length}</span>
      {/if}
    </h3>
    {#if parseErrors.length === 0}
      <p class="text-sm text-fg-secondary">No parse errors</p>
    {:else}
      <div class="parse-errors-list">
        {#each parseErrors as error ((error.filePath ?? "") + ":" + error.message)}
          <div class="parse-error-item">
            {#if error.filePath}
              <span class="parse-error-file">{error.filePath}</span>
            {/if}
            <span class="parse-error-message">{error.message}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .filter-trigger {
    @apply flex items-center gap-1.5 px-3 py-1.5 rounded-sm bg-primary-50 text-sm text-primary-600;
  }
  .filter-trigger:hover {
    @apply bg-primary-100;
  }

  .filter-badge {
    @apply text-[10px] font-semibold bg-primary-500 text-white rounded-full w-4 h-4 flex items-center justify-center;
  }

  .filter-pills-row {
    @apply flex items-center min-h-7 relative;
  }

  .filter-pills-scroll {
    @apply flex items-center gap-1.5 flex-1 min-w-0 overflow-hidden;
  }

  .filter-pills-clear {
    @apply shrink-0 text-xs text-fg-primary hover:underline whitespace-nowrap pl-2 pr-1;
  }

  .filter-pill {
    @apply flex items-center gap-1.5 text-xs font-medium text-fg-primary border border-gray-300 rounded-sm px-2 py-1 whitespace-nowrap shrink-0;
  }
  .filter-pill:hover {
    @apply bg-surface-hover;
  }

  .toolbar-icon-btn {
    @apply p-1.5 rounded-sm text-fg-primary;
  }
  .toolbar-icon-btn:hover {
    @apply bg-surface-hover;
  }

  .view-toggle {
    @apply flex rounded-sm border border-gray-200 overflow-hidden shrink-0;
  }
  .toggle-btn {
    @apply flex items-center p-1.5 text-fg-primary no-underline;
  }
  .toggle-btn:hover {
    @apply bg-surface-hover;
  }
  .toggle-btn.active {
    @apply bg-primary-100 text-primary-600;
  }

  .parse-errors {
    @apply pt-4 mt-2;
  }
  .parse-errors-header {
    @apply text-sm font-semibold text-fg-primary flex items-center gap-2 mb-3;
  }
  .parse-errors-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .parse-errors-list {
    @apply flex flex-col gap-2;
  }
  .parse-error-item {
    @apply flex flex-col gap-0.5 px-3 py-2 rounded-md bg-red-50 text-sm;
  }
  .parse-error-file {
    @apply font-mono text-xs text-fg-secondary;
  }
  .parse-error-message {
    @apply text-red-700;
  }
</style>
