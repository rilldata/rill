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
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import {
    invalidateMetricViewData,
    useGetMetricViewMeta,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { onMount } from "svelte";
  import {
    getDefaultTimeGrain,
    getDefaultTimeRangeName,
    getSelectableTimeGrains,
    makeTimeRange,
    TimeGrainOption,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeNameSelector from "./TimeRangeNameSelector.svelte";

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let selectedTimeRangeName;
  const setSelectedTimeRangeName = (evt) => {
    selectedTimeRangeName = evt.detail.timeRangeName;
  };

  let selectedTimeGrain;
  const setSelectedTimeGrain = (evt) => {
    selectedTimeGrain = evt.detail.timeGrain;
  };

  // query the `/meta` endpoint to get the all time range of the dataset
  $: metaQuery = useGetMetricViewMeta(metricsDefId);
  $: allTimeRange = $metaQuery.data?.timeDimension?.timeRange;

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

  // we get the selectableTimeGrains so that we can assess whether or not the
  // existing selectedTimeGrain is valid whenever the selectedTimeRangeName changes
  let selectableTimeGrains: TimeGrainOption[];
  $: selectableTimeGrains = getSelectableTimeGrains(
    selectedTimeRangeName,
    allTimeRange
  );

  const checkValidTimeGrain = (timeGrain: TimeGrain) => {
    const timeGrainOption = selectableTimeGrains.find(
      (timeGrainOption) => timeGrainOption.timeGrain === timeGrain
    );
    return timeGrainOption?.enabled;
  };

  const queryClient = useQueryClient();

  const makeValidTimeRangeAndUpdateAppState = (
    timeRangeName: TimeRangeName,
    timeGrain: TimeGrain,
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    if (!timeRangeName || !timeGrain || !allTimeRangeInDataset) return;

    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    const isValidTimeGrain = checkValidTimeGrain(timeGrain);
    if (!isValidTimeGrain) {
      selectedTimeGrain = getDefaultTimeGrain(
        timeRangeName,
        allTimeRangeInDataset
      );
    }

    const newTimeRange = makeTimeRange(
      selectedTimeRangeName,
      selectedTimeGrain,
      allTimeRange
    );

    // don't update if time range hasn't changed
    if (
      newTimeRange.start === metricsExplorer?.selectedTimeRange?.start &&
      newTimeRange.end === metricsExplorer?.selectedTimeRange?.end &&
      newTimeRange.interval === metricsExplorer?.selectedTimeRange?.interval
    )
      return;

    metricsExplorerStore.setSelectedTimeRange(metricsDefId, newTimeRange);
    invalidateMetricViewData(queryClient, metricsDefId);
  };

  // reactive statement that makes a new valid time range whenever the selected options change
  $: makeValidTimeRangeAndUpdateAppState(
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
    {selectedTimeGrain}
    {selectableTimeGrains}
    on:select-time-grain={setSelectedTimeGrain}
  />
</div>
