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
  // Measure to compare against over the primary time range (e.g. a target),
  // instead of the time comparison. Set by the KPI grid from its
  // `measure_comparisons`; it is not part of the `kpi` component spec.
  export let comparisonMeasureName: string | undefined = undefined;

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

  $: comparisonEphemeralDef = ephemeralMeasures?.find(
    (def) => def.name === comparisonMeasureKey,
  );
  $: comparisonMeasureStore = getMeasureForMetricView(
    comparisonMeasureKey,
    metricsViewName,
  );
  $: comparisonMeasure =
    $comparisonMeasureStore ??
    (comparisonEphemeralDef
      ? ephemeralMeasureToSpecMeasure(comparisonEphemeralDef)
      : undefined);

  // A comparison measure without a single total has no value to compare
  // against, so the card shows no comparison at all.
  $: comparisonSupportsTotal =
    !!comparisonMeasure && measureSupportsTotalsQuery(comparisonMeasure);

  $: comparisonEnabled = measureComparison
    ? comparisonSupportsTotal
    : showTimeComparison;

  $: showSparkline = sparkline !== "none" && hasTimeSeries;

  $: showComparison = !!comparisonOptions?.length && comparisonEnabled;

  $: comparisonLabel = measureComparison
    ? (comparisonMeasure?.displayName ?? comparisonMeasureKey)
    : comparisonTimeRangeState?.selectedComparisonTimeRange?.name &&
      (
        TIME_COMPARISON[
          comparisonTimeRangeState?.selectedComparisonTimeRange.name
        ]?.label as string | undefined
      )?.toLowerCase();

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

  // The comparison measure is exact when the surface's selection holds it,
  // so it has the same value as on its own card in the grid.
  $: comparisonExactNames = measureComparison
    ? requestExactNames([comparisonMeasureKey], $exactContextStore)
    : exactNames;
  $: comparisonQueryMeasures = measureComparison
    ? applyExactMeasures(
        mapEphemeralMeasuresForRequest(
          [{ name: comparisonMeasureKey }],
          ephemeralMeasures,
        ),
        comparisonExactNames,
      )
    : queryMeasures;
  $: ({
    measureNames: comparisonTsMeasureNames,
    ephemeralMeasures: comparisonTsEphemeralMeasures,
  } = measureComparison
    ? splitTimeSeriesMeasures([comparisonMeasureKey], ephemeralMeasures)
    : { measureNames: tsMeasureNames, ephemeralMeasures: tsEphemeralMeasures });
  $: comparisonTsExactMeasureNames = exactTimeSeriesMeasureNames(
    comparisonTsMeasureNames,
    comparisonExactNames,
  );
  $: comparisonTsRequestEphemeralMeasures =
    comparisonTsEphemeralMeasures &&
    applyExactMeasures(comparisonTsEphemeralMeasures, comparisonExactNames);

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
      measures: comparisonQueryMeasures,
      timeRange: measureComparison
        ? { start, end, timeZone }
        : comparisonTimeRange,
      where,
      priority: 50,
    },
    {
      query: {
        // Without a comparison interval the time comparison would cover all
        // time; a measure comparison uses the primary range instead.
        enabled: measureComparison
          ? showComparison &&
            isValid &&
            supportsTotal &&
            visible &&
            (!hasTimeSeries || (!!start && !!end))
          : !!comparisonTimeRange?.start &&
            !!comparisonTimeRange?.end &&
            showComparison &&
            isValid &&
            supportsTotal &&
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
      measureNames: tsMeasureNames,
      ephemeralMeasures: tsRequestEphemeralMeasures,
      exactMeasureNames: tsExactMeasureNames,
      timeStart: start,
      timeEnd: end,
      timeGranularity: timeGrain,
      timeZone,
      where,
      priority: 10,
    },
    {
      query: {
        // The sparkline renders only once the grain is known, so wait for it
        // instead of querying at a fallback grain.
        enabled:
          !!start &&
          !!end &&
          !!timeGrain &&
          isValid &&
          showSparkline &&
          visible,
      },
    },
  );

  $: comparisonSparklineQuery = createQueryServiceMetricsViewTimeSeries(
    client,
    {
      metricsViewName,
      measureNames: comparisonTsMeasureNames,
      ephemeralMeasures: comparisonTsRequestEphemeralMeasures,
      exactMeasureNames: comparisonTsExactMeasureNames,
      timeStart: measureComparison ? start : comparisonTimeRange?.start,
      timeEnd: measureComparison ? end : comparisonTimeRange?.end,
      timeGranularity: timeGrain,
      timeZone,
      where,
      priority: 10,
    },
    {
      query: {
        // Wait for the comparison interval and the grain, as the primary sparkline does.
        enabled: measureComparison
          ? !!start &&
            !!end &&
            !!timeGrain &&
            isValid &&
            showSparkline &&
            showComparison &&
            visible
          : !!comparisonTimeRange?.start &&
            !!comparisonTimeRange?.end &&
            !!timeGrain &&
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
  showTimeComparison={comparisonEnabled}
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
