<script lang="ts">
  import { goto } from "$app/navigation";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createAndPreviewExplore } from "./create-and-preview-explore";
  import { GitBranch } from "lucide-svelte";
  import ResourceGraphOverlay from "@rilldata/web-common/features/resource-graph/ResourceGraphOverlay.svelte";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: ({ instanceId } = $runtime);
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;

  /**
   * Get the name of the dashboard's underlying model (if any).
   * Note that not all dashboards have an underlying model. Some dashboards are
   * underpinned by a source/table.
   */
  $: referenceModelName = $resourceQuery?.data?.meta?.refs?.filter(
    (ref) => ref.kind === ResourceKind.Model,
  )?.[0]?.name;

  const editModel = async () => {
    if (!referenceModelName) return;
    const artifact = fileArtifacts.findFileArtifact(
      ResourceKind.Model,
      referenceModelName,
    );
    if (!artifact) return;
    const previousScreenName = getScreenNameFromPage();
    await goto(`/files${artifact.path}`);
    await behaviourEvent?.fireNavigationEvent(
      referenceModelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousScreenName,
      MetricsEventScreenName.Model,
    );
  };

  let graphOverlayOpen = false;

  $: resourcesQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
    {
      query: {
        retry: 2,
        refetchOnMount: true,
        refetchOnWindowFocus: false,
        enabled: !!instanceId && graphOverlayOpen,
      },
    },
    queryClient,
  );

  $: allResources = $resourcesQuery.data?.resources ?? [];
  $: resourcesLoading = $resourcesQuery.isLoading;
  $: resourcesError = $resourcesQuery.error
    ? "Failed to load project resources."
    : null;

  function viewGraph() {
    graphOverlayOpen = true;
  }
</script>

{#if referenceModelName}
  <NavigationMenuItem on:click={editModel}>
    <Model slot="icon" />
    Edit model
  </NavigationMenuItem>
{/if}
{#if resource}
  <NavigationMenuItem
    on:click={() => createAndPreviewExplore(queryClient, instanceId, resource)}
  >
    <ExploreIcon slot="icon" />
    Generate dashboard
  </NavigationMenuItem>
{/if}

<NavigationMenuItem on:click={viewGraph}>
  <GitBranch slot="icon" size="14px" />
  View dependency graph
</NavigationMenuItem>

<ResourceGraphOverlay
  bind:open={graphOverlayOpen}
  anchorResource={resource}
  resources={allResources}
  isLoading={resourcesLoading}
  error={resourcesError}
/>
