<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import {
    parseGraphUrlParams,
    tokenForKind,
    tokenForSeedString,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";
  import type { ResourceStatusFilterValue } from "@rilldata/web-common/features/resource-graph/shared/types";
  import { setGraphNavigation } from "@rilldata/web-common/features/resource-graph/shared/graph-navigation-context";
  import RefreshConfirmDialog from "@rilldata/web-common/features/resource-graph/shared/RefreshConfirmDialog.svelte";
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

  // Parse URL parameters
  $: urlParams = parseGraphUrlParams($page.url);
  $: derivedKindFromResource =
    urlParams.resources.length > 0
      ? tokenForSeedString(urlParams.resources[0])
      : null;
  $: activeKind = urlParams.kind ?? derivedKindFromResource ?? "dashboards";
  $: seeds = urlParams.kind
    ? [urlParams.kind]
    : urlParams.resources.length > 0
      ? urlParams.resources
      : [activeKind];

  // Sidebar selection from URL ?resource= param
  $: hasResourceParam = urlParams.resources.length > 0;
  $: selectedGroupId = hasResourceParam ? urlParams.resources[0] : null;

  function handleSelectedGroupChange(groupId: string | null) {
    if (!groupId) return;
    const name = groupId.includes(":") ? groupId.split(":").pop() : groupId;
    const kindPart = groupId.includes(":")
      ? groupId.split(":").slice(0, -1).join(":")
      : null;
    const derivedKind = kindPart ? tokenForKind(kindPart) : null;
    const params = new URLSearchParams();
    params.set("kind", derivedKind ?? activeKind);
    if (name) params.set("resource", name);
    goto(`${graphBasePath}?${params.toString()}`, {
      replaceState: true,
      noScroll: true,
    });
  }

  function handleKindChange(kind: string | null) {
    if (kind) {
      goto(`${graphBasePath}?kind=${kind}`);
    } else {
      goto(graphBasePath);
    }
  }

  // Status filter state
  let selectedStatuses: ResourceStatusFilterValue[] = [];

  const statusOptions: { label: string; value: ResourceStatusFilterValue }[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Warning", value: "warning" },
    { label: "Errored", value: "errored" },
  ];

  function toggleStatus(value: ResourceStatusFilterValue) {
    if (selectedStatuses.includes(value)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== value);
    } else {
      selectedStatuses = [...selectedStatuses, value];
    }
  }

  $: hasUrlFilters = !!urlParams.kind || urlParams.resources.length > 0;

  function handleClearFilters() {
    selectedStatuses = [];
    handleKindChange(null);
  }

  // Refresh all
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
    statusFilterOptions={statusOptions}
    onStatusToggle={toggleStatus}
    onClearFilters={handleClearFilters}
    onSelectAll={() => goto(graphBasePath)}
    {hasUrlFilters}
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
    @apply flex flex-col w-full overflow-hidden;
    min-height: 600px;
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
