<script lang="ts">
  import { measureSupportsTotalsQuery } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createQueryServiceMetricsViewAggregation,
    createQueryServiceMetricsViewTimeSeries,
  } from "@rilldata/web-common/runtime-client";
  import type { KPISpec } from ".";
  import { KPI } from ".";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { validateKPISchema } from "./selector";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    spec,
    canvasName,
    expressionFilterManager,
    timeFilterManager,
    visible,
  }: {
    spec: KPISpec;
    canvasName: string;
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager: TimeFilterManager;
    visible: boolean;
  } = $props();

  const client = useRuntimeClient();

  let ctx = $derived(getCanvasStore(canvasName, client.instanceId));
  let {
    metricsView: { getMeasureForMetricView },
  } = $derived(ctx.canvasEntity);

  let {
    metrics_view: metricsViewName,
    measure: measureName,
    sparkline,
    comparison: comparisonOptions,
    hide_time_range: hideTimeRange,
  } = $derived(spec);

  let where = $derived(
    expressionFilterManager.exprByMetricsView[metricsViewName],
  );

  let {
    interval,
    apiTimeRange,
    timeGrain,
    timeZone,

    comparisonTimeRange,
    apiComparisonTimeRange,
    showComparison: showTimeComparison,
    hasTimeSeries,
  } = $derived(timeFilterManager);

  let schema = $derived(validateKPISchema(ctx, spec));
  let { isValid } = $derived($schema);

  let measureStore = $derived(
    getMeasureForMetricView(measureName, metricsViewName),
  );
  let measure = $derived($measureStore);

  // Measures with required dimensions (e.g. a rolling window ordered by the time
  // dimension) produce one value per dimension value and have no single total,
  // so we skip the totals queries; the KPI shows an explanatory hint instead.
  // The measure metadata must have loaded before we can tell, so the queries
  // also wait for it.
  let supportsTotal = $derived(
    !!measure && measureSupportsTotalsQuery(measure),
  );

  let showSparkline = $derived(sparkline !== "none" && hasTimeSeries);

  let showComparison = $derived(
    !!comparisonOptions?.length && showTimeComparison,
  );

  let comparisonLabel = $derived(
    comparisonTimeRange &&
      (TIME_COMPARISON[comparisonTimeRange]?.label as string | undefined),
  );

  let queryMeasures = $derived([{ name: measureName }]);

  let totalQuery = $derived(
    createQueryServiceMetricsViewAggregation(
      client,
      {
        metricsView: metricsViewName,
        measures: queryMeasures,
        timeRange: apiTimeRange,
        where,
        priority: 50,
      },
      {
        query: {
          enabled:
            isValid &&
            supportsTotal &&
            visible &&
            (!hasTimeSeries || (!!apiTimeRange.start && !!apiTimeRange.end)),
        },
      },
    ),
  );

  let comparisonTotalQuery = $derived(
    createQueryServiceMetricsViewAggregation(
      client,
      {
        metricsView: metricsViewName,
        measures: queryMeasures,
        timeRange: apiComparisonTimeRange,
        where,
        priority: 50,
      },
      {
        query: {
          enabled:
            apiComparisonTimeRange &&
            showComparison &&
            isValid &&
            supportsTotal &&
            !!apiTimeRange.start &&
            !!apiTimeRange.end &&
            visible,
        },
      },
    ),
  );

  let primarySparklineQuery = $derived(
    createQueryServiceMetricsViewTimeSeries(
      client,
      {
        metricsViewName,
        measureNames: [measureName],
        timeStart: apiTimeRange.start,
        timeEnd: apiTimeRange.end,
        timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
        timeZone,
        where,
        priority: 10,
      },
      {
        query: {
          enabled:
            !!apiTimeRange.start &&
            !!apiTimeRange.end &&
            isValid &&
            showSparkline &&
            visible,
        },
      },
    ),
  );

  let comparisonSparklineQuery = $derived(
    createQueryServiceMetricsViewTimeSeries(
      client,
      {
        metricsViewName,
        measureNames: [measureName],
        timeStart: apiComparisonTimeRange?.start,
        timeEnd: apiComparisonTimeRange?.end,
        timeGranularity: timeGrain || V1TimeGrain.TIME_GRAIN_HOUR,
        timeZone,
        where,
        priority: 10,
      },
      {
        query: {
          enabled:
            apiComparisonTimeRange &&
            isValid &&
            showSparkline &&
            showComparison &&
            visible,
        },
      },
    ),
  );
</script>

{#if interval}
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
{/if}
