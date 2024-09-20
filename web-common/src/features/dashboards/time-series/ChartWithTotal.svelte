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
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "../stores/dashboard-stores";
  import { createQueryServiceMetricsViewTimeSeries } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { mergeMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
  import { sanitiseExpression } from "../stores/filter-utils";
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

  const timeControlsStore = useTimeControlStore(getStateManagers());

  export let measure: MetricsViewSpecMeasureV2;
  export let isValidPercTotal: boolean;
  export let parentElement: HTMLElement;
  export let metricViewName: string;
  export let isComparison: boolean;
  export let mouseoverValue: DomainCoordinates<number | Date> | undefined;
  export let expandedMeasureName: string | undefined;
  export let isAllTime: boolean;

  export let scrubStart;
  export let scrubEnd;
  export let startValue;
  export let endValue;
  export let isScrubbing: boolean;

  let visible = false;

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
  let container: HTMLElement;
  onMount(() => {
    observer.observe(container);
  });

  $: isInTimeDimensionView = Boolean(expandedMeasureName);

  $: measureName = measure.name as string;

  $: ({ instanceId } = $runtime);

  $: dashboardStoreStore = useDashboardStore(metricViewName);

  $: dashboardStore = $dashboardStoreStore;

  $: timeControls = $timeControlsStore;

  $: comparisonTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricViewName,
    {
      measureNames: [measureName],
      where: sanitiseExpression(mergeMeasureFilters(dashboardStore), undefined),
      timeStart: timeControls.comparisonAdjustedStart,
      timeEnd: timeControls.comparisonAdjustedEnd,
      timeGranularity:
        timeControls.selectedTimeRange?.interval ?? timeControls.minTimeGrain,
      timeZone: dashboardStore.selectedTimezone,
    },
    {
      query: {
        enabled:
          visible &&
          !!timeControls.ready &&
          !!dashboardStore &&
          // in case of comparison, we need to wait for the comparison start time to be available
          (!isComparison || !!timeControls.comparisonAdjustedStart),

        keepPreviousData: true,
      },
    },
  );

  $: primaryTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricViewName,
    {
      measureNames: [measureName],
      where: sanitiseExpression(mergeMeasureFilters(dashboardStore), undefined),
      timeStart: timeControls.adjustedStart,
      timeEnd: timeControls.adjustedEnd,
      timeGranularity:
        timeControls.selectedTimeRange?.interval ?? timeControls.minTimeGrain,
      timeZone: dashboardStore.selectedTimezone,
    },
    {
      query: {
        enabled: visible && !!timeControls.ready && !!dashboardStore,

        keepPreviousData: true,
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [measure],
      where: sanitiseExpression(mergeMeasureFilters(dashboardStore), undefined),
      timeRange: {
        start: timeControls?.comparisonTimeStart,

        end: timeControls?.comparisonTimeEnd,
      },
    },
    {
      query: {
        enabled:
          visible && !isComparison && !!timeControls.ready && !!dashboardStore,
      },
    },
  );

  $: primaryTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [measure],
      where: sanitiseExpression(mergeMeasureFilters(dashboardStore), undefined),
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

  $: interval =
    timeControls.selectedTimeRange?.interval ??
    timeControls.minTimeGrain ??
    V1TimeGrain.TIME_GRAIN_DAY;

  $: ({
    data: primaryData,
    error: primaryError,
    isFetching: primaryIsFetching,
  } = $primaryTimeSeriesQuery);
  $: comparison = $comparisonTimeSeriesQuery;

  $: intervalDuration = TIME_GRAIN[interval]?.duration as Period;

  $: formattedData = prepareTimeSeries(
    primaryData?.data || [],
    comparison?.data?.data || [],
    intervalDuration,
    dashboardStore.selectedTimezone,
  );

  $: isPercOfTotalAsContextColumn =
    dashboardStore?.leaderboardContextColumn ===
    LeaderboardContextColumn.PERCENT;

  $: ({
    error: primaryTotalError,
    isFetching: primaryTotalIsFetching,
    data: primaryTotalData,
  } = $primaryTotalQuery);

  $: primaryTotal = primaryTotalData?.data?.[0]?.[measureName];

  $: ({ data: comparisonTotalData } = $comparisonTotalQuery);

  $: comparisonTotal = comparisonTotalData?.data?.[0]?.[measureName];
  $: tddChartType = dashboardStore?.tdd?.chartType;

  $: dimensionData = [];
</script>

<div class="flex flex-row gap-x-4" bind:this={container}>
  <MeasureBigNumber
    {measure}
    value={primaryTotal}
    isMeasureExpanded={isInTimeDimensionView}
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
      timeGrain={interval}
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
  {:else if formattedData && interval}
    <MeasureChart
      bind:mouseoverValue
      {measure}
      {isInTimeDimensionView}
      {isScrubbing}
      {scrubStart}
      {scrubEnd}
      {metricViewName}
      data={formattedData}
      {dimensionData}
      zone={dashboardStore.selectedTimezone}
      xAccessor="ts_position"
      labelAccessor="ts"
      timeGrain={interval}
      yAccessor={measure.name}
      xMin={startValue}
      xMax={endValue}
      showComparison={isComparison}
      validPercTotal={isPercOfTotalAsContextColumn && isValidPercTotal
        ? primaryTotal
        : null}
      mouseoverTimeFormat={(value) => {
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
