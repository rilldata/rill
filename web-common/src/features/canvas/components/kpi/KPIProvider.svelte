<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createQueryServiceMetricsViewAggregation,
    createQueryServiceMetricsViewTimeSeries,
  } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import type { Readable } from "svelte/store";
  import type { KPISpec } from ".";
  import { KPI } from ".";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { validateKPISchema } from "./selector";

  export let spec: KPISpec;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let canvasName: string;
  export let visible: boolean;

  const client = useRuntimeClient();

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: ({
    metricsView: { getMeasureForMetricView },
  } = ctx.canvasEntity);

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    comparison: comparisonOptions,
    comparison_measure: comparisonMeasureName,
    hide_time_range: hideTimeRange,
  } = spec);

  // Compare against another measure over the primary time range, instead of
  // against the same measure over the comparison range.
  $: comparisonMeasureKey = comparisonMeasureName ?? "";
  $: measureComparison = comparisonMeasureKey !== "";

  $: ({
    timeGrain,
    timeRange: { timeZone, start, end },
    where,
    comparisonTimeRange,
    showTimeComparison,
    comparisonTimeRangeState,
    hasTimeSeries,
  } = $timeAndFilterStore);

  $: schema = validateKPISchema(ctx, spec);
  $: ({ isValid } = $schema);

  $: measureStore = getMeasureForMetricView(measureName, metricsViewName);
  $: measure = $measureStore;

  $: comparisonMeasureStore = getMeasureForMetricView(
    comparisonMeasureKey,
    metricsViewName,
  );
  $: comparisonMeasure = $comparisonMeasureStore;

  $: showSparkline = sparkline !== "none" && hasTimeSeries;

  $: showComparison =
    !!comparisonOptions?.length && (showTimeComparison || measureComparison);

  $: comparisonLabel = measureComparison
    ? (comparisonMeasure?.displayName ?? comparisonMeasureKey)
    : comparisonTimeRangeState?.selectedComparisonTimeRange?.name &&
      (
        TIME_COMPARISON[
          comparisonTimeRangeState?.selectedComparisonTimeRange.name
        ]?.label as string | undefined
      )?.toLowerCase();

  $: queryMeasures = [{ name: measureName }];

  $: totalQuery = createQueryServiceMetricsViewAggregation(
    client,
    {
      metricsView: metricsViewName,
      measures: queryMeasures,
      timeRange: {
        start,
        end,
        timeZone,
      },
      where,
      priority: 50,
    },
    {
      query: {
        enabled: isValid && visible && (!hasTimeSeries || (!!start && !!end)),
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    client,
    {
      metricsView: metricsViewName,
      measures: measureComparison
        ? [{ name: comparisonMeasureKey }]
        : queryMeasures,
      timeRange: measureComparison
        ? { start, end, timeZone }
        : comparisonTimeRange,
      where,
      priority: 50,
    },
    {
      query: {
        enabled: measureComparison
          ? showComparison &&
            isValid &&
            visible &&
            (!hasTimeSeries || (!!start && !!end))
          : comparisonTimeRange &&
            showComparison &&
            isValid &&
            !!start &&
            !!end &&
            visible,
      },
    },
  );

  // KPI.svelte reads comparison values keyed by the primary measure name.
  // Only rewritten once the data is in: spreading the result while loading or
  // in error breaks TanStack Query's discriminated union.
  $: comparisonTotalResult = !measureComparison
    ? $comparisonTotalQuery
    : !$comparisonTotalQuery.data
      ? $comparisonTotalQuery
      : {
          ...$comparisonTotalQuery,
          data: {
            ...$comparisonTotalQuery.data,
            data: $comparisonTotalQuery.data.data?.map((row) => ({
              ...row,
              [measureName]: row[comparisonMeasureKey],
            })),
          },
        };

  $: primarySparklineQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: [measureName],
      timeStart: start,
      timeEnd: end,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
      priority: 10,
    },
    {
      query: {
        enabled: !!start && !!end && isValid && showSparkline && visible,
      },
    },
  );

  $: comparisonSparklineQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: measureComparison ? [comparisonMeasureKey] : [measureName],
      timeStart: measureComparison ? start : comparisonTimeRange?.start,
      timeEnd: measureComparison ? end : comparisonTimeRange?.end,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
      priority: 10,
    },
    {
      query: {
        enabled: measureComparison
          ? isValid &&
            showSparkline &&
            showComparison &&
            visible &&
            !!start &&
            !!end
          : comparisonTimeRange &&
            isValid &&
            showSparkline &&
            showComparison &&
            visible,
      },
    },
  );

  $: comparisonSparklineResult = !measureComparison
    ? $comparisonSparklineQuery
    : !$comparisonSparklineQuery.data
      ? $comparisonSparklineQuery
      : {
          ...$comparisonSparklineQuery,
          data: {
            ...$comparisonSparklineQuery.data,
            data: $comparisonSparklineQuery.data.data?.map((point) => ({
              ...point,
              records: point.records && {
                ...point.records,
                [measureName]: (point.records as Record<string, unknown>)[
                  comparisonMeasureKey
                ],
              },
            })),
          },
        };

  $: interval = Interval.fromDateTimes(
    DateTime.fromISO(start ?? "").setZone(timeZone),
    DateTime.fromISO(end ?? "").setZone(timeZone),
  );
</script>

<KPI
  {measure}
  {timeGrain}
  {timeZone}
  showTimeComparison={showTimeComparison || measureComparison}
  {hasTimeSeries}
  {comparisonLabel}
  {interval}
  sparkline={spec.sparkline}
  {hideTimeRange}
  comparisonOptions={spec.comparison}
  primaryTotalResult={$totalQuery}
  {comparisonTotalResult}
  primarySparklineResult={$primarySparklineQuery}
  {comparisonSparklineResult}
/>
