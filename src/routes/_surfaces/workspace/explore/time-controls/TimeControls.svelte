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
  import { useMetaQuery } from "$lib/svelte-query/queries/metrics-view";
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
  let selectedTimeGrain;

  // query the `/meta` endpoint to get the all time range of the dataset
  $: metaQuery = useMetaQuery(metricsDefId);
  $: allTimeRange = $metaQuery.data?.timeDimension?.timeRange;

  const initializeState = (metricsExplorer: MetricsExplorerEntity) => {
    if (
      metricsExplorer?.selectedTimeRange?.name &&
      metricsExplorer?.selectedTimeRange?.interval
    ) {
      selectedTimeRangeName = metricsExplorer.selectedTimeRange.name;
      selectedTimeGrain = metricsExplorer.selectedTimeRange.interval;
    } else {
      selectedTimeRangeName = getDefaultTimeRangeName();
      selectedTimeGrain = getDefaultTimeGrain(
        selectedTimeRangeName,
        allTimeRange
      );
    }
  };
  $: initializeState(metricsExplorer);

  const setSelectedTimeRangeName = (evt) => {
    selectedTimeRangeName = evt.detail.timeRangeName;
  };
  const setSelectedTimeGrain = (evt) => {
    selectedTimeGrain = evt.detail.timeGrain;
  };

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
