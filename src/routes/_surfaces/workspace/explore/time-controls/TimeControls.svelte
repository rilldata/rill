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
  import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
    invalidateMetricViewData,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQuery, useQueryClient } from "@sveltestack/svelte-query";
  import { onMount } from "svelte";
  import {
    getDefaultTimeGrain,
    getDefaultTimeRangeName,
    makeTimeRange,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeNameSelector from "./TimeRangeNameSelector.svelte";
  import { MetricsExplorerStore } from "$lib/application-state-stores/explorer-stores";

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $MetricsExplorerStore.entities[metricsDefId];

  let selectedTimeRangeName;
  const setSelectedTimeRangeName = (evt) => {
    selectedTimeRangeName = evt.detail.timeRangeName;
  };

  let selectedTimeGrain;
  const setSelectedTimeGrain = (evt) => {
    selectedTimeGrain = evt.detail.timeGrain;
  };

  // query the `/meta` endpoint to get the all time range of the dataset
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }
  let allTimeRange: TimeSeriesTimeRange;
  $: allTimeRange = $queryResult.data?.timeDimension?.timeRange;

  const queryClient = useQueryClient();

  // initialize the component with the default options
  onMount(() => {
    const defaultTimeRangeName = getDefaultTimeRangeName();
    selectedTimeRangeName = defaultTimeRangeName;
    const defaultTimeGrain = getDefaultTimeGrain(
      selectedTimeRangeName,
      allTimeRange
    );
    selectedTimeGrain = defaultTimeGrain;
  });

  const makeTimeRangeAndUpdateStore = (
    timeRangeName: TimeRangeName,
    timeGrain: TimeGrain,
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    if (!timeRangeName || !timeGrain || !allTimeRangeInDataset) return;

    const newTimeRange = makeTimeRange(
      selectedTimeRangeName,
      selectedTimeGrain,
      allTimeRange
    );

    if (
      newTimeRange.start === metricsExplorer?.selectedTimeRange?.start &&
      newTimeRange.end === metricsExplorer?.selectedTimeRange?.end &&
      newTimeRange.interval === metricsExplorer?.selectedTimeRange?.interval
    )
      return;

    MetricsExplorerStore.setSelectedTimeRange(metricsDefId, newTimeRange);

    invalidateMetricViewData(queryClient, metricsDefId);
  };

  // reactive statement that makes a new time range whenever the selected options change
  $: makeTimeRangeAndUpdateStore(
    selectedTimeRangeName,
    selectedTimeGrain,
    allTimeRange
  );
</script>

<div class="flex flex-row">
  <TimeRangeNameSelector
    {metricsDefId}
    {selectedTimeRangeName}
    on:select-time-range-name={setSelectedTimeRangeName}
  />
  <TimeGrainSelector
    {metricsDefId}
    {selectedTimeRangeName}
    {selectedTimeGrain}
    on:select-time-grain={setSelectedTimeGrain}
  />
</div>
