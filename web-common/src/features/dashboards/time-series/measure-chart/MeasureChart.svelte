<script lang="ts">
  import { splitTimeSeriesMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import InlineErrorIndicator from "@rilldata/web-common/features/dashboards/errors/InlineErrorIndicator.svelte";
  import TDDMeasureChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDChart.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    MinSupportedGrain,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import { createQueryServiceMetricsViewTimeSeries } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { keepPreviousData } from "@tanstack/svelte-query";
  import { DateTime, Interval } from "luxon";
  import { onDestroy, onMount } from "svelte";
  import { createAnnotationsQuery } from "../annotations-selectors";
  import { adjustTimeInterval, localToTimeZoneOffset } from "../utils";
  import { resolveEffectiveChartType, usesVegaRenderer } from "./chart-series";
  import { chartBrushStore, chartHoverStore, hoverIndex } from "./hover-index";
  import { createVisibilityObserver } from "./interactions";
  import MeasureChartBody from "./MeasureChartBody.svelte";
  import { ScrubController } from "./ScrubController";
  import type { TimeSeriesPoint } from "./types";
  import {
    buildDimensionSeriesData,
    createDimensionAggregationQuery,
  } from "./use-dimension-data";
  import { transformTimeSeriesData } from "./use-measure-time-series";
  import { dateToIndex } from "./utils";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  const VISIBILITY_ROOT_MARGIN = "120px";

  let {
    measure,
    ephemeralMeasures = undefined,
    metricsViewName,
    expressionFilterManager,
    timeFilterManager,
    comparisonDimension = undefined,
    dimensionValues = [],
    annotationsEnabled = false,
    timeDimension,
    showTimeDimensionDetail = false,
    tddChartType = TDDChart.DEFAULT,
    scrubController = undefined,
    connectNulls = true,
    dynamicYAxis = false,
    tddChartHeight = 245,
  }: {
    measure: MetricsViewSpecMeasure;
    ephemeralMeasures: EphemeralMeasureDef[] | undefined;
    metricsViewName: string;
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    comparisonDimension?: string | undefined;
    dimensionValues?: (string | null)[];
    annotationsEnabled?: boolean;
    timeDimension: string;
    showTimeDimensionDetail?: boolean;
    tddChartType?: TDDChart;
    scrubController?: ScrubController | undefined;
    connectNulls?: boolean;
    dynamicYAxis?: boolean;
    tddChartHeight?: number;
  } = $props();

  const client = useRuntimeClient();
  const { visible, observe } = createVisibilityObserver(VISIBILITY_ROOT_MARGIN);

  let {
    interval,
    timeStart,
    timeEnd,
    timeGrain,
    timeZone,

    showComparison,
    comparisonInterval,
    comparisonTimeStart,
    comparisonTimeEnd,

    scrubInterval,
    canPanLeft,
    canPanRight,

    ready,
  } = $derived(timeFilterManager);
  let timeGranularity = $derived(timeGrain ?? MinSupportedGrain);

  let chartScrubInterval = $derived.by(() => {
    if (!scrubInterval) return undefined;
    const [start, end] =
      scrubInterval.start <= scrubInterval.end
        ? [scrubInterval.start, scrubInterval.end]
        : [scrubInterval.end, scrubInterval.start];
    return Interval.fromDateTimes(start, end) as Interval<true>;
  });

  let where = $derived(
    expressionFilterManager.exprByMetricsView[metricsViewName],
  );
  let dimensionWhere = $derived(
    expressionFilterManager.topLevelJoiner.dimensionOnlyExpr[metricsViewName],
  );

  let container: HTMLDivElement;
  let unobserve: (() => void) | undefined;

  onMount(() => {
    if (container) unobserve = observe(container);
  });

  onDestroy(() => {
    unobserve?.();
  });

  let measureName = $derived(measure.name ?? "");
  let { measureNames: tsMeasureNames, ephemeralMeasures: tsEphemeralMeasures } =
    $derived(splitTimeSeriesMeasures([measureName], ephemeralMeasures));
  let height = $derived(showTimeDimensionDetail ? tddChartHeight : 145);

  // Dimension comparison data
  let hasDimensionComparison = $derived(
    !!comparisonDimension && dimensionValues.length > 0 && !!timeDimension,
  );

  let effectiveChartType = $derived(
    resolveEffectiveChartType(tddChartType, hasDimensionComparison),
  );
  let usesVegaChart = $derived(usesVegaRenderer(effectiveChartType));

  // Seed the shared brush store from the persisted scrub interval so TDD Vega
  // charts can render the brush on mount, chart-type switch, or page refresh.
  $effect(() => {
    const brushStoreEmpty = $chartBrushStore.startMs === undefined;
    if (
      usesVegaChart &&
      brushStoreEmpty &&
      scrubInterval?.start &&
      scrubInterval?.end
    ) {
      const systemTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      const vegaStartMs = scrubInterval.start
        .setZone(systemTimeZone, { keepLocalTime: true })
        .toMillis();
      const vegaEndMs = scrubInterval.end
        .setZone(systemTimeZone, { keepLocalTime: true })
        .toMillis();
      chartBrushStore.set({ startMs: vegaStartMs, endMs: vegaEndMs });
    }
  });

  // Time series queries
  let timeSeriesQuery = $derived(
    createQueryServiceMetricsViewTimeSeries(
      client,
      {
        metricsViewName,
        measureNames: tsMeasureNames,
        ephemeralMeasures: tsEphemeralMeasures,
        where,
        timeDimension,
        timeStart,
        timeEnd,
        timeGranularity,
        timeZone,
      },
      {
        query: {
          enabled: $visible && ready && !!timeStart,
          placeholderData: keepPreviousData,
          refetchOnMount: false,
        },
      },
    ),
  );

  let comparisonTimeSeriesQuery = $derived(
    createQueryServiceMetricsViewTimeSeries(
      client,
      {
        metricsViewName,
        measureNames: tsMeasureNames,
        ephemeralMeasures: tsEphemeralMeasures,
        where,
        timeDimension,
        timeStart: comparisonTimeStart,
        timeEnd: comparisonTimeEnd,
        timeGranularity,
        timeZone,
      },
      {
        query: {
          enabled: $visible && ready && showComparison && !!comparisonTimeStart,
          placeholderData: keepPreviousData,
          refetchOnMount: false,
        },
      },
    ),
  );

  // Transform query results
  let comparisonData = $derived(
    showComparison && !$comparisonTimeSeriesQuery.isFetching
      ? $comparisonTimeSeriesQuery.data?.data
      : undefined,
  );

  let data = $derived(
    $timeSeriesQuery.isFetching || !$timeSeriesQuery.data?.data
      ? ([] as TimeSeriesPoint[])
      : transformTimeSeriesData(
          $timeSeriesQuery.data.data,
          comparisonData,
          measureName,
          timeZone,
        ),
  );

  let isError = $derived($timeSeriesQuery.isError);
  let error = $derived($timeSeriesQuery.error?.message);

  let dimAggQuery = $derived(
    hasDimensionComparison
      ? createDimensionAggregationQuery(
          client,
          metricsViewName,
          measureName,
          ephemeralMeasures,
          comparisonDimension!,
          dimensionValues,
          dimensionWhere,
          timeDimension,
          timeStart,
          timeEnd,
          timeGranularity,
          timeZone,
          $visible && ready && !!timeStart,
        )
      : undefined,
  );

  let dimCompAggQuery = $derived(
    hasDimensionComparison && showComparison && !!comparisonTimeStart
      ? createDimensionAggregationQuery(
          client,
          metricsViewName,
          measureName,
          ephemeralMeasures,
          comparisonDimension!,
          dimensionValues,
          dimensionWhere,
          timeDimension,
          comparisonTimeStart,
          comparisonTimeEnd,
          timeGranularity,
          timeZone,

          $visible && ready && !!comparisonTimeStart,
        )
      : undefined,
  );

  let dimIsFetching = $derived(
    (dimAggQuery ? $dimAggQuery?.isFetching : false) ||
      (dimCompAggQuery ? $dimCompAggQuery?.isFetching : false),
  );

  let dimensionData = $derived(
    hasDimensionComparison && timeDimension && timeGranularity
      ? buildDimensionSeriesData(
          measureName,
          comparisonDimension!,
          dimensionValues,
          timeDimension,
          timeGranularity,
          timeZone,
          $timeSeriesQuery.data?.data,
          dimAggQuery ? $dimAggQuery?.data?.data : undefined,
          showComparison ? $comparisonTimeSeriesQuery.data?.data : undefined,
          dimCompAggQuery ? $dimCompAggQuery?.data?.data : undefined,
          !!dimIsFetching,
        )
      : [],
  );

  let isFetching = $derived(
    $timeSeriesQuery.isFetching ||
      (showComparison && $comparisonTimeSeriesQuery.isFetching) ||
      !!dimIsFetching,
  );

  // Annotations query
  let annotationsQuery = $derived(
    createAnnotationsQuery(
      client,
      metricsViewName,
      measureName,
      timeDimension,
      timeStart,
      timeEnd,
      timeGranularity,
      timeZone,
      annotationsEnabled && !!timeStart && !!timeEnd && !!timeGranularity,
    ),
  );

  // TDD handlers
  function handleTddHover(
    dimension: undefined | string | null,
    ts: Date | undefined,
  ) {
    if (ts && !isNaN(ts.getTime())) {
      // The component chart applies adjustDataForTimeZone which shifts epochs
      // so Vega displays correct wall-clock times in the browser's local timezone.
      // Reverse that shift before comparing against the UTC-based data array.
      const adjustedTs = localToTimeZoneOffset(ts, timeZone);
      const idx = dateToIndex(data, adjustedTs.getTime());
      if (idx !== null) hoverIndex.set(idx, "tddChart");
      // Propagate to sibling TDD Vega charts in Explore
      chartHoverStore.set({ dimensionValue: dimension ?? undefined, time: ts });
    } else {
      hoverIndex.clear("tddChart");
      chartHoverStore.set({ dimensionValue: undefined, time: undefined });
    }
  }

  function handleTddBrushEnd(interval: { start: Date; end: Date }) {
    // Write raw Vega epoch values to shared brush store for visual sync across
    // sibling charts. These are unsnapped so drags don't progressively shrink.
    chartBrushStore.set({
      startMs: interval.start.getTime(),
      endMs: interval.end.getTime(),
    });

    // Snap to grain boundaries for the dashboard store
    const { start, end } = adjustTimeInterval(interval, timeZone);
    let startDt = DateTime.fromJSDate(start, { zone: timeZone });
    let endDt = DateTime.fromJSDate(end, { zone: timeZone });

    if (timeGranularity) {
      const unit = V1TimeGrainToDateTimeUnit[timeGranularity];
      const startFloor = startDt.startOf(unit);
      startDt =
        +startFloor < +startDt ? startFloor.plus({ [unit]: 1 }) : startDt;
      endDt = endDt.startOf(unit);
    }

    // Guard: if brush was within a single grain, snapping can invert the range
    if (+endDt <= +startDt) return;

    timeFilterManager.onScrubRange({
      start: startDt,
      end: endDt,
      isScrubbing: false,
    });
  }

  function handleTddBrushClear() {
    chartBrushStore.set({ startMs: undefined, endMs: undefined });
    timeFilterManager.resetScrubRange();
  }
</script>

<div bind:this={container} class="size-full relative">
  {#if !$visible || (isFetching && data.length === 0)}
    <div class="flex items-center justify-center" style:height="{height}px">
      <Spinner status={EntityStatus.Running} size="24px" />
    </div>
  {:else if isError}
    <div class="flex items-center justify-center" style:height="{height}px">
      <InlineErrorIndicator message={error} />
    </div>
  {:else if usesVegaChart && data.length > 0}
    <div class="w-full" style:height="{height}px">
      <TDDMeasureChart
        chartType={effectiveChartType}
        {metricsViewName}
        {measure}
        {expressionFilterManager}
        {timeFilterManager}
        {ephemeralMeasures}
        {timeDimension}
        {comparisonDimension}
        {dimensionValues}
        {dimensionData}
        {showTimeDimensionDetail}
        {dynamicYAxis}
        onChartHover={handleTddHover}
        onChartBrushEnd={handleTddBrushEnd}
        onChartBrushClear={handleTddBrushClear}
      />
    </div>
  {:else if data.length > 0}
    <MeasureChartBody
      {measure}
      {measureName}
      {data}
      {dimensionData}
      annotations={$annotationsQuery.data ?? []}
      {showComparison}
      {showTimeDimensionDetail}
      {timeGranularity}
      {interval}
      {comparisonInterval}
      {chartScrubInterval}
      {canPanLeft}
      {canPanRight}
      onPanLeft={() => timeFilterManager.onPan("left")}
      onPanRight={() => timeFilterManager.onPan("right")}
      onScrub={(range) => timeFilterManager.onScrubRange(range)}
      onScrubClear={() => timeFilterManager.resetScrubRange()}
      {scrubController}
      {metricsViewName}
      {connectNulls}
      {dynamicYAxis}
      {tddChartType}
      {tddChartHeight}
    />
  {:else}
    <div class="flex items-center justify-center h-full text-gray-400 text-sm">
      No data available
    </div>
  {/if}
</div>
