<script lang="ts">
  import { ephemeralSpecsToDefs } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
  import {
    applyExactMeasures,
    ephemeralMeasureToSpecMeasure,
    exactTimeSeriesMeasureNames,
    mapEphemeralMeasuresForRequest,
    splitTimeSeriesMeasures,
  } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { requestExactNames } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
  import {
    canvasExactMeasureContext,
    canvasPrecisionChip,
  } from "../exact-measures";
  import { measureSupportsTotalsQuery } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
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
  // The exact-selection surface of the KPI (its component); see `canvasExactSurface`.
  export let surface: string;

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

  $: ephemeralMeasures = ephemeralSpecsToDefs(spec.adhoc_measures);
  $: ephemeralDef = ephemeralMeasures?.find((def) => def.name === measureName);

  $: measureStore = getMeasureForMetricView(measureName, metricsViewName);
  // ephemeral measures have no spec entry; synthesize one.
  $: measure =
    $measureStore ??
    (ephemeralDef ? ephemeralMeasureToSpecMeasure(ephemeralDef) : undefined);

  // Measures with required dimensions (e.g. a rolling window ordered by the time
  // dimension) produce one value per dimension value and have no single total,
  // so we skip the totals queries; the KPI shows an explanatory hint instead.
  // The measure metadata must have loaded before we can tell, so the queries
  // also wait for it.
  $: supportsTotal = !!measure && measureSupportsTotalsQuery(measure);

  $: showSparkline = sparkline !== "none" && hasTimeSeries;

  $: showComparison = !!comparisonOptions?.length && showTimeComparison;

  $: comparisonLabel =
    comparisonTimeRangeState?.selectedComparisonTimeRange?.name &&
    (TIME_COMPARISON[comparisonTimeRangeState?.selectedComparisonTimeRange.name]
      ?.label as string | undefined);

  // Exact measures: the selection of this KPI's surface.
  $: exactContextStore = canvasExactMeasureContext(
    ctx.canvasEntity,
    surface,
    metricsViewName,
  );
  $: exact = requestExactNames([measureName], $exactContextStore).has(
    measureName,
  );
  $: exactNames = exact ? new Set([measureName]) : undefined;

  $: queryMeasures = applyExactMeasures(
    mapEphemeralMeasuresForRequest([{ name: measureName }], ephemeralMeasures),
    exactNames,
  );
  $: ({ measureNames: tsMeasureNames, ephemeralMeasures: tsEphemeralMeasures } =
    splitTimeSeriesMeasures([measureName], ephemeralMeasures));
  $: tsExactMeasureNames = exactTimeSeriesMeasureNames(
    tsMeasureNames,
    exactNames,
  );
  $: tsRequestEphemeralMeasures =
    tsEphemeralMeasures && applyExactMeasures(tsEphemeralMeasures, exactNames);

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
        enabled:
          isValid &&
          supportsTotal &&
          visible &&
          (!hasTimeSeries || (!!start && !!end)),
      },
    },
  );

  $: comparisonTotalQuery = createQueryServiceMetricsViewAggregation(
    client,
    {
      metricsView: metricsViewName,
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
          supportsTotal &&
          !!start &&
          !!end &&
          visible,
      },
    },
  );

  $: primarySparklineQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: tsMeasureNames,
      ephemeralMeasures: tsRequestEphemeralMeasures,
      exactMeasureNames: tsExactMeasureNames,
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
      measureNames: tsMeasureNames,
      ephemeralMeasures: tsRequestEphemeralMeasures,
      exactMeasureNames: tsExactMeasureNames,
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

  // Declared after the queries: `$:` blocks run in source order for store
  // reads, so reading `$totalQuery` above its declaration throws.
  $: precision = canvasPrecisionChip(
    measureName,
    measure?.displayName || measureName,
    $exactContextStore,
    !!$totalQuery?.isFetching ||
      (showComparison && !!$comparisonTotalQuery?.isFetching),
    $totalQuery?.error?.message ?? undefined,
    () => ctx.canvasEntity.toggleExactMeasure(surface, measureName),
  );
</script>

<KPI
  {measure}
  {precision}
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
