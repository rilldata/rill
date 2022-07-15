<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import { syncExplore } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import type { Readable } from "svelte/store";
  import ExploreContainer from "./ExploreContainer.svelte";
  import ExploreHeader from "./ExploreHeader.svelte";
  import LeaderboardDisplay from "./leaderboards/LeaderboardDisplay.svelte";
  import MetricsTimeSeriesCharts from "./time-series-charts/MetricsTimeSeriesCharts.svelte";

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
</script>

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
      interval={$metricsLeaderboard?.selectedTimeRange?.interval ||
        $metricsLeaderboard?.timeRange?.interval}
    />
  </svelte:fragment>
  <svelte:fragment slot="leaderboards">
    <LeaderboardDisplay {columns} {whichReferenceValue} {metricsDefId} />
  </svelte:fragment>
</ExploreContainer>
