<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import { flip } from "svelte/animate";
  // import { fly } from "svelte/transition";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import { DimensionFilterMode } from "../../dashboards/filters/dimension-filters/constants";
  // import PreviewButton from "../../explores/PreviewButton.svelte";
  import LeaderboardIcon from "../icons/LeaderboardIcon.svelte";

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
      setDefaultFilters,
      filterManager: {
        _allDimensions,
        _activeUIFilters,
        toggleMultipleDimensionValueSelections,
        toggleDimensionFilterMode,
        applyDimensionInListMode,
        addTemporaryFilter,
        applyDimensionContainsMode,
        removeDimensionFilter,
      },
      filters: {
        whereFilter,
        setMeasureFilter,
        removeMeasureFilter,
        clearAllFilters,
        dimensionHasFilter,
        temporaryFilters,
        allDimensionFilterItems,
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

  $: ({ dimensions } = $_activeUIFilters);

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
      )
    : Interval.invalid("Unable to parse time range");

  $: allDimensionFilters = $allDimensionFilterItems;

  $: allMeasureFilters = $allMeasureFilterItems;

  $: canPan = $_canPan;

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    allDimensionFilters.size + allMeasureFilters.length >
    $temporaryFilters.size;

  $: isComplexFilter = isExpressionUnsupported($whereFilter);

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

    <Button
      label="Preview"
      type="secondary"
      preload={false}
      compact
      onClick={setDefaultFilters}
    >
      <LeaderboardIcon size="16px" color="currentColor" />
      <div class="flex gap-x-1 items-center">Save as default</div>
    </Button>
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
        <!-- {:else if !allDimensionFilters.size && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div> -->
      {:else}
        {#each dimensions as [id, entry] (id)}
          {@const metricsViewNames = Array.from(entry.dimensions.keys())}
          {@const dimension = entry.dimensions.get(metricsViewNames[0])}
          {@const name = dimension?.name || id}
          {#if dimension}
            <DimensionFilter
              {metricsViewNames}
              {readOnly}
              {name}
              label={dimension.displayName ||
                dimension.name ||
                dimension.column ||
                "Unnamed Dimension"}
              mode={entry.mode}
              selectedValues={entry.selectedValues}
              inputText={entry.inputText}
              {timeStart}
              pinned={entry.pinned}
              {timeEnd}
              openOnMount={justAdded}
              timeControlsReady={!!$timeRangeStateStore}
              excludeMode={entry.isInclude === false}
              whereFilter={$whereFilter}
              onRemove={() => removeDimensionFilter(name, metricsViewNames)}
              onToggleFilterMode={() =>
                toggleDimensionFilterMode(name, metricsViewNames)}
              onSelect={(value) =>
                toggleMultipleDimensionValueSelections(
                  name,
                  [value],
                  metricsViewNames,
                  true,
                )}
              onMultiSelect={(values) =>
                toggleMultipleDimensionValueSelections(
                  name,
                  values,
                  metricsViewNames,
                  true,
                )}
              onApplyInList={(values) =>
                applyDimensionInListMode(name, values, metricsViewNames)}
              onApplyContainsMode={(searchText) =>
                applyDimensionContainsMode(name, searchText, metricsViewNames)}
            />
          {/if}
        {/each}
        {#each allMeasureFilters as { name, label, dimensionName, filter, dimensions: dimensionsForMeasure } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensionsForMeasure || $allDimensions}
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
          allDimensions={Array.from($_allDimensions.values())}
          filteredSimpleMeasures={$allSimpleMeasures}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          setTemporaryFilterName={(n) => {
            justAdded = true;
            addTemporaryFilter(n);
          }}
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
