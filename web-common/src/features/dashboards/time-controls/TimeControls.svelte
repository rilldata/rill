<script lang="ts">
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
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
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { initLocalUserPreferenceStore } from "../user-preferences";
  import NoTimeDimensionCTA from "./NoTimeDimensionCTA.svelte";
  import TimeComparisonSelector from "./TimeComparisonSelector.svelte";
  import TimeGrainSelector from "./TimeGrainSelector.svelte";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";
  import TimeZoneSelector from "./TimeZoneSelector.svelte";
  import ComparisonSelector from "./ComparisonSelector.svelte";

  export let metricViewName: string;

  const localUserPreferences = initLocalUserPreferenceStore(metricViewName);

  const queryClient = useQueryClient();
  $: dashboardStore = useDashboardStore(metricViewName);

  let baseTimeRange: TimeRange;
  let minTimeGrain: V1TimeGrain;
  let availableTimeZones: string[] = [];
  let dimensions = [];

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  const timeControlsStore = useTimeControlStore(getStateManagers());
  $: allTimeRange = $timeControlsStore.allTimeRange;
  $: minTimeGrain = $timeControlsStore.minTimeGrain;

  $: if (
    $timeControlsStore.ready &&
    !!$metaQuery?.data?.model &&
    !!$metaQuery?.data?.timeDimension
  ) {
    availableTimeZones = $metaQuery?.data?.availableTimeZones;

    /**
     * Remove the timezone selector if no timezone key is present
     * or the available timezone list is empty. Set the default
     * timezone to UTC in such cases.
     *
     */
    if (
      !availableTimeZones?.length &&
      $dashboardStore?.selectedTimezone !== "Etc/UTC"
    ) {
      metricsExplorerStore.setTimeZone(metricViewName, "Etc/UTC");
      localUserPreferences.set({ timeZone: "Etc/UTC" });
    }

    dimensions = $metaQuery?.data?.dimensions;
    baseTimeRange ??= {
      ...$dashboardStore.selectedTimeRange,
    };
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrain[];
  $: timeGrainOptions = getAllowedTimeGrains(
    new Date($timeControlsStore.timeStart),
    new Date($timeControlsStore.timeEnd)
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
      undefined
    );
  }

  function onSelectTimeGrain(timeGrain: V1TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      timeGrain,
      $dashboardStore?.selectedComparisonTimeRange
    );
  }

  function onSelectTimeZone(timeZone: string) {
    metricsExplorerStore.setTimeZone(metricViewName, timeZone);
    localUserPreferences.set({ timeZone });
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
  }

  function enableComparison(type: string, name: string) {
    if (type === "time") {
      metricsExplorerStore.displayTimeComparison(metricViewName, true);
    } else {
      metricsExplorerStore.setComparisonDimension(metricViewName, name);
    }
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined
  ) {
    cancelDashboardQueries(queryClient, metricViewName);

    metricsExplorerStore.selectTimeRange(
      metricViewName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      $timeControlsStore.allTimeRange
    );
  }

  let availableComparisons;

  $: if (allTimeRange?.start && $timeControlsStore.ready) {
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
      on:remove-scrub={() => {
        metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
      }}
    />
    {#if availableTimeZones?.length}
      <TimeZoneSelector
        on:select-time-zone={(e) => onSelectTimeZone(e.detail.timeZone)}
        {metricViewName}
        {availableTimeZones}
        now={allTimeRange?.end}
      />
    {/if}
    <ComparisonSelector
      on:enable-comparison={(e) => {
        enableComparison(e.detail.type, e.detail.name);
      }}
      on:disable-all-comparison={() =>
        metricsExplorerStore.disableAllComparisons(metricViewName)}
      showTimeComparison={$dashboardStore?.showTimeComparison}
      selectedDimension={$dashboardStore?.selectedComparisonDimension}
      {dimensions}
    />
    {#if $dashboardStore?.showTimeComparison}
      <TimeComparisonSelector
        on:select-comparison={(e) => {
          onSelectComparisonRange(e.detail.name, e.detail.start, e.detail.end);
        }}
        {minTimeGrain}
        currentStart={$timeControlsStore.selectedTimeRange.start}
        currentEnd={$timeControlsStore.selectedTimeRange.end}
        boundaryStart={allTimeRange.start}
        boundaryEnd={allTimeRange.end}
        showComparison={$timeControlsStore?.showComparison}
        selectedComparison={$timeControlsStore?.selectedComparisonTimeRange}
        zone={$dashboardStore?.selectedTimezone}
        comparisonOptions={availableComparisons}
      />
    {/if}
    <TimeGrainSelector
      on:select-time-grain={(e) => onSelectTimeGrain(e.detail.timeGrain)}
      {metricViewName}
      {timeGrainOptions}
      {minTimeGrain}
    />
  {/if}
</div>
