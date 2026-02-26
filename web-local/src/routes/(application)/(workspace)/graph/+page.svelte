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
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  $: ({ instanceId } = $runtime);

  // Parse URL parameters
  $: urlParams = parseGraphUrlParams($page.url);
  $: derivedKindFromResource =
    urlParams.resources.length > 0
      ? tokenForSeedString(urlParams.resources[0])
      : null;
  $: activeKind = urlParams.kind ?? derivedKindFromResource ?? "connector";
  $: seeds = urlParams.kind
    ? [urlParams.kind]
    : urlParams.resources.length > 0
      ? urlParams.resources
      : [activeKind];

  // Sidebar selection from URL ?resource= param
  $: selectedResource =
    urlParams.resources.length > 0 ? urlParams.resources[0] : null;
  $: selectedGroupId = selectedResource;

  function handleSelectedGroupChange(groupId: string | null) {
    if (!groupId) return;
    const name = groupId.includes(":") ? groupId.split(":").pop() : groupId;
    // Derive kind from the fully qualified group ID (e.g. "rill.runtime.v1.Model:orders")
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

  // Status filter state
  let selectedStatuses: ResourceStatusFilterValue[] = [];

  const statusOptions: { label: string; value: ResourceStatusFilterValue }[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Errored", value: "errored" },
  ];

  function toggleStatus(value: ResourceStatusFilterValue) {
    if (selectedStatuses.includes(value)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== value);
    } else {
      selectedStatuses = [...selectedStatuses, value];
    }
  }

  // Clear all filters (reset to OLAP connector default)
  function handleClearFilters() {
    selectedStatuses = [];
    handleKindChange("connector");
  }

  // Refresh all
  let isConfirmDialogOpen = false;

  function handleRefreshAll() {
    isConfirmDialogOpen = true;
  }

  function refreshAllSourcesAndModels() {
    isConfirmDialogOpen = false;
    void $createTrigger
      .mutateAsync({
        instanceId,
        data: { all: true },
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
    statusFilterOptions={statusOptions}
    onStatusToggle={toggleStatus}
    onClearFilters={handleClearFilters}
  />
</div>

<AlertDialog.Root bind:open={isConfirmDialogOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Refresh all sources and models?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="flex flex-col gap-y-2 mt-1">
          <p>This will refresh all project sources and models.</p>
          <p>
            <span class="font-medium">Note:</span> To refresh a single resource,
            click the '...' button on a node and select the refresh option.
          </p>
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="tertiary"
        onClick={() => {
          isConfirmDialogOpen = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={refreshAllSourcesAndModels}
        >Yes, refresh</Button
      >
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  .graph-wrapper {
    @apply flex flex-col size-full overflow-hidden;
  }
</style>
