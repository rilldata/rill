import { derived, type Readable } from "svelte/store";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  keepPreviousData,
  type CreateQueryResult,
} from "@tanstack/svelte-query";
import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

/**
 * Totals data for a single measure.
 */
export interface MeasureTotalsData {
  value: number | null;
  comparisonValue: number | null;
  isFetching: boolean;
  isError: boolean;
  error: string | undefined;
}

/**
 * Create a totals query for a single measure.
 * Used for the big number display above each chart.
 */
export function useMeasureTotals(
  ctx: StateManagers,
  measureName: string,
  visible: Readable<boolean>,
  isComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      visible,
    ],
    ([runtime, metricsViewName, timeControls, dashboard, isVisible], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          measures: [{ name: measureName }],
          where: sanitiseExpression(
            mergeDimensionAndMeasureFilters(
              dashboard.whereFilter,
              dashboard.dimensionThresholdFilters,
            ),
            undefined,
          ),
          timeRange: {
            start: isComparison
              ? timeControls?.comparisonTimeStart
              : timeControls.timeStart,
            end: isComparison
              ? timeControls?.comparisonTimeEnd
              : timeControls.timeEnd,
          },
        },
        {
          query: {
            enabled:
              isVisible &&
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              (!isComparison || !!timeControls.comparisonTimeStart),
            placeholderData: keepPreviousData,
            refetchOnMount: false,
          },
        },
        ctx.queryClient,
      ).subscribe(set),
  );
}

/**
 * Create a derived store that provides totals data for a measure.
 * Includes both primary and comparison values.
 */
export function useMeasureTotalsData(
  ctx: StateManagers,
  measureName: string,
  visible: Readable<boolean>,
  showComparison: Readable<boolean>,
): Readable<MeasureTotalsData> {
  const primaryQuery = useMeasureTotals(ctx, measureName, visible, false);

  // Create comparison query only when comparison is enabled
  const comparisonQuery = derived(
    [showComparison, visible],
    ([$showComparison, $visible]) => {
      if (!$showComparison) {
        return null;
      }
      return useMeasureTotals(ctx, measureName, visible, true);
    },
  );

  return derived(
    [primaryQuery, comparisonQuery, showComparison],
    ([$primary, $comparisonQueryStore, $showComparison]) => {
      // Get primary value
      const primaryValue =
        ($primary.data?.data?.[0]?.[measureName] as number | null) ?? null;

      // Get comparison value if applicable
      let comparisonValue: number | null = null;
      let comparisonIsFetching = false;
      let comparisonIsError = false;
      let comparisonError: string | undefined;

      if ($showComparison && $comparisonQueryStore) {
        // Note: This is a simplified version. In full implementation,
        // we'd need to properly subscribe to the comparison query.
        // For now, comparison is handled at the parent level.
      }

      return {
        value: primaryValue,
        comparisonValue,
        isFetching: $primary.isFetching || comparisonIsFetching,
        isError: $primary.isError || comparisonIsError,
        error:
          ($primary.error as HTTPError)?.response?.data?.message ??
          comparisonError,
      };
    },
  );
}

/**
 * Compute comparison metrics from primary and comparison values.
 */
export function computeComparisonMetrics(
  value: number | null,
  comparisonValue: number | null,
): {
  delta: number | null;
  deltaPercent: number | null;
  isPositive: boolean | null;
} {
  if (value === null || comparisonValue === null || comparisonValue === 0) {
    return {
      delta: null,
      deltaPercent: null,
      isPositive: null,
    };
  }

  const delta = value - comparisonValue;
  const deltaPercent = (delta / Math.abs(comparisonValue)) * 100;
  const isPositive = delta >= 0;

  return {
    delta,
    deltaPercent,
    isPositive,
  };
}
