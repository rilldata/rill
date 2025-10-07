<script lang="ts">
  import { goto } from "$app/navigation";
  import Model from "@rilldata/web-common/components/icons/Model.svelte";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { WandIcon } from "lucide-svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import {
    useCreateMetricsViewFromTableUIAction,
    createModelAndMetricsViewAndExplore as createModelAndMetricsViewAndExploreFromTable,
  } from "../../metrics-views/ai-generation/generateMetricsView";
  import {
    createSqlModelFromTable,
    createYamlModelFromTable,
  } from "../code-utils";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;
  export let showGenerateMetricsAndDashboard: boolean = false;
  export let showGenerateModel: boolean = false;
  export let isModelingSupported: boolean | undefined = false;
  export let isOlapConnector: boolean = false;

  const { ai } = featureFlags;

  $: ({ instanceId } = $runtime);
  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector,
    database,
    databaseSchema,
    table,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
  $: createExploreFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector,
    database,
    databaseSchema,
    table,
    true,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  async function handleCreateModel(
    modelCreationFn: () => Promise<[string, string]>,
  ) {
    try {
      const previousActiveEntity = getScreenNameFromPage();
      const [newModelPath, newModelName] = await modelCreationFn();
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

  async function handleCreateModelFromTable() {
    if (isModelingSupported) {
      await handleCreateModel(() =>
        createSqlModelFromTable(
          queryClient,
          connector,
          database,
          databaseSchema,
          table,
        ),
      );
    } else if (showGenerateModel) {
      await handleCreateModel(() =>
        createYamlModelFromTable(
          queryClient,
          connector,
          database,
          databaseSchema,
          table,
        ),
      );
    }
  }

  // Create both metrics view and explore dashboard
  async function handleGenerateMetricsAndExplore() {
    if (isOlapConnector) {
      // For OLAP connectors, create both in parallel
      await Promise.all([
        createMetricsViewFromTable(),
        createExploreFromTable(),
      ]);
    } else {
      // For non-OLAP connectors, follow Rill architecture:
      // 1. Create model (ingests from source → OLAP)
      // 2. Create metrics view (on top of model)
      // 3. Create explore dashboard (on top of metrics view)
      await createModelAndMetricsViewAndExploreFromTable(
        instanceId,
        connector,
        database,
        databaseSchema,
        table,
      );
    }
  }
</script>

{#if isModelingSupported || showGenerateModel}
  <NavigationMenuItem on:click={handleCreateModelFromTable}>
    <Model slot="icon" />
    Create model
  </NavigationMenuItem>
{/if}

{#if isOlapConnector}
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
{/if}

{#if showGenerateMetricsAndDashboard && !isOlapConnector}
  <NavigationMenuItem on:click={handleGenerateMetricsAndExplore}>
    <ExploreIcon slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate dashboard
      {#if $ai}
        with AI
        <WandIcon class="w-3 h-3" />
      {/if}
    </div>
  </NavigationMenuItem>
{/if}
