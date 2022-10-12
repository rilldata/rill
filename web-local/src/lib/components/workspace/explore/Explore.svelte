<script lang="ts">
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./time-series-charts/MetricsTimeSeriesCharts.svelte";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../application-state-stores/explorer-stores";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];
  $: selectedDimensionId = metricsExplorer?.selectedDimensionId;
</script>

<ExploreContainer let:columns>
  <ExploreHeader slot="header" {metricsDefId} />
  <MetricsTimeSeriesCharts slot="metrics" {metricsDefId} />
  {#if selectedDimensionId}
    <DimensionDisplay
      slot="leaderboards"
      {metricsDefId}
      dimensionId={selectedDimensionId}
    />
  {:else}
    <LeaderboardDisplay slot="leaderboards" {metricsDefId} />
  {/if}
</ExploreContainer>
