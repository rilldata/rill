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
    createRuntimeServiceCreateTriggerMutation,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

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

  // Sidebar selection from URL ?resource= param.
  // Only use controlled mode when the URL explicitly names a resource;
  // otherwise let the sidebar auto-select internally without touching the URL.
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
    goto(`/graph?${params.toString()}`, {
      replaceState: true,
      noScroll: true,
    });
  }

  function handleKindChange(kind: string | null) {
    if (kind) {
      goto(`/graph?kind=${kind}`);
    } else {
      goto("/graph");
    }
  }

  // Status filter state — synced to URL ?status= param
  $: selectedStatuses = (
    $page.url.searchParams.get("status")?.split(",") ?? []
  ).filter(Boolean) as ResourceStatusFilterValue[];

  const statusOptions: { label: string; value: ResourceStatusFilterValue }[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Warning", value: "warning" },
    { label: "Errored", value: "errored" },
  ];

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

  // True when the URL has any explicit filter params
  $: hasUrlFilters =
    !!urlParams.kind ||
    urlParams.resources.length > 0 ||
    selectedStatuses.length > 0;

  // Clear all filters
  function handleClearFilters() {
    const url = new URL($page.url);
    url.searchParams.delete("status");
    url.searchParams.delete("kind");
    url.searchParams.delete("resource");
    goto(url.toString(), { replaceState: true, noScroll: true });
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
    compactToolbar
    statusFilterOptions={statusOptions}
    onStatusToggle={toggleStatus}
    onClearFilters={handleClearFilters}
    onSelectAll={() => goto("/graph")}
    {hasUrlFilters}
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
