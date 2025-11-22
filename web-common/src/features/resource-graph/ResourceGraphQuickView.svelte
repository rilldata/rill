<script lang="ts">
  import ResourceGraphOverlay from "@rilldata/web-common/features/resource-graph/ResourceGraphOverlay.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    closeResourceGraphQuickView,
    resourceGraphQuickViewState,
  } from "@rilldata/web-common/features/resource-graph/resource-graph-quick-view-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";

  const quickViewState = resourceGraphQuickViewState;

  let overlayOpen = false;

  $: currentState = $quickViewState;
  $: anchorResource = currentState.anchorResource ?? undefined;

  $: ({ instanceId } = $runtime);

  $: shouldFetchResources = overlayOpen && !!instanceId;

  $: resourcesQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
    {
      query: {
        retry: 2,
        refetchOnMount: true,
        refetchOnWindowFocus: false,
        enabled: shouldFetchResources,
      },
    },
    queryClient,
  );

  $: allResources = $resourcesQuery.data?.resources ?? [];
  $: resourcesLoading = $resourcesQuery.isLoading;
  $: resourcesError = $resourcesQuery.error
    ? "Failed to load project resources."
    : null;

  $: if (currentState.open && anchorResource && !overlayOpen) {
    overlayOpen = true;
  } else if (!currentState.open && overlayOpen) {
    overlayOpen = false;
  }

  $: if (!overlayOpen && currentState.open) {
    closeResourceGraphQuickView();
  }
</script>

{#if anchorResource}
  <ResourceGraphOverlay
    bind:open={overlayOpen}
    {anchorResource}
    resources={allResources}
    isLoading={resourcesLoading}
    error={resourcesError}
  />
{/if}
