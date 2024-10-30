<script lang="ts">
  import { goto } from "$app/navigation";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { WandIcon } from "lucide-svelte";
  import ExploreIcon from "../../../components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createModelFromTable } from "./createModel";
  import { useIsModelingSupportedForOlapDriver } from "./selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  const { ai } = featureFlags;

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;

  $: isModelingSupportedForOlapDriver = useIsModelingSupportedForOlapDriver(
    $runtime.instanceId,
    connector,
  );
  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    $runtime.instanceId,
    connector,
    database,
    databaseSchema,
    table,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
  $: createExploreFromTable = useCreateMetricsViewFromTableUIAction(
    $runtime.instanceId,
    connector,
    database,
    databaseSchema,
    table,
    true,
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

{#if $isModelingSupportedForOlapDriver}
  <NavigationMenuItem on:click={handleCreateModel}>
    <Model slot="icon" />
    Create new model
  </NavigationMenuItem>
{/if}

<NavigationMenuItem on:click={createMetricsViewFromTable}>
  <MetricsViewIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate metrics
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
</NavigationMenuItem>

<NavigationMenuItem on:click={createExploreFromTable}>
  <ExploreIcon slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard
    {#if $ai}
      with AI
      <WandIcon class="w-3 h-3" />
    {/if}
  </div>
</NavigationMenuItem>
