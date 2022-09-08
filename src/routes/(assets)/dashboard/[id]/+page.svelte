<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import ExploreContainer from "../../../_surfaces/workspace/explore/ExploreContainer.svelte";
  import ExploreHeader from "../../../_surfaces/workspace/explore/ExploreHeader.svelte";
  import LeaderboardDisplay from "../../../_surfaces/workspace/explore/leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "../../../_surfaces/workspace/explore/time-series-charts/MetricsTimeSeriesCharts.svelte";

  export let data;

  $: metricsDefId = data.metricsDefId;

  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsExplorer,
    metricsDefId,
  ]);
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<ExploreContainer let:columns>
  <svelte:fragment slot="header">
    <ExploreHeader {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="metrics">
    <MetricsTimeSeriesCharts {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="leaderboards">
    <LeaderboardDisplay {metricsDefId} />
  </svelte:fragment>
</ExploreContainer>
