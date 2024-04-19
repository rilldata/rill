<script lang="ts">
  import { goto } from "$app/navigation";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useCreateDashboardFromTableUIAction } from "../metrics-views/ai-generation/generateMetricsView";
  import { createModelFromSource } from "../sources/createModel";
  import { useIsModelingSupportedForCurrentOlapDriver } from "./selectors";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string;
  export let table: string;

  const queryClient = useQueryClient();

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: tableName = table;

  const handleCreateModel = async () => {
    try {
      const previousActiveEntity = getScreenNameFromPage();
      const [newModelPath, newModelName] = await createModelFromSource(
        queryClient,
        tableName,
        tableName,
        "models",
      );
      await goto(`/file/${newModelPath}`);
      await behaviourEvent.fireNavigationEvent(
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

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    connector,
    database,
    databaseSchema,
    tableName,
    "dashboards",
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
</script>

{#if $isModelingSupportedForCurrentOlapDriver.data}
  <NavigationMenuItem on:click={() => handleCreateModel()}>
    <Model slot="icon" />
    Create new model
  </NavigationMenuItem>
{/if}

<NavigationMenuItem on:click={createDashboardFromTable}>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard with AI
    <WandIcon class="w-3 h-3" />
  </div>
</NavigationMenuItem>
