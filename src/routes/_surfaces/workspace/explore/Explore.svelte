<script lang="ts">
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import LeaderboardDisplay from "./LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./MetricsTimeSeriesCharts.svelte";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import { store } from "$lib/redux-store/store-root";
  import { initMetricsLeaderboard } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { getMetricsLeaderboardById } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-readables";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import type { Readable } from "svelte/store";

  export let metricsDefId: string;

  let metricsLeaderboard: Readable<MetricsLeaderboardEntity>;
  $: metricsLeaderboard = getMetricsLeaderboardById(metricsDefId);
  $: if (metricsDefId) {
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  let dimensions: Readable<Array<DimensionDefinitionEntity>>;
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  $: if ($dimensions) {
    store.dispatch(initMetricsLeaderboard(metricsDefId, $dimensions));
  }

  /** State for the reference value toggle */
  let whichReferenceValue: string;
  $: stagedReferenceValue =
    whichReferenceValue === "filtered"
      ? $metricsLeaderboard?.bigNumber
      : $metricsLeaderboard?.referenceValue;
</script>

<ExploreContainer let:columns>
  <svelte:fragment slot="header">
    <ExploreHeader bind:whichReferenceValue {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="metrics">
    <MetricsTimeSeriesCharts {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="leaderboards">
    <LeaderboardDisplay
      {columns}
      referenceValue={stagedReferenceValue}
      {metricsDefId}
    />
  </svelte:fragment>
</ExploreContainer>
