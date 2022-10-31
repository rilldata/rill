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

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];
  $: selectedDimensionId = metricsExplorer?.selectedDimensionId;
</script>

<WorkspaceContainer bgClass="bg-white" inspector={false} assetID={metricsDefId}>
  <ExploreContainer slot="body" let:columns>
    <ExploreHeader slot="header" {metricsDefId} />
    <MetricsTimeSeriesCharts slot="metrics" {metricsDefId} />
    <svelte:fragment slot="leaderboards">
      {#if selectedDimensionId}
        <DimensionDisplay {metricsDefId} dimensionId={selectedDimensionId} />
      {:else}
        <LeaderboardDisplay {metricsDefId} />
      {/if}
    </svelte:fragment>
  </ExploreContainer>
</WorkspaceContainer>
