<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasFilterButton from "../../dashboards/filters/CanvasFilterButton.svelte";

  export let readOnly = false;
  export let maxWidth: number;
  export let builder = false;
  export let canvasName: string;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  let showDefaultItem = false;

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: {
      filterManager: {
        _allDimensions,
        _allMeasures,
        _activeUIFilters,
        _filterMap,
        _temporaryFilterKeys,
        actions: {
          toggleDimensionValueSelections,
          toggleDimensionFilterMode,
          applyDimensionInListMode,
          addTemporaryFilter,
          applyDimensionContainsMode,
          removeDimensionFilter,
          setMeasureFilter,
          removeMeasureFilter,
          toggleFilterPin,
        },
        clearAllFilters,
      },

      metricsView: { allDimensions },
      timeControls: {
        _canPan,
        allTimeRange,
        timeRangeStateStore,
        comparisonRangeStateStore,
        selectedTimezone,
        minTimeGrain,
        set,
        _defaultTimeRange,
        _timeRangeOptions,
        _availableTimeZones,
        _allowCustomRange,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: activeTimeZone = $selectedTimezone;
  $: temporaryFilterKeys = $_temporaryFilterKeys;

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $_defaultTimeRange;
  $: availableTimeZones = $_availableTimeZones;
  $: timeRanges = $_timeRangeOptions;
  $: allowCustomTimeRange = $_allowCustomRange;

  $: ({
    dimensionFilters,
    hasFilters,
    measureFilters,
    complexFilters,
    hasClearableFilters,
  } = $_activeUIFilters);

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.invalid("Unable to parse time range");

  $: canPan = $_canPan;

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
          {allowCustomTimeRange}
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

      {#each dimensionFilters as [id, filterData] (id)}
        <DimensionFilter
          {readOnly}
          {filterData}
          {timeStart}
          {timeEnd}
          openOnMount={temporaryFilterKeys.has(id)}
          timeControlsReady={!!$timeRangeStateStore}
          expressionMap={$_filterMap}
          {removeDimensionFilter}
          {toggleDimensionFilterMode}
          {toggleDimensionValueSelections}
          {applyDimensionInListMode}
          {applyDimensionContainsMode}
          toggleFilterPin={builder ? toggleFilterPin : undefined}
        />
      {/each}

      {#each measureFilters as [id, filterData] (id)}
        {@const metricsViewNames = filterData.measures
          ? Array.from(filterData.measures.keys())
          : []}

        <MeasureFilter
          {filterData}
          allDimensions={filterData.dimensions || $allDimensions}
          openOnMount={temporaryFilterKeys.has(id)}
          onRemove={async () => {
            await removeMeasureFilter(
              filterData.dimensionName,
              filterData.name,
              metricsViewNames,
            );
          }}
          onApply={({ dimension, filter, oldDimension }) =>
            setMeasureFilter(dimension, filter, oldDimension, metricsViewNames)}
          toggleFilterPin={builder ? toggleFilterPin : undefined}
        />
      {/each}

      {#if !readOnly}
        <CanvasFilterButton
          allDimensions={$_allDimensions}
          filteredSimpleMeasures={$_allMeasures}
          dimensionHasFilter={(name) => dimensionFilters.has(name)}
          measureHasFilter={(name) => measureFilters.has(name)}
          setTemporaryFilterName={addTemporaryFilter}
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
