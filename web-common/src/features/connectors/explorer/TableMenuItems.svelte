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
  import { runtime } from "../../../runtime-client/runtime-store";
  import { generateMetricsFromTable } from "../../metrics-views/ai-generation/generateMetricsView";
  import {
    createSqlModelFromTable,
    createYamlModelFromTable,
  } from "../code-utils";
  import GenerateMenuItem from "./GenerateMenuItem.svelte";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;
  export let showGenerateMetricsAndDashboard: boolean = false;
  export let showGenerateModel: boolean = false;
  export let isModelingSupported: boolean | undefined = false;
  export let isOlapConnector: boolean = false;

  $: ({ instanceId } = $runtime);

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

  async function handleGenerateMetrics() {
    await generateMetricsFromTable(
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
      false, // Don't create explore dashboard
      isOlapConnector,
    );
  }

  async function handleGenerateDashboard() {
    await generateMetricsFromTable(
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
      true, // Create explore dashboard
      isOlapConnector,
    );
  }
</script>

{#if isModelingSupported || showGenerateModel}
  <NavigationMenuItem on:click={handleCreateModelFromTable}>
    <Model slot="icon" />
    Create model
  </NavigationMenuItem>
{/if}

{#if isOlapConnector || showGenerateMetricsAndDashboard}
  <GenerateMenuItem type="metrics" onClick={handleGenerateMetrics} />
  <GenerateMenuItem type="dashboard" onClick={handleGenerateDashboard} />
{/if}
