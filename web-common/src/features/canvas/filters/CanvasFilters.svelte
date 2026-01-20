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
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasFilterButton from "../../dashboards/filters/CanvasFilterButton.svelte";
  import { Tooltip } from "bits-ui";
  import Metadata from "../../dashboards/time-controls/super-pill/components/Metadata.svelte";

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
        allDimensionsStore,
        allMeasuresStore,
        activeUIFiltersStore,
        filterMapStore,
        temporaryFilterKeysStore,
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

      timeManager: {
        state: {
          comparisonIntervalStore,
          showTimeComparisonStore,
          timeZoneStore,
          canPanStore,
          set,
          rangeStore,
          grainStore,
          comparisonRangeStore,
          interval: intervalStore,
          minMaxTimeStamps,
        },
        hasTimeSeriesStore,
        largestMinTimeGrain,
        defaultTimeRangeStore,
        timeRangeOptionsStore,
        availableTimeZonesStore,
        allowCustomRangeStore,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: selectedRange = $rangeStore;
  $: interval = $intervalStore;
  $: minTimeGrain = $largestMinTimeGrain;

  $: timeStart = interval?.start.toISO();
  $: timeEnd = interval?.end.toISO();

  $: minMax = $minMaxTimeStamps;

  $: hasTimeSeries = $hasTimeSeriesStore;

  $: minDate = minMax?.min;
  $: maxDate = minMax?.max;
  $: showTimeComparison = $showTimeComparisonStore;

  $: activeTimeZone = $timeZoneStore;
  $: temporaryFilterKeys = $temporaryFilterKeysStore;
  $: comparisonInterval = $comparisonIntervalStore;
  $: comparisonRange = $comparisonRangeStore;

  $: activeTimeGrain = $grainStore;
  $: defaultTimeRange = $defaultTimeRangeStore;
  $: availableTimeZones = $availableTimeZonesStore;
  $: timeRanges = $timeRangeOptionsStore;
  $: allowCustomTimeRange = $allowCustomRangeStore;

  $: ({
    dimensionFilters,
    hasFilters,
    measureFilters,
    complexFilters,
    hasClearableFilters,
  } = $activeUIFiltersStore);

  $: canPan = $canPanStore;

  function onPan(direction: "left" | "right") {
    if (!interval || !selectedRange) return;
    const getPanRange = getPanRangeForTimeRange(
      {
        end: interval?.end.toJSDate(),
        start: interval?.start.toJSDate(),
        name: selectedRange,
      },
      activeTimeZone,
    );
    const panRange = getPanRange(direction);

    if (!panRange) return;
    const { start, end } = panRange;

    if (!activeTimeGrain) return;

    set.range(`${start.toISOString()},${end.toISOString()}`);
  }
</script>

<div
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
>
  {#if hasTimeSeries}
    <div class="p-2 flex justify-between size-full py-0">
      <div class="flex items-center size-full">
        <div class="flex-none h-full pt-1.5 pointer-events-auto">
          <Tooltip.Root openDelay={0}>
            <Tooltip.Trigger class="cursor-default">
              <Calendar size="16px" />
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom" sideOffset={10} class="z-50">
              <Metadata
                timeZone={activeTimeZone}
                timeStart={minDate?.toJSDate()}
                timeEnd={maxDate?.toJSDate()}
              />
            </Tooltip.Content>
          </Tooltip.Root>
        </div>
        <div
          class="flex flex-wrap gap-x-2 gap-y-1.5 pl-2 pointer-events-auto size-full pr-2"
        >
          <SuperPill
            context={canvasName}
            {minDate}
            {maxDate}
            selectedRangeAlias={selectedRange}
            showPivot={false}
            {minTimeGrain}
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
            {minDate}
            {maxDate}
            {comparisonInterval}
            {comparisonRange}
            {interval}
            {selectedRange}
            {activeTimeGrain}
            {activeTimeZone}
            {minTimeGrain}
            {showTimeComparison}
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
  {/if}
  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start ml-2">
    {#if !readOnly}
      <Filter size="16px" className="text-fg-secondary flex-none mt-[5px]" />
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
          openOnMount={!!temporaryFilterKeys.get(id)}
          timeControlsReady={!!interval}
          expressionMap={$filterMapStore}
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
          allDimensions={filterData.dimensions ?? []}
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
          allDimensions={$allDimensionsStore}
          filteredSimpleMeasures={$allMeasuresStore}
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
