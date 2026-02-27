<script lang="ts">
  import { goto } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { openResourceGraphQuickView } from "@rilldata/web-common/features/resource-graph/quick-view/quick-view-store";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { GitBranch, WandIcon } from "lucide-svelte";
  import CanvasIcon from "../../../components/icons/CanvasIcon.svelte";
  import ExploreIcon from "../../../components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import Model from "../../../components/icons/Model.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { createSqlModelFromTable } from "../../connectors/code-utils";
  import { getScreenNameFromPage } from "../../file-explorer/telemetry";
  import {
    createCanvasDashboardFromTableWithAgent,
    useCreateMetricsViewFromTableUIAction,
    useCreateMetricsViewWithCanvasUIAction,
  } from "../../metrics-views/ai-generation/generateMetricsView";

  const { ai, generateCanvas, developerChat } = featureFlags;
  const queryClient = useQueryClient();

  export let filePath: string;

  $: ({ instanceId } = $runtime);

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: modelHasError = fileArtifact.getHasErrors(queryClient, instanceId);
  $: modelQuery = fileArtifact.getResource(queryClient, instanceId);
  $: modelResource = $modelQuery.data;
  $: connector = modelResource?.model?.spec?.outputConnector;
  $: modelIsIdle =
    modelResource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $modelHasError || !modelIsIdle;
  $: tableName = modelResource?.model?.state?.resultTable ?? "";

  function viewGraph() {
    if (!modelResource) {
      console.warn(
        "[ModelMenuItems] Cannot open resource graph: resource unavailable.",
      );
      return;
    }
    openResourceGraphQuickView(modelResource);
  }

  async function handleCreateModel() {
    try {
      const previousActiveEntity = getScreenNameFromPage();
      const addDevLimit = false; // Typically, the `dev` limit would be applied on the Source itself
      const [newModelPath, newModelName] = await createSqlModelFromTable(
        queryClient,
        connector as string,
        "",
        "",
        tableName,
        addDevLimit,
      );

      await goto(`/files${newModelPath}`);

      await behaviourEvent?.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousActiveEntity,
        MetricsEventScreenName.Model,
      );
    } catch (err) {
      console.error(err);
    }
  }

  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector as string,
    "",
    "",
    tableName,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  $: createExploreFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector as string,
    "",
    "",
    tableName,
    true,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  $: createCanvasDashboardFromTable = useCreateMetricsViewWithCanvasUIAction(
    instanceId,
    connector as string,
    "",
    "",
    tableName,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  async function onGenerateCanvasDashboard() {
    // Use developer agent if enabled, otherwise fall back to RPC
    if ($developerChat) {
      await createCanvasDashboardFromTableWithAgent(
        instanceId,
        connector as string,
        "",
        "",
        tableName,
      );
    } else {
      await createCanvasDashboardFromTable();
    }
  }
</script>

<NavigationMenuItem on:click={viewGraph}>
  <GitBranch slot="icon" size="14px" />
  View DAG graph
</NavigationMenuItem>

<NavigationMenuItem on:click={handleCreateModel}>
  <Model slot="icon" />
  Create new model
</NavigationMenuItem>

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createMetricsViewFromTable}
>
  <MetricsViewIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate Metrics View
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
  <svelte:fragment slot="description">
    {#if $modelHasError}
      Model has errors
    {:else if !modelIsIdle}
      Dependencies are being reconciled.
    {/if}
  </svelte:fragment>
</NavigationMenuItem>

{#if $generateCanvas}
  <NavigationMenuItem
    disabled={disableCreateDashboard}
    on:click={onGenerateCanvasDashboard}
  >
    <CanvasIcon slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate Canvas Dashboard
      {#if $ai}
        with AI
        <WandIcon class="w-3 h-3" />
      {/if}
    </div>
    <svelte:fragment slot="description">
      {#if $modelHasError}
        Model has errors
      {:else if !modelIsIdle}
        Dependencies are being reconciled.
      {/if}
    </svelte:fragment>
  </NavigationMenuItem>
{/if}

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createExploreFromTable}
>
  <ExploreIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate Explore Dashboard
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
  <svelte:fragment slot="description">
    {#if $modelHasError}
      Model has errors
    {:else if !modelIsIdle}
      Dependencies are being reconciled.
    {/if}
  </svelte:fragment>
</NavigationMenuItem>
