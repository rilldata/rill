<script lang="ts">
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  export let filePath: string;

  const { ai } = featureFlags;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const { customDashboards } = featureFlags;

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

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    connector as string,
    "",
    "",
    tableName,
    "dashboards",
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
</script>

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createDashboardFromTable}
>
  <Explore slot="icon" />
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
{#if $customDashboards}
  <NavigationMenuItem
    disabled={disableCreateDashboard}
    on:click={() => {
      dispatch("generate-chart", {
        table: $modelQuery.data?.model?.state?.resultTable,
        connector: $modelQuery.data?.model?.state?.resultConnector,
      });
    }}
  >
    <Explore slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate Chart
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
