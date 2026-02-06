<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { TimeSeriesPoint } from "./types";
  import { createVisibilityObserver } from "./interactions";
  import { ScrubController } from "./ScrubController";
  import MeasureChartBody from "./MeasureChartBody.svelte";
  import TDDAlternateChart from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/TDDAlternateChart.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import { adjustTimeInterval } from "../utils";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    createQueryServiceMetricsViewTimeSeries,
    type V1Expression,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { keepPreviousData } from "@tanstack/svelte-query";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import { transformTimeSeriesData } from "./use-measure-time-series";
  import {
    createDimensionAggregationQuery,
    buildDimensionSeriesData,
  } from "./use-dimension-data";
  import { createAnnotationsQuery } from "../annotations-selectors";
  import { dateToIndex } from "./utils";
  import { hoverIndex } from "./hover-index";

  const VISIBILITY_ROOT_MARGIN = "120px";

  export let measure: MetricsViewSpecMeasure;
  export let instanceId: string;
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

  // Extract ISO strings for API calls
  $: timeStart = interval?.start?.toISO() ?? undefined;
  $: timeEnd = interval?.end?.toISO() ?? undefined;
  $: comparisonTimeStart = comparisonInterval?.start?.toISO() ?? undefined;
  $: comparisonTimeEnd = comparisonInterval?.end?.toISO() ?? undefined;

  // Time series queries
  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
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
    instanceId,
    metricsViewName,
    {
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
  $: error = ($timeSeriesQuery.error as HTTPError | undefined)?.response?.data
    ?.message;

  // Dimension comparison data
  $: hasDimensionComparison =
    !!comparisonDimension && dimensionValues.length > 0 && !!timeDimension;

  $: dimAggQuery = hasDimensionComparison
    ? createDimensionAggregationQuery(
        instanceId,
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
          instanceId,
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
    instanceId,
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
    if (ts) {
      const idx = dateToIndex(data, ts.getTime());
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
    onScrub?.({
      start: DateTime.fromJSDate(start, { zone: timeZone }),
      end: DateTime.fromJSDate(end, { zone: timeZone }),
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
      <TDDAlternateChart
        chartType={tddChartType}
        expandedMeasureName={measureName}
        timeSeriesPoints={data}
        dimensionSeriesData={dimensionData}
        timeGrain={timeGranularity}
        xMin={undefined}
        xMax={undefined}
        isTimeComparison={showComparison}
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
    />
  {:else}
    <div class="flex items-center justify-center h-full text-gray-400 text-sm">
      No data available
    </div>
  {/if}
</div>
