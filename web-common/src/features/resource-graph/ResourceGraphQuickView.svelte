<script lang="ts">
  import ResourceGraphOverlay from "./ResourceGraphOverlay.svelte";
  import {
    closeResourceGraphOverlay,
    resourceGraphOverlayAnchor,
  } from "./resource-graph-overlay-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  let currentAnchor: V1Resource | null = null;
  let anchor: V1Resource | null = null;
  let overlayOpen = false;

  $: anchor = $resourceGraphOverlayAnchor;

  $: if (anchor !== currentAnchor) {
    currentAnchor = anchor;
    overlayOpen = !!currentAnchor;
  }

  $: ({ instanceId } = $runtime);

  $: resourcesQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
    {
      query: {
        retry: 2,
        refetchOnMount: true,
        refetchOnWindowFocus: false,
        enabled: !!instanceId,
      },
    },
    queryClient,
  );

  $: resources = $resourcesQuery.data?.resources ?? [];
  $: resourcesError = $resourcesQuery.error
    ? "Failed to load project resources."
    : null;

  $: if (!overlayOpen && currentAnchor) {
    closeResourceGraphOverlay();
    currentAnchor = null;
  }
</script>

{#if currentAnchor}
  <ResourceGraphOverlay
    bind:open={overlayOpen}
    anchorResource={currentAnchor}
    {resources}
    isLoading={$resourcesQuery.isLoading}
    error={resourcesError}
  />
{/if}
