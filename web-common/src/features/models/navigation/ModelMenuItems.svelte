<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { useModel, useModelRoutes } from "../selectors";
  import { goto } from "$app/navigation";
  import { getNextRoute } from "../utils/navigate-to-next";

  export let modelName: string;
  export let open: boolean;

  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: fileArtifact = fileArtifacts.getFileArtifact(modelPath);

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const { customDashboards } = featureFlags;

  $: modelRoutesQuery = useModelRoutes($runtime.instanceId);
  $: modelRoutes = $modelRoutesQuery.data ?? [];
  $: modelHasError = fileArtifact.getHasErrors(
    queryClient,
    $runtime.instanceId,
  );
  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: connector = $modelQuery.data?.model?.spec?.connector;
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $modelHasError || !modelIsIdle;

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    connector as string,
    "",
    "",
    modelName,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  const handleDeleteModel = async (modelName: string) => {
    try {
      await deleteFileArtifact(
        $runtime.instanceId,
        getFileAPIPathFromNameAndType(modelName, EntityType.Model),
        EntityType.MetricsDefinition,
      );

      if (open) await goto(getNextRoute(modelRoutes));
    } catch (e) {
      console.error(e);
    }
  };
</script>

<NavigationMenuItem
  disabled={disableCreateDashboard}
  on:click={createDashboardFromTable}
>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard with AI
    <WandIcon class="w-3 h-3" />
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
        table: $modelQuery.data?.model?.state?.table,
        connector: $modelQuery.data?.model?.state?.connector,
      });
    }}
  >
    <Explore slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate chart with AI
      <WandIcon class="w-3 h-3" />
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

<NavigationMenuSeparator />

<NavigationMenuItem
  on:click={() => {
    dispatch("rename-asset");
  }}
>
  <EditIcon slot="icon" />
  Rename...
</NavigationMenuItem>
<NavigationMenuItem on:click={() => handleDeleteModel(modelName)}>
  <Cancel slot="icon" />
  Delete
</NavigationMenuItem>
