<script lang="ts">
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { createTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getAvailableComparisonsForTimeRange } from "@rilldata/web-common/lib/time/comparisons";
  import {
    getDefaultTimeGrain,
    getAllowedTimeGrains,
  } from "@rilldata/web-common/lib/time/grains";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeGrain,
    TimeRange,
    TimeRangeType,
  } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import NoTimeDimensionCTA from "./NoTimeDimensionCTA.svelte";
  import TimeComparisonSelector from "./TimeComparisonSelector.svelte";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();
  $: dashboardStore = useDashboardStore(metricViewName);

  let baseTimeRange: TimeRange;
  let minTimeGrain: V1TimeGrain;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  $: timeControlsStore = createTimeControlStore(
    $runtime.instanceId,
    metricViewName,
    $metaQuery?.data
  );
  $: allTimeRange = $timeControlsStore.allTimeRange;
  $: minTimeGrain = $timeControlsStore.minTimeGrain;

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrain[];
  $: timeGrainOptions = getAllowedTimeGrains(
    new Date($timeControlsStore.startTime),
    new Date($timeControlsStore.endTime)
  );

  function onSelectTimeRange(name: TimeRangeType, start: Date, end: Date) {
    baseTimeRange = {
      name,
      start: new Date(start),
      end: new Date(end),
    };

    const defaultTimeGrain = getDefaultTimeGrain(
      baseTimeRange.start,
      baseTimeRange.end
    ).grain;

    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      defaultTimeGrain,
      // reset the comparison range
      {}
    );
  }

  function onSelectTimeGrain(timeGrain: V1TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      timeGrain,
      $dashboardStore?.selectedComparisonTimeRange
    );
  }

  function onSelectComparisonRange(
    name: TimeComparisonOption,
    start: Date,
    end: Date
  ) {
    metricsExplorerStore.setSelectedComparisonRange(metricViewName, {
      name,
      start,
      end,
    });
    metricsExplorerStore.displayComparison(metricViewName, true);
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls
  ) {
    cancelDashboardQueries(queryClient, metricViewName);

    metricsExplorerStore.setSelectedTimeRange(metricViewName, {
      ...timeRange,
      interval: timeGrain,
    });
    metricsExplorerStore.setSelectedComparisonRange(
      metricViewName,
      comparisonTimeRange
    );
  }

  let availableComparisons;

  $: if (allTimeRange?.start && $timeControlsStore.hasTime) {
    availableComparisons = getAvailableComparisonsForTimeRange(
      allTimeRange.start,
      allTimeRange.end,
      $timeControlsStore.selectedTimeRange.start,
      $timeControlsStore.selectedTimeRange.end,
      [...Object.values(TimeComparisonOption)],
      [
        $timeControlsStore.selectedComparisonTimeRange
          ?.name as TimeComparisonOption,
      ]
    );
  }
</script>

<div class="flex flex-row items-center gap-x-1">
  {#if !hasTimeSeries}
    <NoTimeDimensionCTA {metricViewName} modelName={$metaQuery?.data?.model} />
  {:else if allTimeRange?.start}
    <TimeRangeSelector
      {metricViewName}
      {minTimeGrain}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      selectedRange={$timeControlsStore?.selectedTimeRange}
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeComparisonSelector
      on:select-comparison={(e) => {
        onSelectComparisonRange(e.detail.name, e.detail.start, e.detail.end);
      }}
      on:disable-comparison={() =>
        metricsExplorerStore.displayComparison(metricViewName, false)}
      {minTimeGrain}
      currentStart={$timeControlsStore.selectedTimeRange.start}
      currentEnd={$timeControlsStore.selectedTimeRange.end}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      showComparison={$timeControlsStore?.showComparison}
      selectedComparison={$timeControlsStore?.selectedComparisonTimeRange}
      comparisonOptions={availableComparisons}
    />
    <TimeGrainSelector
      on:select-time-grain={(e) => onSelectTimeGrain(e.detail.timeGrain)}
      {metricViewName}
      {timeGrainOptions}
      {minTimeGrain}
    />
  {/if}
</div>
