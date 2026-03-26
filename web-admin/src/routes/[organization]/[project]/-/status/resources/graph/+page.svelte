<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import type { ResourceStatusFilterValue } from "@rilldata/web-common/features/resource-graph/shared/types";
  import { setGraphNavigation } from "@rilldata/web-common/features/resource-graph/shared/graph-navigation-context";
  import RefreshConfirmDialog from "@rilldata/web-common/features/resource-graph/shared/RefreshConfirmDialog.svelte";
  import {
    deriveGraphState,
    readIsolatedPreference,
    writeIsolatedPreference,
    buildGroupChangeParams,
    STATUS_FILTER_OPTIONS,
  } from "@rilldata/web-common/features/resource-graph/shared/graph-page-utils";
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

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  let showIsolatedResources = readIsolatedPreference();

  function handleIsolatedChange(value: boolean) {
    showIsolatedResources = value;
    writeIsolatedPreference(value);
  }

  $: graphBasePath = `/${$page.params.organization}/${$page.params.project}/-/status/resources/graph`;

  setGraphNavigation({
    viewLineage(kindToken, resourceName) {
      const params = new URLSearchParams();
      if (kindToken) params.set("kind", kindToken);
      if (resourceName) params.set("resource", resourceName);
      goto(`${graphBasePath}?${params.toString()}`);
    },
  });

  $: ({ instanceId } = runtimeClient);

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

  // Status filter state
  let selectedStatuses: ResourceStatusFilterValue[] = [];

  function toggleStatus(value: ResourceStatusFilterValue) {
    if (selectedStatuses.includes(value)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== value);
    } else {
      selectedStatuses = [...selectedStatuses, value];
    }
  }

  $: hasUrlFilters =
    !!graphState.urlParams.kind || graphState.urlParams.resources.length > 0;

  function handleClearFilters() {
    selectedStatuses = [];
    goto(graphBasePath);
  }

  let isConfirmDialogOpen = false;

  function handleRefreshAll() {
    isConfirmDialogOpen = true;
  }

  function refreshAllSourcesAndModels() {
    isConfirmDialogOpen = false;
    void $triggerMutation
      .mutateAsync({
        all: true,
      })
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            instanceId,
            undefined,
          ),
        });
      })
      .catch((err) => {
        console.error("Failed to refresh all sources and models:", err);
      });
  }

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

<div class="graph-wrapper">
  <GraphContainer
    {seeds}
    statusFilter={selectedStatuses}
    showSummary={false}
    layout="sidebar"
    {selectedGroupId}
    onSelectedGroupChange={handleSelectedGroupChange}
    onRefreshAll={handleRefreshAll}
    statusFilterOptions={STATUS_FILTER_OPTIONS}
    onStatusToggle={toggleStatus}
    onClearFilters={handleClearFilters}
    onSelectAll={() => goto(graphBasePath)}
    {hasUrlFilters}
    flushToolbar
    showTitle={false}
    {showIsolatedResources}
    onShowIsolatedChange={handleIsolatedChange}
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

<RefreshConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .graph-wrapper {
    @apply flex flex-col w-full min-w-0 overflow-hidden;
    height: 600px;
  }

  /* Prevent sidebar-main from overflowing past the toolbar */
  .graph-wrapper :global(.sidebar-main) {
    height: 0;
    min-height: 0;
    flex: 1 1 0%;
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
