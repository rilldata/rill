<script lang="ts">
  import LeaderboardContainer from "./LeaderboardContainer.svelte";
  import LeaderboardHeader from "./LeaderboardHeader.svelte";
  import LeaderboardDisplay from "./LeaderboardDisplay.svelte";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard-slice";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { initMetricsLeaderboard } from "$lib/redux-store/metrics-leaderboard-slice";

  export let metricsDefId: string;
  let metricsLeaderboard: MetricsLeaderboardEntity;
  $: if (
    metricsDefId &&
    $reduxReadable?.metricsLeaderboard?.entities[metricsDefId]
  ) {
    metricsLeaderboard =
      $reduxReadable.metricsLeaderboard.entities[metricsDefId];
  }

  $: if (
    metricsDefId &&
    $reduxReadable?.metricsDefinition?.entities[metricsDefId]
  ) {
    store.dispatch(
      initMetricsLeaderboard(
        $reduxReadable?.metricsDefinition?.entities[metricsDefId]
      )
    );
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
