<script lang="ts">
  import LeaderboardContainer from "./LeaderboardContainer.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardDisplay from "./LeaderboardDisplay.svelte";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { initMetricsLeaderboard } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { selectDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
  import { singleMetricsLeaderboardSelector } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-selectors";

  export let metricsDefId: string;

  let metricsLeaderboard: MetricsLeaderboardEntity;
  $: metricsLeaderboard =
    singleMetricsLeaderboardSelector(metricsDefId)($reduxReadable);
  $: if (metricsDefId) {
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  let dimensions: Array<DimensionDefinitionEntity>;
  $: dimensions = selectDimensionsByMetricsId(metricsDefId)($reduxReadable);

  $: if (dimensions) {
    store.dispatch(initMetricsLeaderboard(metricsDefId, dimensions));
  }

  /** State for the reference value toggle */
  let whichReferenceValue: string;
  $: stagedReferenceValue =
    whichReferenceValue === "filtered"
      ? metricsLeaderboard?.bigNumber
      : metricsLeaderboard?.referenceValue;
</script>

<LeaderboardContainer let:columns>
  <LeaderboardHeader bind:whichReferenceValue {metricsDefId} />
  <LeaderboardDisplay
    {columns}
    referenceValue={stagedReferenceValue}
    {metricsDefId}
  />
</LeaderboardContainer>
