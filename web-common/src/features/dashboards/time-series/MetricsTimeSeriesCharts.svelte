<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/shadcn/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import {
    PivotChipType,
    type PivotChipData,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useExploreState,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import TDDAlternateChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDAlternateChart.svelte";
  import { chartInteractionColumn } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import BackToExplore from "@rilldata/web-common/features/dashboards/time-series/BackToExplore.svelte";
  import {
    useTimeSeriesDataStore,
    type TimeSeriesDatum,
  } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    TimeRangePreset,
    type AvailableTimeGrain,
  } from "@rilldata/web-common/lib/time/types";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import Spinner from "../../entity-management/Spinner.svelte";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import ChartInteractions from "./ChartInteractions.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import type { DimensionDataItem } from "./multiple-dimension-queries";
  import {
    adjustTimeInterval,
    getOrderedStartEnd,
    updateChartInteractionStore,
  } from "./utils";

  export let exploreName: string;
  export let workspaceWidth: number;
  export let timeSeriesWidth: number;
  export let hideStartPivotButton = false;

  const {
    selectors: {
      measures: {
        allMeasures,
        visibleMeasures,
        isMeasureValidPercentOfTotal,
        getMeasureByName,
      },
      dimensionFilters: { includedDimensionValues },
    },
    actions: {
      measures: { toggleMeasureVisibility },
    },
    validSpecStore,
  } = getStateManagers();

  const timeControlsStore = useTimeControlStore(getStateManagers());
  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());

  let scrubStart;
  let scrubEnd;

  let mouseoverValue: DomainCoordinates | undefined = undefined;
  let startValue: Date | undefined;
  let endValue: Date | undefined;

  let dataCopy: TimeSeriesDatum[];
  let dimensionDataCopy: DimensionDataItem[] = [];

  $: exploreState = useExploreState(exploreName);

  $: expandedMeasureName = $exploreState?.tdd?.expandedMeasureName;
  $: isInTimeDimensionView = Boolean(expandedMeasureName);

  $: comparisonDimension = $exploreState?.selectedComparisonDimension;
  $: showComparison = Boolean($timeControlsStore.showTimeComparison);
  $: tddChartType = $exploreState?.tdd?.chartType;

  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;
  $: isScrubbing = $exploreState?.selectedScrubRange?.isScrubbing;
  $: isAllTime =
    $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME;
  $: isPercOfTotalAsContextColumn =
    $exploreState?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;
  $: includedValuesForDimension = $includedDimensionValues(
    comparisonDimension as string,
  );
  $: isAlternateChart = tddChartType !== TDDChart.DEFAULT;

  $: expandedMeasure = $getMeasureByName(expandedMeasureName);
  // List of measures which will be shown on the dashboard
  // List of measures which will be shown on the dashboard
  let renderedMeasures: MetricsViewSpecMeasureV2[];
  $: {
    renderedMeasures = expandedMeasure ? [expandedMeasure] : $visibleMeasures;
  }

  $: totals = $timeSeriesDataStore.total;
  $: totalsComparisons = $timeSeriesDataStore.comparisonTotal;

  // When changing the timeseries query and the cache is empty, $timeSeriesQuery.data?.data is
  // temporarily undefined as results are fetched.
  // To avoid unmounting TimeSeriesBody, which would cause us to lose our tween animations,
  // we make a copy of the data that avoids `undefined` transition states.
  // TODO: instead, try using svelte-query's `keepPreviousData = True` option.

  $: if ($timeSeriesDataStore?.timeSeriesData) {
    dataCopy = $timeSeriesDataStore.timeSeriesData;
  }
  $: formattedData = dataCopy;

  $: if (
    $timeSeriesDataStore?.dimensionChartData?.length ||
    !comparisonDimension ||
    includedValuesForDimension.length === 0
  ) {
    dimensionDataCopy = $timeSeriesDataStore.dimensionChartData || [];
  }
  $: dimensionData = dimensionDataCopy;

  // FIXME: move this logic to a function + write tests.
  $: if ($timeControlsStore.ready && interval) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $exploreState?.selectedScrubRange?.start,
      $exploreState?.selectedTimezone,
      timeGrainToDuration(interval),
    );
    scrubEnd = adjustOffsetForZone(
      $exploreState?.selectedScrubRange?.end,
      $exploreState?.selectedTimezone,
      timeGrainToDuration(interval),
    );

    const slicedData = isAllTime
      ? formattedData?.slice(1)
      : formattedData?.slice(1, -1);

    chartInteractionColumn.update((state) => {
      const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

      let startDirection, endDirection;

      if (
        tddChartType === TDDChart.GROUPED_BAR ||
        tddChartType === TDDChart.STACKED_BAR
      ) {
        startDirection = "left";
        endDirection = "right";
      } else {
        startDirection = "center";
        endDirection = "center";
      }

      const { position: startPos } = bisectData(
        start,
        startDirection,
        "ts_position",
        slicedData,
      );

      const { position: endPos } = bisectData(
        end,
        endDirection,
        "ts_position",
        slicedData,
      );

      return {
        yHover: isScrubbing ? undefined : state.yHover,
        xHover: isScrubbing ? undefined : state.xHover,
        scrubStart: startPos,
        scrubEnd: endPos,
      };
    });

    const adjustedChartValue = getAdjustedChartTime(
      $timeControlsStore.selectedTimeRange?.start,
      $timeControlsStore.selectedTimeRange?.end,
      $exploreState?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name,
      $validSpecStore.data?.explore?.defaultPreset?.timeRange,
      $exploreState?.tdd.chartType,
    );

    if (adjustedChartValue?.start) {
      startValue = adjustedChartValue?.start;
    }
    if (adjustedChartValue?.end) {
      endValue = adjustedChartValue?.end;
    }
  }

  $: if (
    isInTimeDimensionView &&
    formattedData &&
    $timeControlsStore.selectedTimeRange &&
    !isScrubbing
  ) {
    updateChartInteractionStore(
      mouseoverValue?.x,
      undefined,
      isAllTime,
      formattedData,
    );
  }

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);
  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }

  $: hasTotalsError = Object.hasOwn($timeSeriesDataStore?.error, "totals");
  $: hasTimeseriesError = Object.hasOwn(
    $timeSeriesDataStore?.error,
    "timeseries",
  );

  $: activeTimeGrain = $timeControlsStore.selectedTimeRange?.interval;

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $exploreState?.pivot;

    const pivotColumns = splitPivotChips(pivot.columns);
    if (
      pivot.rows.length ||
      pivotColumns.measure.length ||
      pivotColumns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function getTimeDimension() {
    return {
      id: $timeControlsStore.selectedTimeRange?.interval,
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
          title: m.displayName || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot(
      exploreName,
      [getTimeDimension()],
      measures,
    );
  }
</script>

<TimeSeriesChartContainer
  enableFullWidth={isInTimeDimensionView}
  end={endValue}
  start={startValue}
  {workspaceWidth}
  {timeSeriesWidth}
>
  <div class:mb-6={isAlternateChart} class="flex items-center gap-x-1 px-2.5">
    {#if isInTimeDimensionView}
      <BackToExplore />
      <ChartTypeSelector
        hasComparison={Boolean(
          showComparison || includedValuesForDimension.length,
        )}
        {exploreName}
        chartType={tddChartType}
      />
    {:else}
      <DashboardVisibilityDropdown
        category="Measures"
        tooltipText="Choose measures to display"
        onSelect={(name) => toggleMeasureVisibility(allMeasureNames, name)}
        selectableItems={$allMeasures.map(({ name, displayName }) => ({
          name: name || "",
          label: displayName || name || "",
        }))}
        selectedItems={visibleMeasureNames}
        onToggleSelectAll={() => {
          toggleMeasureVisibility(allMeasureNames);
        }}
      />

      {#if !hideStartPivotButton}
        <button
          class="h-6 px-1.5 py-px rounded-sm hover:bg-gray-200 text-gray-700 ml-auto"
          on:click={() => {
            startPivotForTimeseries();
          }}
        >
          Start Pivot
        </button>
      {/if}
    {/if}
  </div>

  <div class="z-10 gap-x-9 flex flex-row pt-4" style:padding-left="118px">
    <div class="relative w-full">
      <ChartInteractions {exploreName} {showComparison} timeGrain={interval} />
      {#if tddChartType === TDDChart.DEFAULT}
        <div class="translate-x-5">
          {#if $exploreState?.selectedTimeRange && startValue && endValue}
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
      class:pb-4={!isInTimeDimensionView}
      class="flex flex-col gap-y-2 overflow-y-scroll h-full max-h-fit"
    >
      <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
      {#each renderedMeasures as measure (measure.name)}
        <!-- FIXME: I can't select the big number by the measure id. -->
        <!-- for bigNum, catch nulls and convert to undefined.  -->
        {@const bigNum = measure.name ? totals?.[measure.name] : undefined}
        {@const comparisonValue = measure.name
          ? totalsComparisons?.[measure.name]
          : undefined}
        {@const isValidPercTotal = measure.name
          ? $isMeasureValidPercentOfTotal(measure.name)
          : false}

        <div class="flex flex-row gap-x-4">
          <MeasureBigNumber
            {measure}
            value={bigNum}
            isMeasureExpanded={isInTimeDimensionView}
            {showComparison}
            {comparisonValue}
            errorMessage={$timeSeriesDataStore?.error?.totals}
            status={hasTotalsError
              ? EntityStatus.Error
              : $timeSeriesDataStore?.isFetching
                ? EntityStatus.Running
                : EntityStatus.Idle}
          />

          {#if hasTimeseriesError}
            <div
              class="flex flex-col p-5 items-center justify-center text-xs ui-copy-muted"
            >
              {#if $timeSeriesDataStore.error?.timeseries}
                <span>
                  Error: {$timeSeriesDataStore.error.timeseries}
                </span>
              {:else}
                <span>Unable to fetch data from the API</span>
              {/if}
            </div>
          {:else if expandedMeasureName && tddChartType != TDDChart.DEFAULT}
            <TDDAlternateChart
              timeGrain={interval}
              chartType={tddChartType}
              {expandedMeasureName}
              totalsData={formattedData}
              {dimensionData}
              xMin={startValue}
              xMax={endValue}
              isTimeComparison={showComparison}
              isScrubbing={Boolean(isScrubbing)}
              on:chart-hover={(e) => {
                const { dimension, ts } = e.detail;

                updateChartInteractionStore(
                  ts,
                  dimension,
                  isAllTime,
                  formattedData,
                );
              }}
              on:chart-brush={(e) => {
                const { interval } = e.detail;
                const { start, end } = adjustTimeInterval(
                  interval,
                  $exploreState?.selectedTimezone,
                );

                metricsExplorerStore.setSelectedScrubRange(exploreName, {
                  start,
                  end,
                  isScrubbing: true,
                });
              }}
              on:chart-brush-end={(e) => {
                const { interval } = e.detail;
                const { start, end } = adjustTimeInterval(
                  interval,
                  $exploreState?.selectedTimezone,
                );

                metricsExplorerStore.setSelectedScrubRange(exploreName, {
                  start,
                  end,
                  isScrubbing: false,
                });
              }}
              on:chart-brush-clear={(e) => {
                const { start, end } = e.detail;

                metricsExplorerStore.setSelectedScrubRange(exploreName, {
                  start,
                  end,
                  isScrubbing: false,
                });
              }}
            />
          {:else if formattedData && interval}
            <MeasureChart
              bind:mouseoverValue
              {measure}
              {isInTimeDimensionView}
              {isScrubbing}
              {scrubStart}
              {scrubEnd}
              {exploreName}
              data={formattedData}
              {dimensionData}
              zone={$exploreState?.selectedTimezone}
              xAccessor="ts_position"
              labelAccessor="ts"
              timeGrain={interval}
              yAccessor={measure.name}
              xMin={startValue}
              xMax={endValue}
              {showComparison}
              validPercTotal={isPercOfTotalAsContextColumn && isValidPercTotal
                ? bigNum
                : null}
              mouseoverTimeFormat={(value) => {
                /** format the date according to the time grain */

                return interval
                  ? new Date(value).toLocaleDateString(
                      undefined,
                      TIME_GRAIN[interval].formatDate,
                    )
                  : value.toString();
              }}
            />
          {:else}
            <div class="flex items-center justify-center w-24">
              <Spinner status={EntityStatus.Running} />
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</TimeSeriesChartContainer>

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
