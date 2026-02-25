<script lang="ts">
  import ResourceGraphOverlay from "../embedding/ResourceGraphOverlay.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    closeResourceGraphQuickView,
    resourceGraphQuickViewState,
  } from "./quick-view-store";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";

  const runtimeClient = useRuntimeClient();

  $: currentState = $resourceGraphQuickViewState;
  $: anchorResource = currentState.anchorResource ?? undefined;

  $: ({ instanceId } = runtimeClient);

  $: shouldFetchResources = currentState.open && !!instanceId;

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

  function handleClose() {
    closeResourceGraphQuickView();
  }
</script>

{#if anchorResource}
  <ResourceGraphOverlay
    open={currentState.open}
    onClose={handleClose}
    {anchorResource}
    resources={allResources}
    isLoading={resourcesLoading}
    error={resourcesError}
  />
{/if}
