<script lang="ts">
  import { goto } from "$app/navigation";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { useModelFileNames } from "@rilldata/web-common/features/models/selectors";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useDashboardNames } from "../dashboards/selectors";
  import { createModelFromSource } from "../sources/createModel";
  import { OLAP_DRIVERS_WITHOUT_MODELING } from "./config";
  import { createDashboardFromTable } from "./createDashboardFromTable";

  export let fullyQualifiedTableName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const queryClient = useQueryClient();

  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: tableName = fullyQualifiedTableName.split(".")[1];
  $: runtimeInstanceId = $runtime.instanceId;
  $: modelNames = useModelFileNames($runtime.instanceId);
  $: dashboardNames = useDashboardNames($runtime.instanceId);

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = $appScreen?.type;
      const newModelName = await createModelFromSource(
        runtimeInstanceId,
        $modelNames.data ?? [],
        tableName,
        tableName,
      );

      behaviourEvent.fireNavigationEvent(
        newModelName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousActiveEntity,
        MetricsEventScreenName.Model,
      );
    } catch (err) {
      console.error(err);
    }
  };

  async function handleCreateDashboardFromTable(tableName: string) {
    overlay.set({
      title: "Creating a dashboard for " + tableName,
    });
    const newDashboardName = await createDashboardFromTable(
      queryClient,
      tableName,
      $dashboardNames.data ?? [],
    );
    goto(`/dashboard/${newDashboardName}`);
    const previousActiveEntity = $appScreen?.type;
    behaviourEvent.fireNavigationEvent(
      newDashboardName,
      BehaviourEventMedium.Menu,
      MetricsEventSpace.LeftPanel,
      previousActiveEntity,
      MetricsEventScreenName.Dashboard,
    );
    overlay.set(null);
    toggleMenu(); // unmount component
  }
</script>

{#if olapConnector && !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapConnector)}
  <MenuItem icon on:select={() => handleCreateModel()}>
    <Model slot="icon" />
    Create new model
  </MenuItem>
{/if}

<MenuItem
  icon
  on:select={() => handleCreateDashboardFromTable(tableName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
</MenuItem>
