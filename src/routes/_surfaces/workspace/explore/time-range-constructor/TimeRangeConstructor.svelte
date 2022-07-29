<!--
@component
Constructs a TimeRange object – to be used as the filter in MetricsExplorer – by taking as input:
- the time range name (a semantic understanding of the time range, like "Last 6 Hours" or "Last 30 days")
- the time grain (e.g., "hour" or "day")
- the dataset's full time range (so its end time can be used in relative time ranges)
-->
<script lang="ts">
  import type {
    TimeGrain,
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import { setExploreSelectedTimeRangeAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { store } from "$lib/redux-store/store-root";
  import _ from "lodash";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeNameSelector from "./TimeRangeNameSelector.svelte";
  import { makeTimeRange } from "./timeRangeUtils";

  export let metricsDefId: string;

  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectedTimeRangeName;
  const setSelectedTimeRangeName = (timeRangeName: TimeRangeName) => {
    selectedTimeRangeName = timeRangeName;
  };

  let selectedTimeGrain;
  const setSelectedTimeGrain = (timeGrain: TimeGrain) => {
    selectedTimeGrain = timeGrain;
  };

  const constructTimeRangeAndUpdateStore = (
    timeRangeName: TimeRangeName,
    timeGrain: TimeGrain,
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    if (!timeRangeName || !timeGrain || !allTimeRangeInDataset) return;

    const newTimeRange = makeTimeRange(
      selectedTimeRangeName,
      selectedTimeGrain,
      $metricsExplorer.allTimeRange
    );

    if (_.isEqual(newTimeRange, $metricsExplorer.selectedTimeRange)) return;

    setExploreSelectedTimeRangeAndUpdate(store.dispatch, metricsDefId, {
      name: newTimeRange.name,
      start: newTimeRange.start,
      end: newTimeRange.end,
      interval: newTimeRange.interval,
    });
  };

  // reactive statement that makes a new time range whenever the selected options change
  $: constructTimeRangeAndUpdateStore(
    selectedTimeRangeName,
    selectedTimeGrain,
    $metricsExplorer.allTimeRange
  );
</script>

<div class="flex flex-row">
  <TimeRangeNameSelector
    {metricsDefId}
    {selectedTimeRangeName}
    onSelectTimeRangeName={setSelectedTimeRangeName}
  />
  <TimeGrainSelector
    {metricsDefId}
    {selectedTimeRangeName}
    {selectedTimeGrain}
    onSelectTimeGrain={setSelectedTimeGrain}
  />
</div>
