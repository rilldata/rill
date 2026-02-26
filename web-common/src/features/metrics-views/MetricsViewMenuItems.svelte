<script lang="ts">
  import { goto } from "$app/navigation";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { openResourceGraphQuickView } from "@rilldata/web-common/features/resource-graph/quick-view/quick-view-store";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { GitBranch, WandIcon } from "lucide-svelte";
  import { createCanvasDashboardFromMetricsView } from "./ai-generation/generateMetricsView";
  import { createAndPreviewExplore } from "./create-and-preview-explore";

  const runtimeClient = useRuntimeClient();
  const { ai, generateCanvas } = featureFlags;

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: ({ instanceId } = runtimeClient);
  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: resource = $resourceQuery.data;

  /**
   * Get the name of the dashboard's underlying model (if any).
   * Note that not all dashboards have an underlying model. Some dashboards are
   * underpinned by a source/table.
   */
  $: referenceModelName = $resourceQuery?.data?.meta?.refs?.filter(
    (ref) => ref.kind === ResourceKind.Model,
  )?.[0]?.name;

  $: hasMenuItems = Boolean(referenceModelName || resource);

  $: metricsViewName = resource?.meta?.name?.name;

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

  function viewGraph() {
    if (!resource) {
      console.warn(
        "[MetricsViewMenuItems] Cannot open resource graph: resource unavailable.",
      );
      return;
    }
    openResourceGraphQuickView(resource);
  }

  async function handleCreateCanvasDashboard() {
    if (!metricsViewName) return;
    await createCanvasDashboardFromMetricsView(instanceId, metricsViewName);
  }
</script>

{#if hasMenuItems}
  {#if referenceModelName}
    <NavigationMenuItem on:click={editModel}>
      <Model slot="icon" />
      Edit underlying model
    </NavigationMenuItem>
  {/if}
  <NavigationMenuItem on:click={viewGraph}>
    <GitBranch slot="icon" size="14px" />
    View DAG graph
  </NavigationMenuItem>
  {#if resource && $generateCanvas}
    <NavigationMenuItem
      disabled={!metricsViewName}
      on:click={handleCreateCanvasDashboard}
    >
      <CanvasIcon slot="icon" />
      <div class="flex gap-x-2 items-center">
        Generate Canvas Dashboard
        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </div>
    </NavigationMenuItem>
  {/if}
  {#if resource}
    <NavigationMenuItem
      on:click={() =>
        createAndPreviewExplore(
          runtimeClient,
          queryClient,
          instanceId,
          resource,
        )}
    >
      <ExploreIcon slot="icon" />
      <div class="flex gap-x-2 items-center">
        Generate Explore Dashboard
        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </div>
    </NavigationMenuItem>
  {/if}
{:else}
  <NavigationMenuItem on:click={viewGraph}>
    <GitBranch slot="icon" size="14px" />
    View DAG graph
  </NavigationMenuItem>
{/if}
