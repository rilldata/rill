<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import ExploreIcon from "../../../components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";

  const { ai } = featureFlags;
  const queryClient = useQueryClient();

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: modelHasError = fileArtifact.getHasErrors(
    queryClient,
    $runtime.instanceId,
  );
  $: modelQuery = fileArtifact.getResource(queryClient, $runtime.instanceId);
  $: connector = $modelQuery.data?.model?.spec?.outputConnector;
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $modelHasError || !modelIsIdle;
  $: tableName = $modelQuery.data?.model?.state?.resultTable ?? "";

  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    $runtime.instanceId,
    connector as string,
    "",
    "",
    tableName,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  $: createExploreFromTable = useCreateMetricsViewFromTableUIAction(
    $runtime.instanceId,
    connector as string,
    "",
    "",
    tableName,
    true,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
</script>

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
