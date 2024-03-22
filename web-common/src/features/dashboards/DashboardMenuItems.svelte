<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    useDashboard,
    useDashboardFileNames,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
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
  import { createEventDispatcher } from "svelte";

  export let metricsViewName: string;

  const dispatch = createEventDispatcher();

  $: instanceId = $runtime.instanceId;
  $: dashboardNames = useDashboardFileNames(instanceId);
  $: dashboardQuery = useDashboard(instanceId, metricsViewName);

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
    await deleteFileArtifact(
      instanceId,
      metricsViewName,
      EntityType.MetricsDefinition,
      $dashboardNames?.data ?? [],
    );
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
<NavigationMenuSeparator />
<NavigationMenuItem on:click={() => dispatch("rename")}>
  <EditIcon slot="icon" />
  Rename...
</NavigationMenuItem>
<NavigationMenuItem on:click={deleteMetricsDef}>
  <Cancel slot="icon" />
  Delete
</NavigationMenuItem>
