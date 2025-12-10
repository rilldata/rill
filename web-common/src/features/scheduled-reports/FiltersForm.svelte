<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
  import { deriveInterval } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import type { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config.ts";
  import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
  import {
    type DashboardTimeControls,
    TimeComparisonOption,
    type TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types.ts";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { isMetricsViewQuery } from "@rilldata/web-common/runtime-client/invalidation.ts";
  import { DateTime, Interval } from "luxon";
  import { onMount } from "svelte";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  export let filters: Filters;
  export let timeControls: TimeControls;
  export let readOnly = false;
  export let maxWidth: number | undefined = undefined;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  $: ({
    whereFilter,
    allDimensionFilterItems,
    isFilterExcludeMode,
    dimensionHasFilter,
    allMeasureFilterItems,
    measureHasFilter,
    hasFilters,

    removeDimensionFilter,
    toggleDimensionFilterMode,
    toggleMultipleDimensionValueSelections,
    applyDimensionInListMode,
    applyDimensionContainsMode,
    removeMeasureFilter,
    setMeasureFilter,
    setTemporaryFilterName,
    clearAllFilters,
    metricsViewMetadata: {
      metricsViewName,
      allDimensions,
      allSimpleMeasures,
      validSpecQuery,
    },
  } = filters);

  $: ({
    selectedTimezone,
    allTimeRange,
    timeRangeStateStore,
    comparisonRangeStateStore,
    minTimeGrain: _minTimeGrain,
    setTimeZone,
    selectTimeRange,
    setSelectedComparisonRange,
    displayTimeComparison,
  } = timeControls);

  $: exploreSpec = $validSpecQuery.data?.explore ?? {};

  $: isComplexFilter = isExpressionUnsupported($whereFilter);

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});
  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = exploreSpec.defaultPreset?.timeRange;
  $: availableTimeZones = exploreSpec.timeZones ?? [];
  $: timeRanges = exploreSpec.timeRanges ?? [];

  $: minTimeGrain = $_minTimeGrain;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone($selectedTimezone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone($selectedTimezone),
      )
    : Interval.fromDateTimes(
        $allTimeRange?.start ?? new Date(),
        $allTimeRange?.end ?? new Date(),
      );

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

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    selectTimeRange(timeRange, timeGrain, comparisonTimeRange);
  }

  function selectRange(
    range: TimeRange,
    grain?: V1TimeGrain,
    rangeOnly: boolean = false,
  ) {
    const defaultTimeGrain =
      grain ?? getDefaultTimeGrain(range.start, range.end).grain;

    const comparisonOption = DEFAULT_TIME_RANGES[range.name as TimeRangePreset]
      ?.defaultComparison as TimeComparisonOption;

    // Get valid option for the new time range
    const validComparison = $allTimeRange && comparisonOption;

    makeTimeSeriesTimeRangeAndUpdateAppState(
      range,
      defaultTimeGrain,
      rangeOnly
        ? undefined
        : ({
            name: validComparison,
          } as DashboardTimeControls),
    );
  }

  async function onSelectRange(name: string, rangeOnly: boolean = false) {
    if (!$allTimeRange?.end) {
      return;
    }

    const includesTimeZoneOffset = name.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = name.match(/tz (.*)/)?.[1];

      if (timeZone) setTimeZone(timeZone);
    }

    await queryClient.cancelQueries({
      predicate: (query) =>
        isMetricsViewQuery(query.queryHash, metricsViewName),
    });

    const { interval, grain } = await deriveInterval(
      name,

      metricsViewName,
      $selectedTimezone,
    );

    if (interval?.isValid) {
      const validInterval = interval as Interval<true>;
      const baseTimeRange: TimeRange = {
        name: name as TimeRangePreset,
        start: validInterval.start.toJSDate(),
        end: validInterval.end.toJSDate(),
      };

      selectRange(baseTimeRange, grain, rangeOnly);
    }
  }

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        selectedComparisonTimeRange,
      );
    }
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

    setTimeZone(timeZone);
  }

  onMount(() => {
    if (selectedRangeAlias) onSelectRange(selectedRangeAlias, true);
  });
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
  aria-label="Filters form"
>
  <div
    class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center ml-2 pointer-events-auto w-fit"
  >
    <Calendar size="16px" />
    {#if $allTimeRange}
      <SuperPill
        allTimeRange={$allTimeRange}
        {selectedRangeAlias}
        showPivot={false}
        {defaultTimeRange}
        {availableTimeZones}
        {timeRanges}
        complete={false}
        {interval}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        activeTimeZone={$selectedTimezone}
        allowCustomTimeRange={false}
        showDefaultItem
        applyRange={selectRange}
        {onSelectRange}
        {onTimeGrainSelect}
        {onSelectTimeZone}
        hidePan
        onPan={() => {}}
        {minTimeGrain}
        {side}
      />
      <CanvasComparisonPill
        {minTimeGrain}
        allTimeRange={$allTimeRange}
        {selectedTimeRange}
        {selectedComparisonTimeRange}
        showTimeComparison={$comparisonRangeStateStore?.showTimeComparison ??
          false}
        activeTimeZone={$selectedTimezone}
        onDisplayTimeComparison={displayTimeComparison}
        onSetSelectedComparisonRange={setSelectedComparisonRange}
        allowCustomTimeRange={false}
        {side}
      />
    {/if}
  </div>

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start ml-2">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div
      class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 pointer-events-auto"
    >
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={$whereFilter} />
      {:else if !$allDimensionFilterItems.length && !$allMeasureFilterItems.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {:else}
        {#each $allDimensionFilterItems as { name, label, mode, selectedValues, inputText, metricsViewNames } (name)}
          {@const dimension = $allDimensions.find(
            (d) => d.name === name || d.column === name,
          )}
          {@const dimensionName = dimension?.name || dimension?.column}
          <div animate:flip={{ duration: 200 }}>
            {#if dimensionName && metricsViewNames?.length}
              <DimensionFilter
                {metricsViewNames}
                {name}
                {label}
                {mode}
                {selectedValues}
                {inputText}
                {timeStart}
                {timeEnd}
                {side}
                timeControlsReady
                excludeMode={$isFilterExcludeMode(name)}
                whereFilter={$whereFilter}
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
              />
            {/if}
          </div>
        {/each}
        {#each $allMeasureFilterItems as { name, label, dimensionName, filter, dimensions: dimensionsForMeasure } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensionsForMeasure || $allDimensions}
              {name}
              {label}
              {dimensionName}
              {filter}
              {side}
              onRemove={() => removeMeasureFilter(dimensionName, name)}
              onApply={({ dimension, oldDimension, filter }) =>
                handleMeasureFilterApply(dimension, name, oldDimension, filter)}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton
          allDimensions={$allDimensions}
          filteredSimpleMeasures={$allSimpleMeasures}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          {setTemporaryFilterName}
          {side}
        />
        <!-- if filters are present, place a chip at the end of the flex container
      that enables clearing all filters -->
        {#if $hasFilters}
          <Button type="text" onClick={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
