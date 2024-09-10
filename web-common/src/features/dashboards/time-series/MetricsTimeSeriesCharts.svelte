<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import SearchableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterButton.svelte";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import {
    PivotChipData,
    PivotChipType,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { createShowHideMeasuresStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import { chartInteractionColumn } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import BackToOverview from "@rilldata/web-common/features/dashboards/time-series/BackToOverview.svelte";
  import {
    TimeSeriesDatum,
    useTimeSeriesDataStore,
  } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
  import { getAdjustedChartTime } from "@rilldata/web-common/lib/time/ranges";
  import {
    AvailableTimeGrain,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { TIME_GRAIN } from "../../../lib/time/config";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ChartInteractions from "./ChartInteractions.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";
  import type { DimensionDataItem } from "./multiple-dimension-queries";
  import { getOrderedStartEnd, updateChartInteractionStore } from "./utils";
  import ChartWithTotal from "./ChartWithTotal.svelte";

  export let metricViewName: string;
  export let workspaceWidth: number;

  const {
    selectors: {
      measures: {
        isMeasureValidPercentOfTotal,
        getFilteredMeasuresAndDimensions,
      },
      dimensionFilters: { includedDimensionValues },
    },
  } = getStateManagers();

  const timeControlsStore = useTimeControlStore(getStateManagers());
  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());

  let scrubStart;
  let scrubEnd;

  let mouseoverValue: DomainCoordinates | undefined = undefined;
  let startValue: Date;
  let endValue: Date;

  let dataCopy: TimeSeriesDatum[];
  let dimensionDataCopy: DimensionDataItem[] = [];

  $: dashboardStore = useDashboardStore(metricViewName);
  $: instanceId = $runtime.instanceId;

  // query the `/meta` endpoint to get the measures and the default time grain
  $: metricsView = useMetricsView(instanceId, metricViewName);

  $: showHideMeasures = createShowHideMeasuresStore(
    metricViewName,
    metricsView,
  );

  $: expandedMeasureName = $dashboardStore?.tdd?.expandedMeasureName;
  $: isInTimeDimensionView = Boolean(expandedMeasureName);
  $: comparisonDimension = $dashboardStore?.selectedComparisonDimension;
  $: showComparison = Boolean(
    !comparisonDimension && $timeControlsStore.showTimeComparison,
  );
  $: tddChartType = $dashboardStore?.tdd?.chartType;
  $: interval =
    $timeControlsStore.selectedTimeRange?.interval ??
    $timeControlsStore.minTimeGrain;
  $: isScrubbing = Boolean($dashboardStore?.selectedScrubRange?.isScrubbing);
  $: isAllTime =
    $timeControlsStore.selectedTimeRange?.name === TimeRangePreset.ALL_TIME;
  $: includedValuesForDimension = $includedDimensionValues(
    comparisonDimension as string,
  );
  $: isAlternateChart = tddChartType != TDDChart.DEFAULT;

  // List of measures which will be shown on the dashboard
  let renderedMeasures: MetricsViewSpecMeasureV2[];
  $: {
    renderedMeasures =
      $metricsView.data?.measures?.filter(
        expandedMeasureName
          ? (measure) => measure.name === expandedMeasureName
          : (_, i) => $showHideMeasures.selectedItems[i],
      ) ?? [];
    const { measures } = $getFilteredMeasuresAndDimensions(
      $metricsView.data ?? {},
      renderedMeasures.map((m) => m.name ?? ""),
    );
    renderedMeasures = renderedMeasures.filter((rm) =>
      measures.includes(rm.name ?? ""),
    );
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
  $: if ($timeControlsStore.ready) {
    // adjust scrub values for Javascript's timezone changes
    scrubStart = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore.selectedTimezone,
    );
    scrubEnd = adjustOffsetForZone(
      $dashboardStore?.selectedScrubRange?.end,
      $dashboardStore.selectedTimezone,
    );

    const slicedData = isAllTime
      ? formattedData?.slice(1)
      : formattedData?.slice(1, -1);
    chartInteractionColumn.update((state) => {
      const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

      const { position: startPos } = bisectData(
        start,
        "center",
        "ts_position",
        slicedData,
      );
      const { position: endPos } = bisectData(
        end,
        "center",
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
      $dashboardStore?.selectedTimezone,
      interval,
      $timeControlsStore.selectedTimeRange?.name,
      $metricsView?.data?.defaultTimeRange,
      $dashboardStore?.tdd.chartType,
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

  const toggleMeasureVisibility = (e) => {
    showHideMeasures.toggleVisibility(e.detail.name);
  };
  const setAllMeasuresNotVisible = () => {
    showHideMeasures.setAllToNotVisible();
  };
  const setAllMeasuresVisible = () => {
    showHideMeasures.setAllToVisible();
  };

  $: hasTotalsError = Object.hasOwn($timeSeriesDataStore?.error, "totals");
  $: hasTimeseriesError = Object.hasOwn(
    $timeSeriesDataStore?.error,
    "timeseries",
  );

  $: activeTimeGrain = $timeControlsStore.selectedTimeRange?.interval;

  let showReplacePivotModal = false;
  function startPivotForTimeseries() {
    const pivot = $dashboardStore?.pivot;

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
          title: m.label || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot(
      metricViewName,
      { dimension: [getTimeDimension()] },
      {
        dimension: [],
        measure: measures,
      },
    );
  }

  let parentElement: HTMLDivElement;
</script>

<TimeSeriesChartContainer
  enableFullWidth={isInTimeDimensionView}
  end={endValue}
  start={startValue}
  {workspaceWidth}
>
  <div class:mb-6={isAlternateChart} class="flex items-center gap-x-1 px-2.5">
    {#if isInTimeDimensionView}
      <BackToOverview {metricViewName} />
      <ChartTypeSelector
        hasComparison={Boolean(
          showComparison || includedValuesForDimension.length,
        )}
        {metricViewName}
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

  {#if tddChartType == TDDChart.DEFAULT || !expandedMeasureName}
    <div class="z-10 gap-x-9 flex flex-row pt-4" style:padding-left="118px">
      <div class="relative w-full">
        <ChartInteractions
          {metricViewName}
          {showComparison}
          timeGrain={interval}
        />
        <div class="translate-x-5">
          {#if $dashboardStore?.selectedTimeRange && startValue && endValue}
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
      </div>
    </div>
  {/if}

  <!-- bignumbers and line charts -->
  {#if renderedMeasures}
    <div
      bind:this={parentElement}
      class:pb-4={!isInTimeDimensionView}
      class="flex flex-col gap-y-2 overflow-y-scroll h-full max-h-fit"
    >
      <!-- FIXME: this is pending the remaining state work for show/hide measures and dimensions -->
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
          {metricViewName}
          isValidPercTotal={measure.name
            ? $isMeasureValidPercentOfTotal(measure.name)
            : false}
        />
      {/each}
    </div>
  {/if}
</TimeSeriesChartContainer>
<ReplacePivotDialog
  open={showReplacePivotModal}
  on:close={() => {
    showReplacePivotModal = false;
  }}
  on:replace={() => createPivot()}
/>
