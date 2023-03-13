<!--
@component
Constructs a TimeSeriesTimeRange object – to be used as the filter in MetricsExplorer – by taking as input:
- a base time range
- a time grain (e.g., "hour" or "day")
- the dataset's full time range (so its end time can be used in relative time ranges)

We should rename TimeSeriesTimeRange to a better name.
-->
<script lang="ts">
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    TimeRange,
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
  import {
    useRuntimeServiceGetCatalogEntry,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import type { Readable } from "svelte/store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
    useDashboardStore,
  } from "../dashboard-stores";
  import ComparisonSelector from "./ComparisonSelector.svelte";
  import CurrentTimeRangeSelector from "./CurrentTimeRangeSelector.svelte";
  import NoTimeDimensionCTA from "./NoTimeDimensionCTA.svelte";
  import {
    addGrains,
    checkValidTimeGrain,
    floorDate,
    getComparisonOptionsForTimeRange,
    getComparisonTimeRange,
    getDefaultTimeGrain,
    getDefaultTimeRange,
    getTimeGrainOptions,
    ISODurationToTimeRange,
    makeRelativeTimeRange,
    supportedTimeGrainEnums,
    TimeGrainOption,
  } from "./time-range-utils";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();
  let dashboardStore: Readable<MetricsExplorerEntity>;
  $: dashboardStore = useDashboardStore(metricViewName);

  $: selectedTimeRange = $dashboardStore?.selectedTimeRange;

  let baseTimeRange: TimeRange;
  let defaultTimeRange: TimeRangeName;
  let minTimeGrain: V1TimeGrain;

  let metricsViewQuery;
  $: if ($runtime.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  let allTimeRangeQuery: UseQueryStoreResult;
  $: if (
    hasTimeSeries &&
    !!$runtime?.instanceId &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.model &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    allTimeRangeQuery = useModelAllTimeRange(
      $runtime.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      $metricsViewQuery.data.entry.metricsView.timeDimension
    );

    defaultTimeRange = ISODurationToTimeRange(
      $metricsViewQuery.data.entry.metricsView?.defaultTimeRange
    );
    minTimeGrain =
      $metricsViewQuery.data.entry.metricsView?.smallestTimeGrain ||
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  }
  $: allTimeRange = $allTimeRangeQuery?.data as TimeRange | undefined;

  // once we have the allTimeRange, set the default time range and time grain
  $: if (allTimeRange) {
    if (!$dashboardStore?.selectedTimeRange) {
      setDefaultTimeControls(allTimeRange);
    } else if (!$dashboardStore?.selectedTimeRange.start) {
      setTimeControlsFromUrl(
        $dashboardStore.selectedTimeRange.name,
        allTimeRange
      );
    }
  }

  function setDefaultTimeControls(allTimeRange: TimeRange) {
    baseTimeRange =
      makeRelativeTimeRange(defaultTimeRange, allTimeRange) ||
      getDefaultTimeRange(allTimeRange);

    const timeGrain = getDefaultTimeGrain(
      baseTimeRange.start,
      baseTimeRange.end
    );
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
  }

  function setTimeControlsFromUrl(name: string, allTimeRange: TimeRange) {
    baseTimeRange = makeRelativeTimeRange(
      $dashboardStore?.selectedTimeRange.name,
      allTimeRange
    );
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      $dashboardStore?.selectedTimeRange.interval
    );
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrainOption[];
  $: timeGrainOptions = getTimeGrainOptions(
    new Date(selectedTimeRange?.start),
    new Date(selectedTimeRange?.end)
  );

  let comparisonOptions = [];

  function updateComparisonAppState(newTimeRange) {
    if (!newTimeRange) return;

    comparisonOptions = getComparisonOptionsForTimeRange(
      newTimeRange,
      allTimeRange
    );
    metricsExplorerStore.setShowComparison(
      metricViewName,
      !!comparisonOptions.length
    );

    if (!comparisonOptions.length) return;

    const hasComparisonFromUrl =
      $dashboardStore.selectedComparisonTimeRange &&
      !$dashboardStore.selectedComparisonTimeRange.start;
    const selectedOption = hasComparisonFromUrl
      ? comparisonOptions.find(
          (name) => name === $dashboardStore.selectedComparisonTimeRange.name
        ) ?? comparisonOptions[0]
      : comparisonOptions[0];
    const comparisonTimeRange = getComparisonTimeRange(
      newTimeRange,
      selectedOption
    );
    metricsExplorerStore.setSelectedComparisonRange(metricViewName, {
      ...comparisonTimeRange,
      name: selectedOption,
    });
  }

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

  function onSelectTimeGrain(timeGrain: V1TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain
  ) {
    const { name, start, end } = timeRange;

    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    timeGrainOptions = getTimeGrainOptions(start, end);
    const isValidTimeGrain = checkValidTimeGrain(
      timeGrain,
      timeGrainOptions,
      minTimeGrain
    );
    if (!isValidTimeGrain) {
      const defaultTimeGrain = getDefaultTimeGrain(start, end);
      const timeGrainEnums = supportedTimeGrainEnums();

      const defaultGrainIndex = timeGrainEnums.indexOf(defaultTimeGrain);
      timeGrain = defaultTimeGrain;
      let i = defaultGrainIndex;
      // loop through time grains until we find a valid one
      while (!checkValidTimeGrain(timeGrain, timeGrainOptions, minTimeGrain)) {
        timeGrain = timeGrainEnums[i + 1] as V1TimeGrain;
        i = i == timeGrainEnums.length - 1 ? -1 : i + 1;
        if (i == defaultGrainIndex) {
          // if we've looped through all the time grains and haven't found
          // a valid one, use default
          timeGrain = defaultTimeGrain;
          break;
        }
      }
    }

    // Round start time to nearest lower time grain
    const adjustedStart = floorDate(start, timeGrain);

    // Round end time to start of next grain
    // because the runtime uses exlusive end times, whereas user inputs are inclusive
    let adjustedEnd: Date;
    if (timeRange.name === TimeRangeName.Custom) {
      // Custom Range always snaps to the end of the day
      adjustedEnd = addGrains(new Date(end), 1, V1TimeGrain.TIME_GRAIN_DAY);
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

    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.setSelectedTimeRange(metricViewName, newTimeRange);

    updateComparisonAppState(newTimeRange);
  }
</script>

<div class="flex flex-row gap-x-1">
  {#if !hasTimeSeries}
    <NoTimeDimensionCTA
      {metricViewName}
      modelName={$metricsViewQuery?.data?.entry?.metricsView?.model}
    />
  {:else}
    <CurrentTimeRangeSelector
      {metricViewName}
      {allTimeRange}
      {minTimeGrain}
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />

    {#if comparisonOptions.length}
      <ComparisonSelector {comparisonOptions} {metricViewName} />
    {/if}
    <TimeGrainSelector
      on:select-time-grain={(e) => onSelectTimeGrain(e.detail.timeGrain)}
      {metricViewName}
      {timeGrainOptions}
      {minTimeGrain}
    />
  {/if}
</div>
