<script lang="ts">
  import TDDMeasureChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDChart.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    createQueryServiceMetricsViewTimeSeries,
    V1TimeGrain,
    type V1Expression,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { keepPreviousData } from "@tanstack/svelte-query";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import { DateTime, Interval } from "luxon";
  import { onDestroy, onMount } from "svelte";
  import { createAnnotationsQuery } from "../annotations-selectors";
  import { adjustTimeInterval, localToTimeZoneOffset } from "../utils";
  import { hoverIndex } from "./hover-index";
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

  const VISIBILITY_ROOT_MARGIN = "120px";

  export let measure: MetricsViewSpecMeasure;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let interval: Interval<true> | undefined = undefined;
  export let comparisonInterval: Interval<true> | undefined = undefined;
  export let timeGranularity: V1TimeGrain | undefined = undefined;
  export let timeZone: string = "UTC";
  export let comparisonDimension: string | undefined = undefined;
  export let dimensionValues: (string | null)[] = [];
  export let dimensionWhere: V1Expression | undefined = undefined;
  export let annotationsEnabled: boolean = false;
  export let showComparison: boolean = false;
  export let showTimeDimensionDetail: boolean = false;
  export let ready: boolean = true;
  export let chartScrubInterval: Interval<true> | undefined = undefined;
  export let canPanLeft: boolean = false;
  export let canPanRight: boolean = false;
  export let tddChartType: TDDChart = TDDChart.DEFAULT;
  export let onScrub:
    | ((range: {
        start: DateTime;
        end: DateTime;
        isScrubbing: boolean;
      }) => void)
    | undefined = undefined;
  export let onScrubClear: (() => void) | undefined = undefined;
  export let onPanLeft: (() => void) | undefined = undefined;
  export let onPanRight: (() => void) | undefined = undefined;
  export let scrubController: ScrubController;
  export let connectNulls: boolean = true;
  export let forceLineChart: boolean = false;
  export let dynamicYAxis: boolean = false;

  const client = useRuntimeClient();
  const { visible, observe } = createVisibilityObserver(VISIBILITY_ROOT_MARGIN);

  let container: HTMLDivElement;
  let unobserve: (() => void) | undefined;
  let tddIsScrubbing = false;

  onMount(() => {
    if (container) unobserve = observe(container);
  });

  onDestroy(() => {
    unobserve?.();
  });

  $: measureName = measure.name ?? "";
  $: height = showTimeDimensionDetail ? 245 : 145;

  // Extract ISO strings for API calls (must be UTC for protobuf Timestamp parsing)
  $: timeStart = interval?.start?.toUTC().toISO() ?? undefined;
  $: timeEnd = interval?.end?.toUTC().toISO() ?? undefined;
  $: comparisonTimeStart =
    comparisonInterval?.start?.toUTC().toISO() ?? undefined;
  $: comparisonTimeEnd = comparisonInterval?.end?.toUTC().toISO() ?? undefined;

  // Time series queries
  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: [measureName],
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
  );

  $: comparisonTimeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: [measureName],
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
  );

  // Transform query results
  $: comparisonData =
    showComparison && !$comparisonTimeSeriesQuery.isFetching
      ? $comparisonTimeSeriesQuery.data?.data
      : undefined;

  $: data =
    $timeSeriesQuery.isFetching || !$timeSeriesQuery.data?.data
      ? ([] as TimeSeriesPoint[])
      : transformTimeSeriesData(
          $timeSeriesQuery.data.data,
          comparisonData,
          measureName,
          timeZone,
        );

  $: isError = $timeSeriesQuery.isError;
  $: error = $timeSeriesQuery.error?.message;

  // Dimension comparison data
  $: hasDimensionComparison =
    !!comparisonDimension && dimensionValues.length > 0 && !!timeDimension;

  $: dimAggQuery = hasDimensionComparison
    ? createDimensionAggregationQuery(
        client,
        metricsViewName,
        measureName,
        comparisonDimension!,
        dimensionValues,
        dimensionWhere,
        timeDimension!,
        timeStart,
        timeEnd,
        timeGranularity!,
        timeZone,
        $visible && ready && !!timeStart,
      )
    : undefined;

  $: dimCompAggQuery =
    hasDimensionComparison && showComparison && !!comparisonTimeStart
      ? createDimensionAggregationQuery(
          client,
          metricsViewName,
          measureName,
          comparisonDimension!,
          dimensionValues,
          dimensionWhere,
          timeDimension!,
          comparisonTimeStart,
          comparisonTimeEnd,
          timeGranularity!,
          timeZone,

          $visible && ready && !!comparisonTimeStart,
        )
      : undefined;

  $: dimIsFetching =
    (dimAggQuery ? $dimAggQuery?.isFetching : false) ||
    (dimCompAggQuery ? $dimCompAggQuery?.isFetching : false);

  $: dimensionData =
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
      : [];

  $: isFetching =
    $timeSeriesQuery.isFetching ||
    (showComparison && $comparisonTimeSeriesQuery.isFetching) ||
    !!dimIsFetching;

  // Annotations query
  $: annotationsQuery = createAnnotationsQuery(
    client,
    metricsViewName,
    measureName,
    timeDimension,
    timeStart,
    timeEnd,
    timeGranularity,
    timeZone,
    annotationsEnabled && !!timeStart && !!timeEnd && !!timeGranularity,
  );

  // TDD handlers
  function handleTddHover(
    _dimension: undefined | string | null,
    ts: Date | undefined,
  ) {
    if (ts && !isNaN(ts.getTime())) {
      // The component chart applies adjustDataForTimeZone which shifts epochs
      // so Vega displays correct wall-clock times in the browser's local timezone.
      // Reverse that shift before comparing against the UTC-based data array.
      const adjustedTs = localToTimeZoneOffset(ts, timeZone);
      const idx = dateToIndex(data, adjustedTs.getTime());
      if (idx !== null) hoverIndex.set(idx, "tddChart");
    } else {
      hoverIndex.clear("tddChart");
    }
  }

  function handleTddBrush(_interval: { start: Date; end: Date }) {
    tddIsScrubbing = true;
  }

  function handleTddBrushEnd(interval: { start: Date; end: Date }) {
    tddIsScrubbing = false;
    const { start, end } = adjustTimeInterval(interval, timeZone);
    let startDt = DateTime.fromJSDate(start, { zone: timeZone });
    let endDt = DateTime.fromJSDate(end, { zone: timeZone });

    // Snap to grain boundaries: ceil start, floor end
    if (timeGranularity) {
      const unit = V1TimeGrainToDateTimeUnit[timeGranularity];
      const startFloor = startDt.startOf(unit);
      startDt =
        +startFloor < +startDt ? startFloor.plus({ [unit]: 1 }) : startDt;
      endDt = endDt.startOf(unit);
    }

    // Guard: if brush was within a single grain, snapping can invert the range
    if (+endDt <= +startDt) return;

    onScrub?.({
      start: startDt,
      end: endDt,
      isScrubbing: false,
    });
  }

  function handleTddBrushClear() {
    tddIsScrubbing = false;
    onScrubClear?.();
  }
</script>

<div bind:this={container} class="size-full relative">
  {#if !$visible || (isFetching && data.length === 0)}
    <div
      class="flex items-center justify-center h-[145px]"
      class:h-[245px]={showTimeDimensionDetail}
    >
      <Spinner status={EntityStatus.Running} size="24px" />
    </div>
  {:else if isError}
    <div
      class="flex items-center justify-center text-red-500 text-xs h-[145px]"
      class:h-[245px]={showTimeDimensionDetail}
    >
      {error ?? "Error loading data"}
    </div>
  {:else if tddChartType !== TDDChart.DEFAULT && data.length > 0}
    <div class="w-full" style:height="{height}px">
      <TDDMeasureChart
        chartType={tddChartType}
        {metricsViewName}
        {measure}
        {timeDimension}
        {interval}
        comparisonInterval={showComparison ? comparisonInterval : undefined}
        {timeGranularity}
        {timeZone}
        {where}
        {comparisonDimension}
        {dimensionValues}
        {dimensionData}
        {showComparison}
        {showTimeDimensionDetail}
        isScrubbing={tddIsScrubbing}
        onChartHover={handleTddHover}
        onChartBrush={handleTddBrush}
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
      {onPanLeft}
      {onPanRight}
      {onScrub}
      {onScrubClear}
      {scrubController}
      {metricsViewName}
      {connectNulls}
      {forceLineChart}
      {dynamicYAxis}
    />
  {:else}
    <div class="flex items-center justify-center h-full text-gray-400 text-sm">
      No data available
    </div>
  {/if}
</div>
