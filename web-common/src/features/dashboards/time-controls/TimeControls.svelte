<!--
@component
Constructs a TimeSeriesTimeRange object – to be used as the filter in MetricsExplorer – by taking as input:
- a base time range
- a time grain (e.g., "hour" or "day")
- the dataset's full time range (so its end time can be used in relative time ranges)

We should rename TimeSeriesTimeRange to a better name.
-->
<script lang="ts">
  import {
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    TimeGrain,
    TimeRange,
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
  import { useRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import NoTimeDimensionCTA from "./NoTimeDimensionCTA.svelte";
  import {
    addGrains,
    checkValidTimeGrain,
    floorDate,
    getDefaultTimeGrain,
    getDefaultTimeRange,
    getTimeGrainOptions,
    TimeGrainOption,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName: string;

  $: dashboardStore = useDashboardStore(metricViewName);

  let baseTimeRange: TimeRange;

  let metricsViewQuery;
  $: if ($runtimeStore.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtimeStore.instanceId,
      metricViewName
    );
  }

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtimeStore.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  let allTimeRangeQuery: UseQueryStoreResult;
  $: if (
    hasTimeSeries &&
    !!$runtimeStore?.instanceId &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.model &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    allTimeRangeQuery = useModelAllTimeRange(
      $runtimeStore.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      $metricsViewQuery.data.entry.metricsView.timeDimension
    );
  }
  $: allTimeRange = $allTimeRangeQuery?.data as TimeRange | undefined;

  // once we have the allTimeRange, set the default time range and time grain
  $: if (!$dashboardStore?.selectedTimeRange && allTimeRange)
    setDefaultTimeControls(allTimeRange);

  function setDefaultTimeControls(allTimeRange: TimeRange) {
    baseTimeRange = getDefaultTimeRange(allTimeRange);
    const timeGrain = getDefaultTimeGrain(
      baseTimeRange.start,
      baseTimeRange.end
    );
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrainOption[];
  $: timeGrainOptions = getTimeGrainOptions(
    new Date($dashboardStore?.selectedTimeRange?.start),
    new Date($dashboardStore?.selectedTimeRange?.end)
  );

  function onSelectTimeRange(name: TimeRangeName, start: string, end: string) {
    baseTimeRange = {
      name,
      start: new Date(start),
      end: new Date(end),
    };
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      $dashboardStore.selectedTimeRange.interval
    );
  }

  function onSelectTimeGrain(timeGrain: TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: TimeGrain
  ) {
    const { name, start, end } = timeRange;

    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    timeGrainOptions = getTimeGrainOptions(start, end);
    const isValidTimeGrain = checkValidTimeGrain(timeGrain, timeGrainOptions);
    if (!isValidTimeGrain) {
      timeGrain = getDefaultTimeGrain(start, end);
    }

    // Round start time to nearest lower time grain
    const adjustedStart = floorDate(start, timeGrain);

    // Round end time to start of next grain
    // because the runtime uses exlusive end times, whereas user inputs are inclusive
    let adjustedEnd: Date;
    if (timeRange.name === TimeRangeName.Custom) {
      // Custom Range always snaps to the end of the day
      adjustedEnd = addGrains(new Date(end), 1, TimeGrain.OneDay);
      adjustedEnd = floorDate(adjustedEnd, timeGrain);
    } else {
      adjustedEnd = addGrains(new Date(end), 1, timeGrain);
      adjustedEnd = floorDate(adjustedEnd, timeGrain);
    }

    // the adjusted time range
    const newTimeRange: TimeSeriesTimeRange = {
      name: name,
      start: adjustedStart.toISOString(),
      end: adjustedEnd.toISOString(),
      interval: timeGrain,
    };

    metricsExplorerStore.setSelectedTimeRange(metricViewName, newTimeRange);
  }
</script>

<div class="flex flex-row">
  {#if !hasTimeSeries}
    <NoTimeDimensionCTA
      {metricViewName}
      modelName={$metricsViewQuery?.data?.entry?.metricsView?.model}
    />
  {:else}
    <TimeRangeSelector
      {metricViewName}
      {allTimeRange}
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeGrainSelector
      on:select-time-grain={(e) => onSelectTimeGrain(e.detail.timeGrain)}
      {metricViewName}
      {timeGrainOptions}
    />
  {/if}
</div>
