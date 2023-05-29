<script lang="ts">
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    getAvailableComparisonsForTimeRange,
    getComparisonRange,
  } from "@rilldata/web-common/lib/time/comparisons";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import {
    checkValidTimeGrain,
    getDefaultTimeGrain,
    findValidTimeGrain,
    getAllowedTimeGrains,
  } from "@rilldata/web-common/lib/time/grains";
  import {
    ISODurationToTimePreset,
    convertTimeRangePreset,
    isRangeInsideOther,
  } from "@rilldata/web-common/lib/time/ranges";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeGrain,
    TimeRange,
    TimeRangePreset,
    TimeRangeType,
  } from "@rilldata/web-common/lib/time/types";
  import {
    V1TimeGrain,
    createRuntimeServiceGetCatalogEntry,
  } from "@rilldata/web-common/runtime-client";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
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
  let defaultTimeRange: TimeRangeType;
  let minTimeGrain: V1TimeGrain;

  let metricsViewQuery;
  $: if ($runtime.instanceId) {
    metricsViewQuery = createRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

  let allTimeRangeQuery: CreateQueryResult;
  $: if (
    hasTimeSeries &&
    !!$runtime?.instanceId &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.model &&
    !!$metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    allTimeRangeQuery = useModelAllTimeRange(
      $runtime.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      $metricsViewQuery.data.entry.metricsView.timeDimension,
      {
        query: {
          enabled: !!hasTimeSeries,
        },
      }
    );
    defaultTimeRange = ISODurationToTimePreset(
      $metricsViewQuery.data.entry.metricsView?.defaultTimeRange
    );
    minTimeGrain =
      $metricsViewQuery.data.entry.metricsView?.smallestTimeGrain ||
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  }
  $: allTimeRange = $allTimeRangeQuery?.data as TimeRange;
  $: isDashboardDefined = $dashboardStore !== undefined;
  // Once we have the allTimeRange, set the default time range and time grain.
  // This is a temporary workaround with high potential to break. We should refactor this defaulting logic to live with the store, not as part of a component.
  $: if (allTimeRange && allTimeRange?.start && isDashboardDefined) {
    const selectedTimeRange = $dashboardStore?.selectedTimeRange;

    if (!selectedTimeRange) {
      setDefaultTimeControls(allTimeRange);
    } else {
      setTimeControlsFromUrl(allTimeRange);
    }
  }

  function setDefaultTimeControls(allTimeRange: DashboardTimeControls) {
    baseTimeRange =
      convertTimeRangePreset(
        defaultTimeRange,
        allTimeRange.start,
        allTimeRange.end
      ) || allTimeRange;

    const timeGrain = getDefaultTimeGrain(
      baseTimeRange.start,
      baseTimeRange.end
    );
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      timeGrain.grain,
      {}
    );

    /** enable comparisons by default */
    metricsExplorerStore.toggleComparison(metricViewName, true);
    metricsExplorerStore.allDefaultsSelected(metricViewName);
  }

  function setTimeControlsFromUrl(allTimeRange: TimeRange) {
    metricsExplorerStore.allDefaultsSelected(metricViewName);

    if ($dashboardStore?.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
      /** set the time range to the fixed custom time range */
      baseTimeRange = {
        name: TimeRangePreset.CUSTOM,
        start: new Date($dashboardStore?.selectedTimeRange.start),
        end: new Date($dashboardStore?.selectedTimeRange.end),
      };
    } else {
      /** rebuild off of relative time range */
      baseTimeRange =
        convertTimeRangePreset(
          $dashboardStore?.selectedTimeRange.name,
          allTimeRange.start,
          allTimeRange.end
        ) || allTimeRange;
    }

    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      $dashboardStore?.selectedTimeRange.interval,
      // do not reset the comparison state when pulling from the URL
      $dashboardStore?.selectedComparisonTimeRange
    );
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrain[];
  $: timeGrainOptions = getAllowedTimeGrains(
    new Date($dashboardStore?.selectedTimeRange?.start),
    new Date($dashboardStore?.selectedTimeRange?.end)
  );

  function onSelectTimeRange(name: TimeRangeType, start: Date, end: Date) {
    baseTimeRange = {
      name,
      start: new Date(start),
      end: new Date(end),
    };
    makeTimeSeriesTimeRangeAndUpdateAppState(
      baseTimeRange,
      $dashboardStore.selectedTimeRange?.interval,
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
    metricsExplorerStore.toggleComparison(metricViewName, true);
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
    const { name, start, end } = timeRange;

    // validate time range name + time grain combination
    // (necessary because when the time range name is changed, the current time grain may not be valid for the new time range name)
    timeGrainOptions = getAllowedTimeGrains(start, end);
    const isValidTimeGrain = checkValidTimeGrain(
      timeGrain,
      timeGrainOptions,
      minTimeGrain
    );

    if (!isValidTimeGrain) {
      const defaultTimeGrain = getDefaultTimeGrain(start, end).grain;
      timeGrain = findValidTimeGrain(
        defaultTimeGrain,
        timeGrainOptions,
        minTimeGrain
      );
    }

    // the adjusted time range
    const newTimeRange: DashboardTimeControls = {
      name,
      start,
      end,
      interval: timeGrain,
    };

    cancelDashboardQueries(queryClient, metricViewName);

    metricsExplorerStore.setSelectedTimeRange(metricViewName, newTimeRange);

    // reset comparisonOption to the default for the new time range.

    // if no name in comprisonTimeRange, set selectedComparisonTimeRange to default.
    if (comparisonTimeRange !== undefined) {
      let selectedComparisonTimeRange;
      if (!comparisonTimeRange?.name) {
        const comparisonOption = DEFAULT_TIME_RANGES[name]
          ?.defaultComparison as TimeComparisonOption;
        const range = getComparisonRange(start, end, comparisonOption);

        selectedComparisonTimeRange = {
          ...range,
          name: comparisonOption,
        };
      } else if (comparisonTimeRange.name === TimeComparisonOption.CUSTOM) {
        selectedComparisonTimeRange = comparisonTimeRange;
      } else {
        // variable time range of some kind.
        const comparisonOption =
          comparisonTimeRange.name as TimeComparisonOption;
        const range = getComparisonRange(start, end, comparisonOption);

        selectedComparisonTimeRange = {
          ...range,
          name: comparisonOption,
        };
      }

      metricsExplorerStore.setSelectedComparisonRange(
        metricViewName,
        selectedComparisonTimeRange
      );
    }
  }

  let isComparisonRangeAvailable;
  let availableComparisons;

  $: if (
    allTimeRange?.start &&
    $dashboardStore?.selectedTimeRange?.start &&
    hasTimeSeries
  ) {
    isComparisonRangeAvailable = isRangeInsideOther(
      allTimeRange.start,
      allTimeRange.end,
      $dashboardStore?.selectedComparisonTimeRange?.start,
      $dashboardStore?.selectedComparisonTimeRange?.end
    );

    availableComparisons = getAvailableComparisonsForTimeRange(
      allTimeRange.start,
      allTimeRange.end,
      $dashboardStore?.selectedTimeRange?.start,
      $dashboardStore?.selectedTimeRange?.end,
      [...Object.values(TimeComparisonOption)],
      [
        $dashboardStore?.selectedComparisonTimeRange
          ?.name as TimeComparisonOption,
      ]
    );
  }
</script>

<div class="flex flex-row items-center gap-x-1">
  {#if !hasTimeSeries}
    <NoTimeDimensionCTA
      {metricViewName}
      modelName={$metricsViewQuery?.data?.entry?.metricsView?.model}
    />
  {:else if allTimeRange?.start}
    <TimeRangeSelector
      {metricViewName}
      {minTimeGrain}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      selectedRange={$dashboardStore?.selectedTimeRange}
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeComparisonSelector
      on:select-comparison={(e) => {
        onSelectComparisonRange(e.detail.name, e.detail.start, e.detail.end);
      }}
      on:disable-comparison={() =>
        metricsExplorerStore.toggleComparison(metricViewName, false)}
      {minTimeGrain}
      currentStart={$dashboardStore?.selectedTimeRange?.start}
      currentEnd={$dashboardStore?.selectedTimeRange?.end}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      {isComparisonRangeAvailable}
      showComparison={$dashboardStore?.showComparison}
      selectedComparison={$dashboardStore?.selectedComparisonTimeRange}
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
