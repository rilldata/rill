<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";
  import {
    TimeComparisonOption,
    TimeRangePreset,
    type DashboardTimeControls,
  } from "@rilldata/web-common/lib/time/types";
  import type {
    V1ExploreTimeRange,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { isMetricsViewQuery } from "@rilldata/web-common/runtime-client/invalidation.ts";
  import { DateTime, Interval } from "luxon";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import { applyDimensionInListMode as applyDimensionInListModeDirectly } from "../state-managers/actions/dimension-filters";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "../stores/dashboard-stores";
  import ComparisonPill from "../time-controls/comparison-pill/ComparisonPill.svelte";
  import {
    CUSTOM_TIME_RANGE_ALIAS,
    deriveInterval,
  } from "../time-controls/new-time-controls";
  import SuperPill from "../time-controls/super-pill/SuperPill.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import FilterButton from "./FilterButton.svelte";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import { featureFlags } from "../../feature-flags";
  import Timestamp from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/Timestamp.svelte";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import { Tooltip } from "bits-ui";
  import Metadata from "../time-controls/super-pill/components/Metadata.svelte";
  import { getValidComparisonOption } from "../time-controls/time-range-store";
  import { getPinnedTimeZones } from "../url-state/getDefaultExplorePreset";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const { rillTime } = featureFlags;

  export let readOnly = false;
  export let timeRanges: V1ExploreTimeRange[];
  export let metricsViewName: string;
  export let hasTimeSeries: boolean;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    exploreName,
    validSpecStore,
    actions: {
      dimensionsFilter: {
        toggleMultipleDimensionValueSelections,
        applyDimensionInListMode,
        applyDimensionContainsMode,
        removeDimensionFilter,
        toggleDimensionFilterMode,
      },
      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters, setTemporaryFilterName },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: {
        dimensionHasFilter,
        getDimensionFilterItems,
        getAllDimensionFilterItems,
        isFilterExcludeMode,
      },
      measures: { allMeasures, filteredSimpleMeasures },
      measureFilters: {
        getMeasureFilterItems,
        getAllMeasureFilterItems,
        measureHasFilter,
      },
      pivot: { showPivot },
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    dashboardStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  const dashboardStateSync = DashboardStateSync.getFromContext();

  let showDefaultItem = false;

  $: ({ instanceId } = $runtime);

  $: timeRangeQuery = useMetricsViewTimeRange(instanceId, metricsViewName);

  $: timeRangeSummary = $timeRangeQuery.data?.timeRangeSummary;

  $: watermark = timeRangeSummary?.watermark;

  $: ({
    selectedTimeRange,
    allTimeRange,
    showTimeComparison,
    selectedComparisonTimeRange,
    minTimeGrain,
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $timeControlsStore);

  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

  $: exploreState = useExploreState($exploreName);
  $: activeTimeZone = $exploreState?.selectedTimezone;

  $: selectedRangeAlias =
    selectedTimeRange?.name === TimeRangePreset.CUSTOM
      ? `${selectedTimeRange.start.toISOString()},${selectedTimeRange.end.toISOString()}`
      : selectedTimeRange?.name;

  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = exploreSpec.defaultPreset?.timeRange;

  $: dimensions = $allDimensions;
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measures = $allMeasures;
  $: measureIdMap = getMapFromArray(measures, (m) => m.name as string);

  $: currentDimensionFilters = $getDimensionFilterItems(dimensionIdMap);
  $: allDimensionFilters = $getAllDimensionFilterItems(
    currentDimensionFilters,
    dimensionIdMap,
  );

  $: currentMeasureFilters = $getMeasureFilterItems(measureIdMap);
  $: allMeasureFilters = $getAllMeasureFilterItems(
    currentMeasureFilters,
    measureIdMap,
  );

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    currentDimensionFilters.length > 0 || currentMeasureFilters.length > 0;

  $: isComplexFilter = isExpressionUnsupported($dashboardStore.whereFilter);

  $: availableTimeZones = getPinnedTimeZones(exploreSpec);

  $: allTimeRangeInterval = allTimeRange
    ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
    : Interval.invalid("Invalid interval");

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : allTimeRange
      ? Interval.fromDateTimes(allTimeRange.start, allTimeRange.end)
      : Interval.invalid("Invalid interval");

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    filter: MeasureFilterEntry,
  ) {
    if (oldDimension && oldDimension !== dimension) {
      removeMeasureFilter(oldDimension, measureName);
    }
    setMeasureFilter(dimension, filter);
  }

  function onPan(direction: "left" | "right") {
    const panRange = $getNewPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;

    const timeRange = {
      name: CUSTOM_TIME_RANGE_ALIAS,
      start: start,
      end: end,
    };

    const comparisonTimeRange = {
      name: TimeComparisonOption.CONTIGUOUS,
    } as DashboardTimeControls; // FIXME wrong typecasting across application

    if (!activeTimeGrain) return;
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange as TimeRange,
      activeTimeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }

  async function onSelectRange(alias: string) {
    // If we don't have a valid time range, early return
    if (!allTimeRange?.end) return;

    // This should be returned by the API, but it is not yet implemented
    const includesTimeZoneOffset = alias.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = alias.match(/tz (.*)/)?.[1];

      if (timeZone) metricsExplorerStore.setTimeZone($exploreName, timeZone);
    }

    await queryClient.cancelQueries({
      predicate: (query) =>
        isMetricsViewQuery(query.queryHash, metricsViewName),
    });

    const { interval, grain } = await deriveInterval(
      alias,

      metricsViewName,
      activeTimeZone,
    );

    if (interval.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        name: alias,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange, grain);
    }
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    metricsExplorerStore.selectTimeRange(
      $exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      metricsViewSpec,
    );
  }

  function selectRange(range: TimeRange, grain?: V1TimeGrain) {
    const timeGrain =
      grain ?? getDefaultTimeGrain(range.start, range.end).grain;

    // Get valid option for the new time range
    const validComparison =
      allTimeRange &&
      getValidComparisonOption(
        exploreSpec.timeRanges,
        range,
        $exploreState.selectedComparisonTimeRange?.name as
          | TimeComparisonOption
          | undefined,
        allTimeRange,
      );

    makeTimeSeriesTimeRangeAndUpdateAppState(range, timeGrain, {
      name: validComparison,
    } as DashboardTimeControls);
  }

  function onSelectTimeZone(timeZone: string) {
    if (!interval.isValid) return;

    if (selectedRangeAlias === TimeRangePreset.CUSTOM) {
      selectRange({
        name: TimeRangePreset.CUSTOM,
        start: interval.start
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
        end: interval.end
          ?.setZone(timeZone, { keepLocalTime: true })
          .toJSDate(),
      });
    }

    metricsExplorerStore.setTimeZone($exploreName, timeZone);
  }

  $: usingRillTime =
    !selectedRangeAlias?.startsWith("P") &&
    !selectedRangeAlias?.startsWith("rill-");

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (usingRillTime && selectedRangeAlias) {
      metricsExplorerStore.setTimeGrain($exploreName, timeGrain);
    } else if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        $dashboardStore?.selectedComparisonTimeRange,
      );
    }
  }

  function isUrlTooLongAfterInListFilter(
    dimensionName: string,
    values: string[],
  ) {
    if (!dashboardStateSync) return false;

    const exploreState = structuredClone($dashboardStore);
    applyDimensionInListModeDirectly(
      { dashboard: exploreState },
      dimensionName,
      values,
    );
    const url = dashboardStateSync.getUrlForExploreState(exploreState);
    return isUrlTooLong(url);
  }
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
      <Tooltip.Root openDelay={0}>
        <Tooltip.Trigger class="cursor-default">
          <Calendar size="16px" />
        </Tooltip.Trigger>
        <Tooltip.Content side="bottom" sideOffset={10}>
          <Metadata
            timeZone={activeTimeZone}
            timeStart={allTimeRange?.start}
            timeEnd={allTimeRange?.end}
          />
        </Tooltip.Content>
      </Tooltip.Root>
      {#if allTimeRange?.start && allTimeRange?.end}
        <SuperPill
          {allTimeRange}
          {selectedRangeAlias}
          showPivot={$showPivot}
          {minTimeGrain}
          {defaultTimeRange}
          {availableTimeZones}
          {timeRanges}
          complete={false}
          {interval}
          context={$exploreName}
          {timeStart}
          {timeEnd}
          lockTimeZone={exploreSpec.lockTimeZone}
          allowCustomTimeRange={exploreSpec.allowCustomTimeRange}
          {activeTimeGrain}
          {activeTimeZone}
          canPanLeft={$canPanLeft}
          canPanRight={$canPanRight}
          {showDefaultItem}
          watermark={watermark ? DateTime.fromISO(watermark) : undefined}
          applyRange={selectRange}
          {onSelectRange}
          {onTimeGrainSelect}
          {onSelectTimeZone}
          {onPan}
        />
        <ComparisonPill
          {minTimeGrain}
          {allTimeRange}
          {selectedTimeRange}
          showTimeComparison={!!showTimeComparison}
          {selectedComparisonTimeRange}
        />
      {/if}

      {#if !$rillTime && allTimeRangeInterval?.end?.isValid}
        <Tooltip.Root openDelay={0}>
          <Tooltip.Trigger>
            <span class="text-gray-600 italic">
              as of <Timestamp
                id="filter-bar-as-of"
                italic
                suppress
                showDate={false}
                date={allTimeRangeInterval.end}
                zone={activeTimeZone}
              />
            </span>
          </Tooltip.Trigger>
          <Tooltip.Content side="bottom" sideOffset={10}>
            <Metadata
              timeZone={activeTimeZone}
              timeStart={allTimeRange?.start}
              timeEnd={allTimeRange?.end}
            />
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}
    </div>
  {/if}

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={$dashboardStore.whereFilter} />
      {:else if !allDimensionFilters.length && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {:else}
        {#each allDimensionFilters as { name, label, mode, selectedValues, inputText } (name)}
          {@const dimension = dimensions.find(
            (d) => d.name === name || d.column === name,
          )}
          {@const dimensionName = dimension?.name || dimension?.column}
          <div animate:flip={{ duration: 200 }}>
            {#if dimensionName}
              <DimensionFilter
                whereFilter={$dashboardStore.whereFilter}
                metricsViewNames={[metricsViewName]}
                {readOnly}
                {name}
                {label}
                {mode}
                {selectedValues}
                {inputText}
                {timeStart}
                {timeEnd}
                {timeControlsReady}
                excludeMode={$isFilterExcludeMode(name)}
                onRemove={() => removeDimensionFilter(name)}
                onToggleFilterMode={() => toggleDimensionFilterMode(name)}
                onSelect={(value) =>
                  toggleMultipleDimensionValueSelections(name, [value], true)}
                onMultiSelect={(values) =>
                  toggleMultipleDimensionValueSelections(name, values, true)}
                onApplyInList={(values) =>
                  applyDimensionInListMode(name, values)}
                onApplyContainsMode={(searchText) =>
                  applyDimensionContainsMode(name, searchText)}
                isUrlTooLongAfterInListFilter={(values) =>
                  isUrlTooLongAfterInListFilter(name, values)}
              />
            {/if}
          </div>
        {/each}
        {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensions}
              {name}
              {label}
              {dimensionName}
              {filter}
              onRemove={() => removeMeasureFilter(dimensionName, name)}
              onApply={({ dimension, oldDimension, filter }) =>
                handleMeasureFilterApply(dimension, name, oldDimension, filter)}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton
          allDimensions={dimensions}
          filteredSimpleMeasures={$filteredSimpleMeasures()}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          {setTemporaryFilterName}
        />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasFilters}
          <Button type="text" onClick={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
