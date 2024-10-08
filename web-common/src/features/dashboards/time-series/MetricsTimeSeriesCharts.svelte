<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import SearchableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterButton.svelte";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import {
    type PivotChipData,
    PivotChipType,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useExploreStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    type AvailableTimeGrain,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import {
    V1TimeGrain,
    type MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import { TIME_GRAIN } from "../../../lib/time/config";

  import ChartInteractions from "./ChartInteractions.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import ChartWithTotal from "./ChartWithTotal.svelte";
  import { mergeMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import { sanitiseExpression } from "../stores/filter-utils";
  import { adjustTimeInterval, updateChartInteractionStore } from "./utils";

  export let exploreName: string;
  export let workspaceWidth: number;

  const StateManagers = getStateManagers();
  const {
    metricsViewName,
    selectors: {
      measures: {
        allMeasures,
        isMeasureValidPercentOfTotal,
        getFilteredMeasuresAndDimensions,
      },
      dimensionFilters: { includedDimensionValues },
      dimensions: { comparisonDimension },
    },
    validSpecStore,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: timeControls = $timeControlsStore;

  let scrubStart: Date | undefined = undefined;
  let scrubEnd: Date | undefined = undefined;
  let mouseoverValue: DomainCoordinates | undefined = undefined;
  let parentElement: HTMLDivElement;

  $: exploreStore = useExploreStore(exploreName);

  $: showHideMeasures = createShowHideMeasuresStore(
    exploreName,
    validSpecStore,
  );

  $: expandedMeasureName = $exploreStore?.tdd?.expandedMeasureName;
  $: isInTimeDimensionView = Boolean(expandedMeasureName);

  $: leaderboardContextColumn = $exploreStore.leaderboardContextColumn;

  $: showComparison = Boolean(
    !$comparisonDimension && timeControls.showTimeComparison,
  );
  $: tddChartType = $exploreStore?.tdd?.chartType;
  $: interval =
    timeControls.selectedTimeRange?.interval ??
    timeControls.minTimeGrain ??
    V1TimeGrain.TIME_GRAIN_DAY;
  $: isScrubbing = Boolean($exploreStore?.selectedScrubRange?.isScrubbing);
  $: isAllTime =
    timeControls.selectedTimeRange?.name === TimeRangePreset.ALL_TIME;
  $: includedValuesForDimension = $comparisonDimension?.name
    ? $includedDimensionValues($comparisonDimension.name)
    : [];
  $: isAlternateChart = tddChartType !== TDDChart.DEFAULT;

  // List of measures which will be shown on the dashboard
  let renderedMeasures: MetricsViewSpecMeasureV2[];
  $: {
    renderedMeasures = $allMeasures.filter(
      expandedMeasureName
        ? (measure) => measure.name === expandedMeasureName
        : (_, i) => $showHideMeasures.selectedItems[i],
    );
    const { measures } = $getFilteredMeasuresAndDimensions(
      $validSpecStore.data?.metricsView ?? {},
      renderedMeasures.map((m) => m.name ?? ""),
    );
    renderedMeasures = renderedMeasures.filter((rm) =>
      measures.includes(rm.name ?? ""),
    );
  }

  $: if (timeControls.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $exploreStore?.selectedScrubRange?.start,
      $exploreStore.selectedTimezone,
    );
    scrubEnd = adjustOffsetForZone(
      $exploreStore?.selectedScrubRange?.end,
      $exploreStore.selectedTimezone,
    );
  }

  $: ({ start: startValue, end: endValue } = getAdjustedChartTime(
    timeControls.selectedTimeRange?.start,
    timeControls.selectedTimeRange?.end,
    $exploreStore?.selectedTimezone,
    interval,
    timeControls.selectedTimeRange?.name,
    $validSpecStore.data?.explore?.presets?.[0]?.timeRange,
    $exploreStore?.tdd.chartType,
  ));

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };

  $: activeTimeGrain = timeControls.selectedTimeRange?.interval;

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $exploreStore?.pivot;

    if (
      pivot.rows.dimension.length ||
      pivot.columns.measure.length ||
      pivot.columns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function getTimeDimension() {
    return {
      id: timeControls.selectedTimeRange?.interval,
      title: TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label,
      type: PivotChipType.Time,
    } as PivotChipData;
  }

  function createPivot() {
    showReplacePivotModal = false;

    const measures = renderedMeasures
      .filter((m) => m.name !== undefined)
      .map((m) => {
        return {
          id: m.name as string,
          title: m.label || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot(
      exploreName,
      { dimension: [getTimeDimension()] },
      {
        dimension: [],
        measure: measures,
      },
    );
  }
</script>

{#if startValue && endValue}
  <TimeSeriesChartContainer
    enableFullWidth={isInTimeDimensionView}
    end={endValue}
    start={startValue}
    {workspaceWidth}
  >
    <div class:mb-6={isAlternateChart} class="flex items-center gap-x-1 px-2.5">
      {#if isInTimeDimensionView}
        <BackToOverview {exploreName} />
        <ChartTypeSelector
          hasComparison={Boolean(
            showComparison || includedValuesForDimension.length,
          )}
          {exploreName}
          chartType={tddChartType}
        />
      {:else}
        <SearchableFilterButton
          label="Measures"
          on:deselect-all={setAllMeasuresNotVisible}
          on:item-clicked={toggleMeasureVisibility}
          on:select-all={setAllMeasuresVisible}
          selectableItems={$showHideMeasures.selectableItems}
          selectedItems={$showHideMeasures.selectedItems}
          tooltipText="Choose measures to display"
        />

        <button
          class="h-6 px-1.5 py-px rounded-sm hover:bg-gray-200 text-gray-700 ml-auto"
          on:click={() => {
            startPivotForTimeseries();
          }}
        >
          Start Pivot
        </button>
      {/if}
    </div>

    <div class="z-10 gap-x-9 flex flex-row pt-4" style:padding-left="118px">
      <div class="relative w-full">
        <ChartInteractions
          {exploreName}
          {showComparison}
          timeGrain={interval}
        />
        {#if tddChartType === TDDChart.DEFAULT}
          <div class="translate-x-5">
            {#if $exploreStore?.selectedTimeRange && startValue && endValue}
              <SimpleDataGraphic
                height={26}
                overflowHidden={false}
                top={29}
                bottom={0}
                right={isInTimeDimensionView ? 10 : 25}
                xMin={startValue}
                xMax={endValue}
              >
                <Axis superlabel side="top" placement="start" />
              </SimpleDataGraphic>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    {#if renderedMeasures}
      <div
        bind:this={parentElement}
        class:pb-4={!isInTimeDimensionView}
        class="flex flex-col gap-y-2 w-full overflow-y-scroll h-full"
      >
        {#each renderedMeasures as measure (measure.name)}
          <ChartWithTotal
            {isScrubbing}
            {scrubStart}
            {scrubEnd}
            {startValue}
            {endValue}
            {isAllTime}
            {parentElement}
            bind:mouseoverValue
            {expandedMeasureName}
            isComparison={showComparison}
            {measure}
            metricsViewName={$metricsViewName}
            {exploreName}
            {timeControls}
            timeGrain={interval}
            {tddChartType}
            {leaderboardContextColumn}
            onExpandMeasure={() => {
              metricsExplorerStore.setExpandedMeasureName(
                exploreName,
                measure.name,
              );
            }}
            whereFilter={sanitiseExpression(
              mergeMeasureFilters($exploreStore),
              undefined,
            )}
            selectedTimeZone={$exploreStore.selectedTimezone}
            filteredMeasures={$getFilteredMeasuresAndDimensions(
              $validSpecStore.data?.metricsView ?? {},
              $allMeasures.map((m) => m.name ?? "") ?? [],
            ).measures}
            isValidPercTotal={measure.name
              ? $isMeasureValidPercentOfTotal(measure.name)
              : false}
            comparisonDimension={$comparisonDimension}
            stateManagers={StateManagers}
            onChartHover={(dimension, ts, formattedData) => {
              updateChartInteractionStore(
                ts,
                dimension,
                isAllTime,
                formattedData,
              );
            }}
            onChartBrush={(interval) => {
              const { start, end } = adjustTimeInterval(
                interval,
                $exploreStore.selectedTimezone,
              );

              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: true,
              });
            }}
            onChartBrushEnd={(interval) => {
              const { start, end } = adjustTimeInterval(
                interval,
                $exploreStore.selectedTimezone,
              );

              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: false,
              });
            }}
            onChartBrushClear={({ start, end }) => {
              metricsExplorerStore.setSelectedScrubRange(exploreName, {
                start,
                end,
                isScrubbing: false,
              });
            }}
          />
        {/each}
      </div>
    {/if}
  </TimeSeriesChartContainer>
{/if}

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
