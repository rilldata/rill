<script lang="ts">
  import {
    createQueryServiceMetricsViewAggregation,
    MetricsViewSpecMeasureV2,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import MeasureBigNumber from "../big-number/MeasureBigNumber.svelte";
  import TDDAlternateChart from "../time-dimension-details/charts/TDDAlternateChart.svelte";
  import MeasureChart from "./MeasureChart.svelte";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import { createQueryServiceMetricsViewTimeSeries } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { mergeMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import {
    createAndExpression,
    filterExpressions,
    matchExpressionByName,
    sanitiseExpression,
  } from "../stores/filter-utils";
  import { getStateManagers } from "../state-managers/state-managers";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import {
    prepareTimeSeries,
    updateChartInteractionStore,
    adjustTimeInterval,
  } from "./utils";
  import { Period } from "@rilldata/web-common/lib/time/types";
  import { TDDChart } from "../time-dimension-details/types";
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import { onMount } from "svelte";
  import { LeaderboardContextColumn } from "../leaderboard-context-column";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import { getDimensionValueTimeSeries } from "./multiple-dimension-queries";
  import { getFilteredMeasuresAndDimensions } from "../state-managers/selectors/measures";
  import { useMetricsView } from "../selectors";

  const StateManagers = getStateManagers();
  const {
    dashboardStore: dashboardStoreReadable,
    selectors: {
      dimensions: { comparisonDimension },
    },
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  export let measure: MetricsViewSpecMeasureV2;
  export let isValidPercTotal: boolean;
  export let parentElement: HTMLElement;
  export let metricViewName: string;
  export let isComparison: boolean;
  export let mouseoverValue: DomainCoordinates<number | Date> | undefined;
  export let expandedMeasureName: string | undefined;
  export let isAllTime: boolean;
  export let scrubStart: Date | undefined;
  export let scrubEnd: Date | undefined;
  export let startValue: Date;
  export let endValue: Date;
  export let isScrubbing: boolean;

  let visible = false;
  let container: HTMLElement;

  const observer = new IntersectionObserver(
    ([entry]) => {
      visible = entry.isIntersecting;
    },
    {
      root: parentElement,
      rootMargin: "120px",
      threshold: 0,
    },
  );

  onMount(() => {
    observer.observe(container);
  });

  $: ({ instanceId } = $runtime);

  $: metricsViewQuery = useMetricsView(instanceId, metricViewName);

  $: metricsView = $metricsViewQuery.data;

  $: isPercOfTotalAsContextColumn =
    dashboardStore?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;

  $: measureName = measure.name as string;

  $: dashboardStore = $dashboardStoreReadable;

  $: timeControls = $timeControlsStore;

  $: timeGranularity =
    timeControls.selectedTimeRange?.interval ??
    timeControls.minTimeGrain ??
    V1TimeGrain.TIME_GRAIN_DAY;

  $: tddChartType = dashboardStore?.tdd?.chartType;

  $: whereFilter = sanitiseExpression(
    mergeMeasureFilters(dashboardStore),
    undefined,
  );

  $: updatedFilter = filterExpressions(
    whereFilter || createAndExpression([]),
    (e) => !matchExpressionByName(e, $comparisonDimension?.name ?? ""),
  );

  $: primaryTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricViewName,
    {
      measureNames: [measureName],
      where: whereFilter,
      timeStart: timeControls.adjustedStart,
      timeEnd: timeControls.adjustedEnd,
      timeGranularity,
      timeZone: dashboardStore.selectedTimezone,
    },
    {
      query: {
        enabled: visible && !!timeControls.ready && !!dashboardStore,
        keepPreviousData: true,
      },
    },
  );

  $: comparisonTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricViewName,
    {
      measureNames: [measureName],
      where: whereFilter,
      timeStart: timeControls.comparisonAdjustedStart,
      timeEnd: timeControls.comparisonAdjustedEnd,
      timeGranularity,
      timeZone: dashboardStore.selectedTimezone,
    },
    {
      query: {
        enabled:
          visible &&
          !!timeControls.ready &&
          !!dashboardStore &&
          (!isComparison || !!timeControls.comparisonAdjustedStart),
        keepPreviousData: true,
      },
    },
  );

  $: primaryTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [measure],
      where: whereFilter,
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    },
    {
      query: {
        enabled: visible && !!timeControls.ready && !!dashboardStore,
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [measure],
      where: whereFilter,
      timeRange: {
        start: timeControls?.comparisonTimeStart,
        end: timeControls?.comparisonTimeEnd,
      },
    },
    {
      query: {
        enabled:
          visible && isComparison && !!timeControls.ready && !!dashboardStore,
      },
    },
  );

  $: unfilteredTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [measure],
      where: updatedFilter,
      timeRange: {
        start: timeControls.timeStart,
        end: timeControls.timeEnd,
      },
    },
    {
      query: {
        enabled: !!timeControls.ready && !!dashboardStore,
      },
    },
  );

  $: ({ data: primaryData, error: primaryError } = $primaryTimeSeriesQuery);
  $: ({ data: comparisonData } = $comparisonTimeSeriesQuery);

  $: ({
    error: primaryTotalError,
    isFetching: primaryTotalIsFetching,
    data: primaryTotalData,
  } = $primaryTotalQuery);
  $: ({ data: comparisonTotalData } = $comparisonTotalQuery);
  $: ({ data: unfilteredTotalData } = $unfilteredTotalQuery);

  $: unfilteredTotal = unfilteredTotalData?.data?.[0]?.[measureName];
  $: primaryTotal = primaryTotalData?.data?.[0]?.[measureName];
  $: comparisonTotal = comparisonTotalData?.data?.[0]?.[measureName];

  $: intervalDuration = TIME_GRAIN[timeGranularity]?.duration as Period;

  $: formattedData = prepareTimeSeries(
    primaryData?.data || [],
    comparisonData?.data || [],
    intervalDuration,
    dashboardStore.selectedTimezone,
  );

  $: ({ measures: filteredMeasures } = getFilteredMeasuresAndDimensions({
    dashboard: dashboardStore,
  })(metricsView ?? {}, metricsView?.measures?.map((m) => m.name ?? "") ?? []));

  $: dimensionDataQuery = getDimensionValueTimeSeries(
    StateManagers,
    filteredMeasures,
    "chart",
    visible,
  );

  $: dimensionData = $dimensionDataQuery;

  $: if (
    expandedMeasureName &&
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
</script>

<div class="flex flex-row gap-x-4w-full" bind:this={container}>
  <MeasureBigNumber
    {measure}
    value={primaryTotal}
    isMeasureExpanded={!!expandedMeasureName}
    showComparison={isComparison}
    comparisonValue={comparisonTotal}
    errorMessage={primaryTotalError?.message}
    status={primaryTotalError
      ? EntityStatus.Error
      : primaryTotalIsFetching
        ? EntityStatus.Running
        : EntityStatus.Idle}
    on:expand-measure={() => {
      metricsExplorerStore.setExpandedMeasureName(metricViewName, measure.name);
    }}
  />

  {#if primaryError}
    <div
      class="flex flex-col p-5 items-center justify-center text-xs ui-copy-muted"
    >
      {#if primaryError.response.data?.message}
        <span>
          Error: {primaryError.response.data.message}
        </span>
      {:else}
        <span>Unable to fetch data from the API</span>
      {/if}
    </div>
  {:else if expandedMeasureName && tddChartType != TDDChart.DEFAULT}
    <TDDAlternateChart
      timeGrain={timeGranularity}
      chartType={tddChartType}
      {expandedMeasureName}
      totalsData={formattedData}
      {dimensionData}
      xMin={startValue}
      xMax={endValue}
      isTimeComparison={isComparison}
      isScrubbing={Boolean(isScrubbing)}
      on:chart-hover={(e) => {
        const { dimension, ts } = e.detail;

        updateChartInteractionStore(ts, dimension, isAllTime, formattedData);
      }}
      on:chart-brush={(e) => {
        const { interval } = e.detail;
        const { start, end } = adjustTimeInterval(
          interval,
          dashboardStore.selectedTimezone,
        );

        metricsExplorerStore.setSelectedScrubRange(metricViewName, {
          start,
          end,
          isScrubbing: true,
        });
      }}
      on:chart-brush-end={(e) => {
        const { interval } = e.detail;
        const { start, end } = adjustTimeInterval(
          interval,
          dashboardStore.selectedTimezone,
        );

        metricsExplorerStore.setSelectedScrubRange(metricViewName, {
          start,
          end,
          isScrubbing: false,
        });
      }}
      on:chart-brush-clear={(e) => {
        const { start, end } = e.detail;

        metricsExplorerStore.setSelectedScrubRange(metricViewName, {
          start,
          end,
          isScrubbing: false,
        });
      }}
    />
  {:else if formattedData && timeGranularity}
    <MeasureChart
      bind:mouseoverValue
      {measure}
      isInTimeDimensionView={!!expandedMeasureName}
      {isScrubbing}
      {scrubStart}
      {scrubEnd}
      {metricViewName}
      data={formattedData}
      {dimensionData}
      zone={dashboardStore.selectedTimezone}
      xAccessor="ts_position"
      labelAccessor="ts"
      timeGrain={timeGranularity}
      yAccessor={measure.name}
      xMin={startValue}
      xMax={endValue}
      showComparison={isComparison}
      validPercTotal={isPercOfTotalAsContextColumn && isValidPercTotal
        ? primaryTotal
        : null}
      mouseoverTimeFormat={(value) => {
        return timeGranularity
          ? new Date(value).toLocaleDateString(
              undefined,
              TIME_GRAIN[timeGranularity].formatDate,
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

{#if expandedMeasureName}
  <hr class="border-t border-gray-200 -ml-4" />
  <TimeDimensionDisplay
    error={primaryError}
    {measure}
    {metricViewName}
    formattedTimeSeriesData={formattedData}
    {primaryTotal}
    {comparisonTotal}
    {unfilteredTotal}
    comparisonDimension={$comparisonDimension}
    showTimeComparison={isComparison}
  />
{/if}
