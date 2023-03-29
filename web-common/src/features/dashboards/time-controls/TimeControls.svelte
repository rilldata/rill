<script lang="ts">
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time//config";
  import {
    getAvailableComparisonsForTimeRange,
    getComparisonRange,
    isComparisonInsideBounds,
  } from "@rilldata/web-common/lib/time/comparisons";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import {
    checkValidTimeGrain,
    getDefaultTimeGrain,
    getTimeGrainOptions,
  } from "@rilldata/web-common/lib/time/grains";
  import {
    convertTimeRangePreset,
    ISODurationToTimePreset,
    isRangeInsideOther,
  } from "@rilldata/web-common/lib/time/ranges";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeGrainOption,
    TimeRange,
    TimeRangePreset,
    TimeRangeType,
  } from "@rilldata/web-common/lib/time/types";
  import {
    useRuntimeServiceGetCatalogEntry,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { useQueryClient } from "@sveltestack/svelte-query";
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
    defaultTimeRange = ISODurationToTimePreset(
      $metricsViewQuery.data.entry.metricsView?.defaultTimeRange
    );
    minTimeGrain =
      $metricsViewQuery.data.entry.metricsView?.smallestTimeGrain ||
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  }
  $: allTimeRange = $allTimeRangeQuery?.data as TimeRange;
  // Once we have the allTimeRange, set the default time range and time grain.
  // This reactive statement feels a bit precarious!
  $: if (allTimeRange && allTimeRange?.start && $dashboardStore !== undefined) {
    if (!$dashboardStore?.selectedTimeRange) {
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
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain.grain);
  }

  function setTimeControlsFromUrl(allTimeRange: TimeRange) {
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
      $dashboardStore?.selectedTimeRange.interval
    );
  }

  // we get the timeGrainOptions so that we can assess whether or not the
  // activeTimeGrain is valid whenever the baseTimeRange changes
  let timeGrainOptions: TimeGrainOption[];
  // FIXME: we should be deprecating this getTimeGrainOptions in favor of getAllowedTimeGrains.
  $: timeGrainOptions = getTimeGrainOptions(
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
      $dashboardStore.selectedTimeRange?.interval
    );
  }

  function onSelectTimeGrain(timeGrain: V1TimeGrain) {
    makeTimeSeriesTimeRangeAndUpdateAppState(baseTimeRange, timeGrain);
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
      const defaultTimeGrain = getDefaultTimeGrain(start, end).grain;
      const timeGrainEnums = Object.values(TIME_GRAIN).map(
        (timeGrain) => timeGrain.grain
      );

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

    // the adjusted time range
    const newTimeRange: DashboardTimeControls = {
      name,
      start,
      end,
      interval: timeGrain,
    };

    // reset comparisonOption to the default for the new time range.
    comparisonOption = DEFAULT_TIME_RANGES[name]?.defaultComparison;
    const comparisonTimeRange = getComparisonRange(
      start,
      end,
      comparisonOption
    );

    const selectedComparisonTimeRange = {
      ...comparisonTimeRange,
      name: comparisonOption,
    };

    cancelDashboardQueries(queryClient, metricViewName);

    metricsExplorerStore.setSelectedTimeRange(metricViewName, newTimeRange);
    metricsExplorerStore.setSelectedComparisonRange(
      metricViewName,
      selectedComparisonTimeRange
    );
  }

  let comparisonOption;
  let isComparisonRangeAvailable;
  let availableComparisons;

  $: if (
    allTimeRange?.start &&
    $dashboardStore?.selectedTimeRange?.start &&
    hasTimeSeries
  ) {
    const { start, end } = $dashboardStore?.selectedTimeRange;

    if (!comparisonOption) {
      comparisonOption =
        DEFAULT_TIME_RANGES[$dashboardStore?.selectedTimeRange?.name]
          .defaultComparison;
    }
    if ($dashboardStore?.selectedComparisonTimeRange?.name) {
      comparisonOption = $dashboardStore?.selectedComparisonTimeRange?.name;
    }

    isComparisonRangeAvailable = isRangeInsideOther(
      allTimeRange.start,
      allTimeRange.end,
      $dashboardStore?.selectedComparisonTimeRange?.start,
      $dashboardStore?.selectedComparisonTimeRange?.end
    );

    console.log(
      "isComparisonRangeAvailable",
      isComparisonRangeAvailable,
      allTimeRange,
      $dashboardStore?.selectedComparisonTimeRange
    );

    availableComparisons = getAvailableComparisonsForTimeRange(
      allTimeRange.start,
      allTimeRange.end,
      start,
      end,
      [...Object.values(TimeComparisonOption)],
      [comparisonOption]
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
      on:select-time-range={(e) =>
        onSelectTimeRange(e.detail.name, e.detail.start, e.detail.end)}
    />
    <TimeComparisonSelector
      on:select-comparison={(e) => {
        onSelectComparisonRange(e.detail.name, e.detail.start, e.detail.end);
      }}
      {minTimeGrain}
      currentStart={$dashboardStore?.selectedTimeRange?.start}
      currentEnd={$dashboardStore?.selectedTimeRange?.end}
      boundaryStart={allTimeRange.start}
      boundaryEnd={allTimeRange.end}
      {metricViewName}
      showComparison={isComparisonRangeAvailable}
      {comparisonOption}
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
