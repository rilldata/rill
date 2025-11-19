<script lang="ts">
  import { goto } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import ExploreIcon from "../../../components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import Model from "../../../components/icons/Model.svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getScreenNameFromPage } from "../../file-explorer/telemetry";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createSqlModelFromTable } from "../../connectors/code-utils";

  const { ai } = featureFlags;
  const queryClient = useQueryClient();

  export let filePath: string;

  $: ({ instanceId } = $runtime);

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: modelHasError = fileArtifact.getHasErrors(queryClient, instanceId);
  $: modelQuery = fileArtifact.getResource(queryClient, instanceId);
  $: connector = $modelQuery.data?.model?.spec?.outputConnector;
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $modelHasError || !modelIsIdle;
  $: tableName = $modelQuery.data?.model?.state?.resultTable ?? "";

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
</script>

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
    Generate metrics
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

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createExploreFromTable}
>
  <ExploreIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard
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

<NavigationMenuSeparator />
