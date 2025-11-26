<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import { flip } from "svelte/animate";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasFilterButton from "../../dashboards/filters/CanvasFilterButton.svelte";
  import { derived } from "svelte/store";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let readOnly = false;
  export let maxWidth: number;
  export let canvasName: string;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  let showDefaultItem = false;
  let justAdded = false;

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: {
      filterManager: {
        _allDimensions,
        _allMeasures,
        _activeUIFilters,
        metricsViewFilters,
        actions: {
          toggleMultipleDimensionValueSelections,
          toggleDimensionFilterMode,
          applyDimensionInListMode,
          addTemporaryFilter,
          applyDimensionContainsMode,
          removeDimensionFilter,
          setMeasureFilter,
          removeMeasureFilter,
        },
        clearAllFilters,
        pinFilter,
      },
      filters: {
        // setMeasureFilter,
        // removeMeasureFilter,
        allMeasureFilterItems,
        measureHasFilter,
      },
      spec,
      metricsView: { allDimensions, allSimpleMeasures },
      timeControls: {
        _canPan,
        allTimeRange,
        timeRangeStateStore,
        comparisonRangeStateStore,
        selectedTimezone,
        minTimeGrain,
        set,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: activeTimeZone = $selectedTimezone;

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $spec?.defaultPreset?.timeRange;
  $: availableTimeZones = $spec?.timeZones ?? [];
  $: timeRanges = $spec?.timeRanges ?? [];

  $: ({
    dimensions,
    hasFilters,
    measures: measureFilters,
    complexFilters,
    hasClearableFilters,
  } = $_activeUIFilters);

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.invalid("Unable to parse time range");

  $: allMeasureFilters = $allMeasureFilterItems;

  $: canPan = $_canPan;

  async function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    filter: MeasureFilterEntry,
    metricsViewNames: string[],
  ) {
    console.log("component");
    // console.log(dimensions, measureName, filter, oldDimension);
    // if (oldDimension && oldDimension !== dimension) {
    //   removeMeasureFilter(oldDimension, measureName);
    // }
    await setMeasureFilter(dimension, filter, metricsViewNames);
  }

  function onPan(direction: "left" | "right") {
    const getPanRange = getPanRangeForTimeRange(
      selectedTimeRange,
      activeTimeZone,
    );
    const panRange = getPanRange(direction);
    if (!panRange) return;
    const { start, end } = panRange;

    if (!activeTimeGrain) return;

    set.range(`${start.toISOString()},${end.toISOString()}`);
    set.comparison(TimeComparisonOption.CONTIGUOUS);
  }

  $: filterMap = derived(
    Array.from(metricsViewFilters.values()).map((p) => p.parsed),
    ($metricsViewFilters) => {
      const map = new Map<string, V1Expression>();
      $metricsViewFilters.forEach((expr, i) => {
        const mvName = Array.from(metricsViewFilters.keys())[i];
        map.set(mvName, expr.where);
      });
      return map;
    },
  );
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
>
  <div class="p-2 flex justify-between size-full py-0">
    <div class="flex items-center size-full">
      <div class="flex-none h-full pt-1.5">
        <Calendar size="16px" />
      </div>
      <div
        class="flex flex-wrap gap-x-2 gap-y-1.5 pl-2 pointer-events-auto size-full pr-2"
      >
        <SuperPill
          context={canvasName}
          allTimeRange={$allTimeRange}
          {selectedRangeAlias}
          showPivot={false}
          minTimeGrain={$minTimeGrain}
          {defaultTimeRange}
          {availableTimeZones}
          {timeRanges}
          complete={false}
          {interval}
          {timeStart}
          {timeEnd}
          {activeTimeGrain}
          {activeTimeZone}
          canPanLeft={canPan.left}
          canPanRight={canPan.right}
          watermark={undefined}
          allowCustomTimeRange={$spec?.allowCustomTimeRange}
          {showDefaultItem}
          applyRange={(timeRange) => {
            const string = `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`;
            set.range(string);
          }}
          onSelectRange={set.range}
          onTimeGrainSelect={set.grain}
          onSelectTimeZone={set.zone}
          {onPan}
        />
        <CanvasComparisonPill
          allTimeRange={$allTimeRange}
          {selectedTimeRange}
          {selectedComparisonTimeRange}
          {activeTimeZone}
          minTimeGrain={$minTimeGrain}
          showTimeComparison={$comparisonRangeStateStore?.showTimeComparison ??
            false}
          onDisplayTimeComparison={set.comparison}
          onSetSelectedComparisonRange={(range) => {
            if (range.name === "CUSTOM_COMPARISON_RANGE") {
              const stringRange = `${range.start.toISOString()},${range.end.toISOString()}`;
              set.comparison(stringRange);
            } else if (range.name) {
              set.comparison(range.name);
            }
          }}
        />
      </div>
    </div>
  </div>
  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start ml-2">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div
      class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 pointer-events-auto"
    >
      {#if !hasFilters}
        <div
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {/if}

      {#each complexFilters as filter, i (i)}
        <AdvancedFilter advancedFilter={filter} />
      {/each}

      {#each dimensions as [id, entry] (id)}
        <DimensionFilter
          {readOnly}
          filterData={entry}
          {timeStart}
          {timeEnd}
          openOnMount={justAdded}
          timeControlsReady={!!$timeRangeStateStore}
          whereFilter={$filterMap}
          onRemove={removeDimensionFilter}
          onToggleFilterMode={toggleDimensionFilterMode}
          onSelect={toggleMultipleDimensionValueSelections}
          onApplyInList={applyDimensionInListMode}
          onApplyContainsMode={applyDimensionContainsMode}
          onPinFilter={pinFilter}
        />
      {/each}

      {#each measureFilters as [id, { name, label, measures, dimensionName, filter, dimensions: dimensionsForMeasure }] (id)}
        {@const metricsViewNames = measures ? Array.from(measures.keys()) : []}
        <div animate:flip={{ duration: 200 }}>
          <MeasureFilter
            allDimensions={dimensionsForMeasure || $allDimensions}
            {name}
            {label}
            {dimensionName}
            {filter}
            onRemove={() =>
              removeMeasureFilter(dimensionName, name, metricsViewNames)}
            onApply={({ dimension, oldDimension, filter }) =>
              handleMeasureFilterApply(
                dimension,
                name,
                oldDimension,
                filter,
                metricsViewNames,
              )}
          />
        </div>
      {/each}

      {#if !readOnly}
        <CanvasFilterButton
          allDimensions={$_allDimensions}
          filteredSimpleMeasures={$_allMeasures}
          dimensionHasFilter={(name) => dimensions.has(name)}
          measureHasFilter={$measureHasFilter}
          setTemporaryFilterName={(n) => {
            justAdded = true;
            addTemporaryFilter(n);
          }}
        />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasClearableFilters}
          <Button type="text" onClick={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
