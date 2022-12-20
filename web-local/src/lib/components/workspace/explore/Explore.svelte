<script lang="ts">
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../application-state-stores/explorer-stores";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./time-series-charts/MetricsTimeSeriesCharts.svelte";

  export let metricViewName: string;

  const switchToMetrics = async (metricViewName: string) => {
    if (!metricViewName) return;

    appStore.setActiveEntity(metricViewName, EntityType.MetricsExplorer);
  };

  $: switchToMetrics(metricViewName);

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;
</script>

<WorkspaceContainer
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <ExploreContainer {metricViewName} slot="body">
    <ExploreHeader {metricViewName} slot="header" />
    <MetricsTimeSeriesCharts {metricViewName} slot="metrics" />
    <svelte:fragment slot="leaderboards">
      {#if selectedDimensionName}
        <DimensionDisplay
          {metricViewName}
          dimensionName={selectedDimensionName}
        />
      {:else}
        <LeaderboardDisplay {metricViewName} />
      {/if}
    </svelte:fragment>
  </ExploreContainer>
</WorkspaceContainer>
