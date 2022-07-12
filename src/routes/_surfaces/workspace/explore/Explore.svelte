<script lang="ts">
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import LeaderboardDisplay from "./LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./MetricsTimeSeriesCharts.svelte";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import type { Readable } from "svelte/store";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import {
    setExploreSelectedTimeRangeAndUpdate,
    syncExplore,
  } from "$lib/redux-store/explore/explore-apis";

  export let metricsDefId: string;

  let metricsLeaderboard: Readable<MetricsExploreEntity>;
  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);
  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }

  let measures: Readable<Array<MeasureDefinitionEntity>>;
  $: measures = getMeasuresByMetricsId(metricsDefId);

  let dimensions: Readable<Array<DimensionDefinitionEntity>>;
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  $: syncExplore(
    store.dispatch,
    metricsDefId,
    $metricsLeaderboard,
    $dimensions,
    $measures
  );

  let whichReferenceValue: string;

  let switcher = false;
</script>

<button
  on:click={() => {
    switcher = !switcher;
    setExploreSelectedTimeRangeAndUpdate(store.dispatch, metricsDefId, {
      start: new Date(
        switcher ? "2017-05-05" : $metricsLeaderboard.timeRange.start
      ).toISOString(),
      end: new Date(
        switcher ? "2018-05-05" : $metricsLeaderboard.timeRange.end
      ).toISOString(),
    });
  }}
>
  {switcher}
</button>

<ExploreContainer let:columns>
  <svelte:fragment slot="header">
    <ExploreHeader bind:whichReferenceValue {metricsDefId} />
  </svelte:fragment>
  <svelte:fragment slot="metrics">
    <MetricsTimeSeriesCharts
      start={$metricsLeaderboard?.selectedTimeRange?.start ||
        $metricsLeaderboard?.timeRange?.start}
      end={$metricsLeaderboard?.selectedTimeRange?.end ||
        $metricsLeaderboard?.timeRange?.end}
      activeMeasureIds={$measures?.map((measure) => measure.id) || []}
      {metricsDefId}
    />
  </svelte:fragment>
  <svelte:fragment slot="leaderboards">
    <LeaderboardDisplay {columns} {whichReferenceValue} {metricsDefId} />
  </svelte:fragment>
</ExploreContainer>
