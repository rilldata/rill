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
  import ExploreIcon from "../../../components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { featureFlags } from "../../feature-flags";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { createAndPreviewExplore } from "../../metrics-views/create-and-preview-explore";
  import { createResourceFile } from "../../file-explorer/new-files";
  import { ResourceKind } from "../../entity-management/resource-selectors";
  import { fileArtifacts } from "../../entity-management/file-artifacts";
  import { waitUntil } from "../../../lib/waitUtils";
  import { get } from "svelte/store";
  import {
    runtimeServicePutFile,
    runtimeServiceGenerateMetricsViewFile,
  } from "../../../runtime-client";
  import {
    createSqlModelFromTable,
    createYamlModelFromTable,
  } from "../code-utils";

  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;
  export let showGenerateMetricsAndDashboard: boolean = false;
  export let showGenerateModel: boolean = false;
  export let isModelingSupported: boolean | undefined = false;
  export let implementsOlap: boolean = false;

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
    if (implementsOlap) {
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
      await createModelAndMetricsViewAndExplore();
    }
  }

  // TODO: move to generateMetricsView.ts
  async function createModelAndMetricsViewAndExplore() {
    try {
      // Step 1: Create model that ingests from source to OLAP
      const [_modelPath, modelName] = await createYamlModelFromTable(
        queryClient,
        connector,
        database,
        databaseSchema,
        table,
      );

      // Step 2: Wait for model to be ready
      const modelResource = fileArtifacts
        .getFileArtifact(`/models/${modelName}.yaml`)
        .getResource(queryClient, instanceId);

      await waitUntil(() => get(modelResource).data !== undefined, 10000);

      // Step 3: Create metrics view using the backend AI generation
      // This will properly analyze the model's schema and generate dimensions/measures
      const metricsViewName = `${table}_metrics`;
      const metricsViewFilePath = `/metrics/${metricsViewName}.yaml`;

      // Use the backend function with the model name instead of table name
      await runtimeServiceGenerateMetricsViewFile(instanceId, {
        model: modelName, // Use model name instead of table
        path: metricsViewFilePath,
        useAi: get(featureFlags.ai),
      });

      // Step 4: Wait for metrics view to be ready
      const metricsViewResource = fileArtifacts
        .getFileArtifact(metricsViewFilePath)
        .getResource(queryClient, instanceId);

      await waitUntil(() => get(metricsViewResource).data !== undefined, 10000);

      const resource = get(metricsViewResource).data;
      if (!resource) {
        throw new Error("Failed to create a Metrics View resource");
      }

      // Step 5: Create explore dashboard
      await createAndPreviewExplore(queryClient, instanceId, resource);
    } catch (err) {
      console.error("Failed to create model and metrics view:", err);
      throw err;
    }
  }
</script>

{#if isModelingSupported || showGenerateModel}
  <NavigationMenuItem on:click={handleCreateModelFromTable}>
    <Model slot="icon" />
    Create model
  </NavigationMenuItem>
{/if}

{#if showGenerateMetricsAndDashboard}
  <NavigationMenuItem on:click={handleGenerateMetricsAndExplore}>
    <MetricsViewIcon slot="icon" />
    <div class="flex gap-x-2 items-center">
      Generate explorable metrics
      {#if $ai}
        with AI
        <WandIcon class="w-3 h-3" />
      {/if}
    </div>
  </NavigationMenuItem>
{/if}
