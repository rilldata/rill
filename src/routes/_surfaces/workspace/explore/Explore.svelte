<script lang="ts">
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./time-series-charts/MetricsTimeSeriesCharts.svelte";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import DimensionDisplay from "./leaderboards/DimensionDisplay.svelte";

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];
  $: selectedDimensionId = metricsExplorer?.selectedDimensionId;
</script>

<ExploreContainer let:columns>
  <svelte:fragment slot="header">
    <ExploreHeader {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="metrics">
    <MetricsTimeSeriesCharts {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="leaderboards">
    {#if selectedDimensionId}
      <DimensionDisplay {metricsDefId} dimensionId={selectedDimensionId} />
    {:else}
      <LeaderboardDisplay {metricsDefId} />
    {/if}
  </svelte:fragment>
</ExploreContainer>
