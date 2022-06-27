<script lang="ts">
  import LeaderboardContainer from "./LeaderboardContainer.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardDisplay from "./LeaderboardDisplay.svelte";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard-slice";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import {
    initMetricsLeaderboard,
    singleMetricsLeaderboardSelector,
  } from "$lib/redux-store/metrics-leaderboard-slice";
  import {
    fetchManyDimensionsApi,
    manyDimensionsSelector,
  } from "$lib/redux-store/dimension-definition-slice";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

  export let metricsDefId: string;

  let metricsLeaderboard: MetricsLeaderboardEntity;
  $: metricsLeaderboard =
    singleMetricsLeaderboardSelector(metricsDefId)($reduxReadable);
  $: if (metricsDefId) {
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  let dimensions: Array<DimensionDefinitionEntity>;
  $: dimensions = manyDimensionsSelector(metricsDefId)($reduxReadable);

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
