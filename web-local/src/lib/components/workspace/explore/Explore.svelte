<script lang="ts">
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

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: selectedDimensionName = metricsExplorer?.selectedDimensionName;
</script>

<WorkspaceContainer
  bgClass="bg-white"
  inspector={false}
  assetID={metricViewName}
>
  <ExploreContainer slot="body">
    <ExploreHeader slot="header" {metricViewName} />
    <MetricsTimeSeriesCharts slot="metrics" {metricViewName} />
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
