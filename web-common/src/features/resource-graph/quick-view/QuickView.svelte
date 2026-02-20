<script lang="ts">
  import ResourceGraphOverlay from "../embedding/ResourceGraphOverlay.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    closeResourceGraphQuickView,
    resourceGraphQuickViewState,
  } from "./quick-view-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServiceListResources,
  } from "@rilldata/web-common/runtime-client";

  $: currentState = $resourceGraphQuickViewState;
  $: anchorResource = currentState.anchorResource ?? undefined;

  $: ({ instanceId } = $runtime);

  $: shouldFetchResources = currentState.open && !!instanceId;

  $: instanceQuery = createRuntimeServiceGetInstance(
    instanceId,
    { sensitive: true },
    { query: { enabled: !!instanceId } },
  );
  $: olapConnectorName = $instanceQuery.data?.instance?.olapConnector;

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

  // Filter out non-OLAP connectors
  $: allResources = ($resourcesQuery.data?.resources ?? []).filter((r) => {
    if (r.meta?.name?.kind !== ResourceKind.Connector) return true;
    return r.meta?.name?.name === olapConnectorName;
  });
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
