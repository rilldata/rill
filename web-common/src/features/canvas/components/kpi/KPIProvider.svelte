<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import {
    createQueryServiceMetricsViewAggregation,
    createQueryServiceMetricsViewTimeSeries,
    V1TimeGrain,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import type { Readable } from "svelte/motion";
  import type { KPISpec } from ".";
  import { KPI } from ".";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { validateKPISchema } from "./selector";

  export let spec: KPISpec;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let canvasName: string;
  export let visible: boolean;

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: ({
    metricsView: { getMeasureForMetricView },
  } = ctx.canvasEntity);

  $: ({ instanceId } = $runtime);

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    comparison: comparisonOptions,
    hide_time_range: hideTimeRange,
  } = spec);

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

  $: showSparkline = sparkline !== "none" && hasTimeSeries;

  $: showComparison = !!comparisonOptions?.length && showTimeComparison;

  $: comparisonLabel =
    comparisonTimeRangeState?.selectedComparisonTimeRange?.name &&
    (TIME_COMPARISON[comparisonTimeRangeState?.selectedComparisonTimeRange.name]
      ?.label as string | undefined);

  $: queryMeasures = [{ name: measureName }];

  $: totalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
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
        enabled: isValid && !!start && !!end && visible,
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: queryMeasures,
      timeRange: comparisonTimeRange,
      where,
      priority: 50,
    },
    {
      query: {
        enabled:
          comparisonTimeRange &&
          showComparison &&
          isValid &&
          !!start &&
          !!end &&
          visible,
      },
    },
  );

  $: primarySparklineQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsViewName,
    {
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
    instanceId,
    metricsViewName,
    {
      measureNames: [measureName],
      timeStart: comparisonTimeRange?.start,
      timeEnd: comparisonTimeRange?.end,
      timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
      timeZone,
      where,
      priority: 10,
    },
    {
      query: {
        enabled:
          comparisonTimeRange &&
          isValid &&
          showSparkline &&
          showComparison &&
          visible,
      },
    },
  );

  $: interval = Interval.fromDateTimes(
    DateTime.fromISO(start ?? "").setZone(timeZone),
    DateTime.fromISO(end ?? "").setZone(timeZone),
  );
</script>

<KPI
  {measure}
  {timeGrain}
  {timeZone}
  {showTimeComparison}
  {hasTimeSeries}
  {comparisonLabel}
  {interval}
  sparkline={spec.sparkline}
  {hideTimeRange}
  comparisonOptions={spec.comparison}
  primaryTotalResult={$totalQuery}
  comparisonTotalResult={$comparisonTotalQuery}
  primarySparklineResult={$primarySparklineQuery}
  comparisonSparklineResult={$comparisonSparklineQuery}
/>
