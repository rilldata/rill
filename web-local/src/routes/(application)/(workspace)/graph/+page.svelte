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
    createRuntimeServiceCreateTriggerMutation,
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

  setGraphNavigation({
    viewLineage(kindToken, resourceName) {
      const params = new URLSearchParams();
      if (kindToken) params.set("kind", kindToken);
      if (resourceName) params.set("resource", resourceName);
      goto(`/graph?${params.toString()}`);
    },
    openFile(filePath) {
      try {
        const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
        localStorage.setItem(
          filePath,
          JSON.stringify({ ...prefs, view: "code" }),
        );
      } catch {
        // ignore
      }
      goto(`/files${filePath}`);
    },
  });

  $: ({ instanceId } = runtimeClient);

  $: graphState = deriveGraphState($page.url);
  $: ({ activeKind, seeds, selectedGroupId } = graphState);

  function handleSelectedGroupChange(groupId: string | null) {
    if (!groupId) return;
    const params = buildGroupChangeParams(groupId, activeKind);
    goto(`/graph?${params.toString()}`, {
      replaceState: true,
      noScroll: true,
    });
  }

  // Status filter state — synced to URL ?status= param
  $: selectedStatuses = (
    $page.url.searchParams.get("status")?.split(",") ?? []
  ).filter(Boolean) as ResourceStatusFilterValue[];

  function toggleStatus(value: ResourceStatusFilterValue) {
    const next = selectedStatuses.includes(value)
      ? selectedStatuses.filter((s) => s !== value)
      : [...selectedStatuses, value];
    const url = new URL($page.url);
    if (next.length > 0) {
      url.searchParams.set("status", next.join(","));
    } else {
      url.searchParams.delete("status");
    }
    goto(url.toString(), { replaceState: true, noScroll: true });
  }

  $: hasUrlFilters =
    !!graphState.urlParams.kind ||
    graphState.urlParams.resources.length > 0 ||
    selectedStatuses.length > 0;

  function handleClearFilters() {
    const url = new URL($page.url);
    url.searchParams.delete("status");
    url.searchParams.delete("kind");
    url.searchParams.delete("resource");
    goto(url.toString(), { replaceState: true, noScroll: true });
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
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

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
    onSelectAll={() => goto("/graph")}
    {hasUrlFilters}
    {showIsolatedResources}
    onShowIsolatedChange={handleIsolatedChange}
  />
</div>

<RefreshConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .graph-wrapper {
    @apply flex flex-col size-full overflow-hidden;
  }
</style>
