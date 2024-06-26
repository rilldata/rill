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
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createModelFromTable } from "./createModel";
  import { useIsModelingSupportedForCurrentOlapDriver } from "./selectors";

  const { ai } = featureFlags;
  const queryClient = useQueryClient();

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    connector,
    database,
    databaseSchema,
    table,
    "dashboards",
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  async function handleCreateModel() {
    try {
      const previousActiveEntity = getScreenNameFromPage();
      const [newModelPath, newModelName] = await createModelFromTable(
        queryClient,
        connector,
        database,
        databaseSchema,
        table,
      );
      await goto(`/files${newModelPath}`);
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
  }
</script>

{#if $isModelingSupportedForCurrentOlapDriver}
  <NavigationMenuItem on:click={handleCreateModel}>
    <Model slot="icon" />
    Create new model
  </NavigationMenuItem>
{/if}

<NavigationMenuItem on:click={createDashboardFromTable}>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
</NavigationMenuItem>
