<script lang="ts">
  import { goto } from "$app/navigation";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
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
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { featureFlags } from "../../feature-flags";
  import { generateMetricsFromTable } from "../../metrics-views/ai-generation/generateMetricsView";
  import {
    createSqlModelFromTable,
    createYamlModelFromTable,
  } from "../code-utils";

  export let driver: string = "";
  export let connector: string;
  export let database: string = "";
  export let databaseSchema: string = "";
  export let table: string;
  export let showGenerateMetricsAndDashboard: boolean = false;
  export let showGenerateModel: boolean = false;
  export let isModelingSupported: boolean | undefined = false;
  export let metricsMode: "import" | "live" | "both" = "import";

  const client = useRuntimeClient();
  const { ai } = featureFlags;
  $: ({ instanceId } = client);

  // Display label for the live-OLAP path in the submenu.
  const DRIVER_DISPLAY_NAMES: Record<string, string> = {
    snowflake: "Snowflake",
    bigquery: "BigQuery",
  };
  $: liveDriverLabel = DRIVER_DISPLAY_NAMES[driver] ?? driver;

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
          client,
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
          client,
          queryClient,
          connector,
          database,
          databaseSchema,
          table,
        ),
      );
    }
  }

  async function handleGenerateMetrics(isOlapConnector: boolean) {
    await generateMetricsFromTable(
      client,
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
      false, // Don't create explore dashboard
      isOlapConnector,
    );
  }

  async function handleGenerateDashboard(isOlapConnector: boolean) {
    await generateMetricsFromTable(
      client,
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

{#if metricsMode !== "both" && (isModelingSupported || showGenerateModel)}
  <NavigationMenuItem onclick={handleCreateModelFromTable}>
    <Model slot="icon" />
    Create model
  </NavigationMenuItem>
{/if}

{#if showGenerateMetricsAndDashboard}
  {#if metricsMode === "both"}
    <DropdownMenu.Group>
      <DropdownMenu.Label>Import to DuckDB</DropdownMenu.Label>
      {#if isModelingSupported || showGenerateModel}
        <NavigationMenuItem onclick={handleCreateModelFromTable}>
          <Model slot="icon" />
          Create model
        </NavigationMenuItem>
      {/if}
      <NavigationMenuItem onclick={() => handleGenerateMetrics(false)}>
        <MetricsViewIcon slot="icon" />
        <div class="flex gap-x-2 items-center">
          Generate metrics
          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </div>
      </NavigationMenuItem>
      <NavigationMenuItem onclick={() => handleGenerateDashboard(false)}>
        <ExploreIcon slot="icon" />
        <div class="flex gap-x-2 items-center">
          Generate dashboard
          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </div>
      </NavigationMenuItem>
    </DropdownMenu.Group>

    <DropdownMenu.Separator />
    <DropdownMenu.Group>
      <DropdownMenu.Label>Live on {liveDriverLabel}</DropdownMenu.Label>
      <NavigationMenuItem onclick={() => handleGenerateMetrics(true)}>
        <MetricsViewIcon slot="icon" />
        <div class="flex gap-x-2 items-center">
          Generate metrics
          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </div>
      </NavigationMenuItem>
      <NavigationMenuItem onclick={() => handleGenerateDashboard(true)}>
        <ExploreIcon slot="icon" />
        <div class="flex gap-x-2 items-center">
          Generate dashboard
          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </div>
      </NavigationMenuItem>
    </DropdownMenu.Group>
  {:else}
    {@const isLive = metricsMode === "live"}
    <NavigationMenuItem onclick={() => handleGenerateMetrics(isLive)}>
      <MetricsViewIcon slot="icon" />
      <div class="flex gap-x-2 items-center">
        Generate metrics
        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </div>
    </NavigationMenuItem>
    <NavigationMenuItem onclick={() => handleGenerateDashboard(isLive)}>
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
{/if}
