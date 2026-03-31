<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import { setGraphNavigation } from "@rilldata/web-common/features/resource-graph/shared/graph-navigation-context";
  import {
    deriveGraphState,
    buildGroupChangeParams,
  } from "@rilldata/web-common/features/resource-graph/shared/graph-page-utils";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { onMount } from "svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { useResources } from "@rilldata/web-admin/features/projects/status/selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
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

  let isConfirmDialogOpen = false;

  $: allResources = useResources(runtimeClient);
  $: hasReconcilingResources = $allResources.data?.resources?.some(
    isResourceReconciling,
  );

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

  $: basePath = `/${$page.params.organization}/${$page.params.project}/-/status/resources`;
  $: graphBasePath = `${basePath}/graph`;

  const filterSync = createUrlFilterSync([
    { key: "status", type: "array" },
    { key: "q", type: "string" },
    { key: "isolated", type: "enum", defaultValue: "hidden" },
  ]);
  filterSync.init($page.url);

  let mounted = false;

  onMount(() => {
    mounted = true;
  });

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
    searchText = parseStringParam($page.url.searchParams.get("q"));
    hideIsolated =
      ($page.url.searchParams.get("isolated") ?? "hidden") === "hidden";
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({
      status: selectedStatuses,
      q: searchText,
      isolated: hideIsolated ? "hidden" : "shown",
    });
  }

  setGraphNavigation({
    viewLineage(kindToken, resourceName) {
      const params = new URLSearchParams();
      if (kindToken) params.set("kind", kindToken);
      if (resourceName) {
        params.set("resource", resourceName);
        params.set("q", resourceName);
        searchText = resourceName;
        searchExpanded = true;
      }
      goto(`${graphBasePath}?${params.toString()}`);
    },
  });

  $: graphState = deriveGraphState($page.url);
  $: ({ activeKind, seeds, selectedGroupId } = graphState);

  function handleSelectedGroupChange(groupId: string | null) {
    if (!groupId) return;
    const params = buildGroupChangeParams(groupId, activeKind);
    goto(`${graphBasePath}?${params.toString()}`, {
      replaceState: true,
      noScroll: true,
    });
  }

  $: hasUrlFilters =
    !!graphState.urlParams.kind || graphState.urlParams.resources.length > 0;

  // Filter state
  let filterDropdownOpen = false;
  let searchExpanded = false;
  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedStatuses: string[] = parseArrayParam($page.url.searchParams.get("status"));
  let hideIsolated =
    ($page.url.searchParams.get("isolated") ?? "hidden") === "hidden";

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Warning", value: "warning" },
    { label: "Errored", value: "errored" },
  ];

  function toggleStatus(status: string) {
    if (selectedStatuses.includes(status)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== status);
    } else {
      selectedStatuses = [...selectedStatuses, status];
    }
  }

  function clearFilters() {
    selectedStatuses = [];
    searchText = "";
    searchExpanded = false;
    hideIsolated = true;
    goto(graphBasePath);
  }

  function toggleSearchExpanded() {
    searchExpanded = !searchExpanded;
    if (!searchExpanded) {
      searchText = "";
    }
  }

  $: activeFilterCount =
    selectedStatuses.length +
    (!hideIsolated ? 1 : 0);
  $: hasActiveFilters =
    selectedStatuses.length > 0 ||
    !hideIsolated ||
    searchText.length > 0;

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
</script>

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
<div class="flex items-center gap-x-2 min-h-8 mt-3">
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
      <DropdownMenu.Group>
        <DropdownMenu.Label class="uppercase text-[10px] tracking-wide"
          >Visibility</DropdownMenu.Label
        >
        <DropdownMenu.CheckboxItem
          closeOnSelect={false}
          checked={hideIsolated}
          onCheckedChange={() => (hideIsolated = !hideIsolated)}
        >
          Hide isolated
        </DropdownMenu.CheckboxItem>
      </DropdownMenu.Group>
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
    <a href={graphBasePath} class="toggle-btn active">
      <LayoutGridIcon size="14px" />
    </a>
    <a href={basePath} class="toggle-btn">
      <ListIcon size="14px" />
    </a>
  </div>
</div>

<hr class="border-t border-gray-200 mt-1.5" />

<!-- Row 3: Filter pills + Clear all (when any filter or search is active) -->
{#if hasActiveFilters}
  <div class="filter-pills-row mt-1.5">
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
      {#if !hideIsolated}
        <button
          class="filter-pill"
          onclick={() => (hideIsolated = true)}
        >
          <span>Show isolated</span>
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

<div class="graph-wrapper">
  <GraphContainer
    {seeds}
    statusFilter={selectedStatuses}
    searchQuery={searchText}
    showSummary={false}
    layout="sidebar"
    {selectedGroupId}
    onSelectedGroupChange={handleSelectedGroupChange}
    onSelectAll={() => goto(graphBasePath)}
    {hasUrlFilters}
    flushToolbar
    showTitle={false}
    showToolbar={false}
    showIsolatedResources={!hideIsolated}
  />
</div>

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

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .graph-wrapper {
    @apply flex flex-col w-full min-w-0 overflow-hidden mt-3;
    height: 600px;
  }

  /* Prevent sidebar-main from overflowing past the toolbar */
  .graph-wrapper :global(.sidebar-main) {
    height: 0;
    min-height: 0;
    flex: 1 1 0%;
  }

  .filter-trigger {
    @apply flex items-center gap-1.5 px-4 py-1.5 rounded-sm bg-primary-50 text-sm text-primary-600;
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
