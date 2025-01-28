import type { KPISpec } from "@rilldata/web-common/features/canvas/components/kpi";
import { validateMeasures } from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewTimeSeries,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function useKPITotals(
  ctx: StateManagers,
  kpiSpec: KPISpec,
  enabled: boolean,
): CreateQueryResult<number | null, HTTPError> {
  const { canvasEntity } = ctx;

  const {
    metrics_view: metricsViewName,
    measure,
    time_range: componentTimeRange,
    dimension_filters: componentFilter,
  } = kpiSpec;

  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentTimeRange,
      componentFilter,
    },
  );

  return derived(
    [ctx.runtime, timeAndFilterStore],
    ([$runtime, { timeRange, where }], set) => {
      return createQueryServiceMetricsViewAggregation(
        $runtime.instanceId,
        metricsViewName,
        {
          measures: [{ name: measure }],
          timeRange,
          where,
        },
        {
          query: {
            enabled:
              enabled &&
              (!!componentTimeRange ||
                (!!timeRange?.start && !!timeRange?.end)),
            select: (data) => {
              return data.data?.[0]?.[measure] ?? null;
            },
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}

export function useKPIComparisonTotal(
  ctx: StateManagers,
  kpiSpec: KPISpec,
  enabled: boolean,
): CreateQueryResult<number | null, HTTPError> {
  const { canvasEntity } = ctx;
  const { showTimeComparison } = canvasEntity.timeControls;

  const {
    metrics_view: metricsViewName,
    measure,
    comparison_range: componentComparisonRange,
    dimension_filters: componentFilter,
  } = kpiSpec;

  // Build the store that yields { finalTimeRange, where }
  const timeAndFilterStore = canvasEntity.createTimeAndFilterStore(
    metricsViewName,
    {
      componentComparisonRange,
      componentFilter,
    },
  );

  return derived(
    [ctx.runtime, timeAndFilterStore, showTimeComparison],
    ([$runtime, { comparisonRange, where }, showComparison], set) => {
      // TODO: Use all time range and then calculate the comparison range

      return createQueryServiceMetricsViewAggregation(
        $runtime.instanceId,
        metricsViewName,
        {
          measures: [{ name: measure }],
          timeRange: comparisonRange,
          where,
        },
        {
          query: {
            enabled:
              enabled &&
              (!!componentComparisonRange ||
                (showComparison &&
                  !!comparisonRange?.start &&
                  !!comparisonRange?.end)),
            select: (data) => {
              return data.data?.[0]?.[measure] ?? null;
            },
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}

export function useKPISparkline(
  ctx: StateManagers,
  kpiSpec: KPISpec,
  enabled: boolean,
): CreateQueryResult<Array<Record<string, unknown>>> {
  const {
    canvasEntity: { createTimeAndFilterStore },
  } = ctx;

  const {
    metrics_view: metricsViewName,
    measure,
    // TODO: Override time range for sparkline when we integrate
    // time range panel with super pill and in turn can get
    // start and end time from the super pill instead of ISO

    // time_range: componentTimeRange,
    dimension_filters: componentFilter,
  } = kpiSpec;

  const timeAndFilterStore = createTimeAndFilterStore(metricsViewName, {
    componentFilter,
  });

  return derived(
    [ctx.runtime, timeAndFilterStore],
    ([$runtime, { timeRange, where, timeGrain }], set) => {
      const { start, end, timeZone } = timeRange;

      const defaultGrain = timeGrain || V1TimeGrain.TIME_GRAIN_HOUR;

      return createQueryServiceMetricsViewTimeSeries(
        $runtime.instanceId,
        metricsViewName,
        {
          measureNames: [measure],
          timeStart: start,
          timeEnd: end,
          timeGranularity: defaultGrain,
          timeZone,
          where,
        },
        {
          query: {
            enabled: !!start && !!end && enabled,
            select: (data) => {
              return prepareTimeSeries(
                data.data || [],
                [],
                TIME_GRAIN[defaultGrain]?.duration,
                timeZone ?? "UTC",
              );
            },
            queryClient: ctx.queryClient,
          },
        },
      ).subscribe(set);
    },
  );
}

export function validateKPISchema(
  ctx: StateManagers,
  kpiSpec: KPISpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view } = kpiSpec;
  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      const measure = kpiSpec.measure;
      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }
      const validateMeasuresRes = validateMeasures(metricsView, [measure]);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure "${invalidMeasures}" selected`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
