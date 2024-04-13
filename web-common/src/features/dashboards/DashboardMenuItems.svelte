<script lang="ts">
  import { goto } from "$app/navigation";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import {
    useDashboard,
    useDashboardRoutes,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { getNextRoute } from "../models/utils/navigate-to-next";

  export let metricsViewName: string;
  export let open: boolean;

  $: filePath = getFileAPIPathFromNameAndType(
    metricsViewName,
    EntityType.MetricsDefinition,
  );
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();
  const { customDashboards } = featureFlags;

  $: instanceId = $runtime.instanceId;
  $: dashboardQuery = useDashboard(instanceId, metricsViewName);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);
  $: dashboardRoutesQuery = useDashboardRoutes(instanceId);
  $: dashboardRoutes = $dashboardRoutesQuery.data ?? [];

  /**
   * Get the name of the dashboard's underlying model (if any).
   * Note that not all dashboards have an underlying model. Some dashboards are
   * underpinned by a source/table.
   */
  $: referenceModelName = $dashboardQuery?.data?.meta?.refs?.filter(
    (ref) => ref.kind === ResourceKind.Model,
  )?.[0]?.name;

  const editModel = async () => {
    if (!referenceModelName) return;
    const previousActiveEntity = $appScreen?.type;
    await goto(`/model/${referenceModelName}`);
    await behaviourEvent.fireNavigationEvent(
      referenceModelName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.Model,
    );
  };

  const editMetrics = async () => {
    await goto(`/dashboard/${metricsViewName}/edit`);

    const previousActiveEntity = $appScreen?.type;
    await behaviourEvent.fireNavigationEvent(
      metricsViewName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.MetricsDefinition,
    );
  };

  const deleteMetricsDef = async () => {
    try {
      await deleteFileArtifact(
        instanceId,
        filePath,
        EntityType.MetricsDefinition,
      );

      if (open) await goto(getNextRoute(dashboardRoutes));
    } catch (e) {
      console.error(e);
    }
  };
</script>

{#if referenceModelName}
  <NavigationMenuItem on:click={editModel}>
    <Model slot="icon" />
    Edit model
  </NavigationMenuItem>
{/if}
<NavigationMenuItem on:click={editMetrics}>
  <MetricsIcon slot="icon" />
  Edit metrics
</NavigationMenuItem>
{#if $customDashboards}
  <NavigationMenuItem
    on:click={() => {
      dispatch("generate-chart");
    }}
    disabled={$hasErrors}
  >
    <Explore slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate chart with AI
      <WandIcon class="w-3 h-3" />
    </div>
    <svelte:fragment slot="description">
      {#if $hasErrors}
        Dashboard has errors
      {/if}
    </svelte:fragment>
  </NavigationMenuItem>
{/if}
<NavigationMenuSeparator />
<NavigationMenuItem on:click={() => dispatch("rename")}>
  <EditIcon slot="icon" />
  Rename...
</NavigationMenuItem>
<NavigationMenuItem on:click={deleteMetricsDef}>
  <Cancel slot="icon" />
  Delete
</NavigationMenuItem>
