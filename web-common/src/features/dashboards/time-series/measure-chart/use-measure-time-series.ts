import { derived, type Readable } from "svelte/store";
import {
  createQueryServiceMetricsViewTimeSeries,
  type V1MetricsViewTimeSeriesResponse,
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
import { isGrainAllowed } from "@rilldata/web-common/lib/time/new-grains";
import type { TimeSeriesPoint } from "./types";
import { DateTime } from "luxon";

/**
 * Create a time series query for a single measure.
 * Each MeasureChart component creates its own query, enabling:
 * - Lazy loading (only fetch when visible)
 * - Per-measure error handling
 * - Independent loading states
 */
export function useMeasureTimeSeries(
  ctx: StateManagers,
  measureName: string,
  visible: Readable<boolean>,
  isComparison = false,
): CreateQueryResult<V1MetricsViewTimeSeriesResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      visible,
    ],
    (
      [runtime, metricsViewName, dashboardStore, timeControls, isVisible],
      set,
    ) => {
      const timeGrain = isGrainAllowed(
        timeControls.selectedTimeRange?.interval,
        timeControls.minTimeGrain,
      )
        ? timeControls.selectedTimeRange?.interval
        : timeControls.minTimeGrain;

      return createQueryServiceMetricsViewTimeSeries(
        runtime.instanceId,
        metricsViewName,
        {
          measureNames: [measureName], // Single measure!
          where: sanitiseExpression(
            mergeDimensionAndMeasureFilters(
              dashboardStore.whereFilter,
              dashboardStore.dimensionThresholdFilters,
            ),
            undefined,
          ),
          timeStart: isComparison
            ? timeControls.comparisonAdjustedStart
            : timeControls.adjustedStart,
          timeEnd: isComparison
            ? timeControls.comparisonAdjustedEnd
            : timeControls.adjustedEnd,
          timeGranularity: timeGrain,
          timeZone: dashboardStore.selectedTimezone,
        },
        {
          query: {
            // Only fetch when visible AND time controls are ready
            enabled:
              isVisible &&
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              (!isComparison || !!timeControls.comparisonAdjustedStart),
            placeholderData: keepPreviousData,
            refetchOnMount: false,
          },
        },
        ctx.queryClient,
      ).subscribe(set);
    },
  );
}

/**
 * Transform raw API time series data to typed TimeSeriesPoint[].
 * Minimal processing: just extract ts, value, and comparison fields.
 * No intermediate position computation — rendering uses indices directly.
 */
export function transformTimeSeriesData(
  primary: V1MetricsViewTimeSeriesResponse["data"],
  comparison: V1MetricsViewTimeSeriesResponse["data"] | undefined,
  measureName: string,
  timezone: string,
): TimeSeriesPoint[] {
  if (!primary) return [];

  return primary.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    if (!originalPt?.ts) {
      return { ts: DateTime.invalid("Invalid timestamp"), value: null };
    }

    const ts = DateTime.fromISO(originalPt.ts, { zone: timezone });

    if (!ts || typeof ts === "string") {
      return { ts: DateTime.invalid("Invalid timestamp"), value: null };
    }

    const value = (originalPt.records?.[measureName] as number | null) ?? null;

    let comparisonValue: number | null | undefined = undefined;
    let comparisonTs: DateTime | undefined = undefined;

    if (comparisonPt?.ts) {
      comparisonValue =
        (comparisonPt.records?.[measureName] as number | null) ?? null;
      comparisonTs = DateTime.fromISO(comparisonPt.ts, { zone: timezone });
    }

    return { ts, value, comparisonValue, comparisonTs };
  });
}

/**
 * Create a derived store that transforms raw query data to TimeSeriesPoint[].
 */
export function useMeasureTimeSeriesData(
  ctx: StateManagers,
  measureName: string,
  visible: Readable<boolean>,
  showComparison: Readable<boolean>,
): Readable<{
  data: TimeSeriesPoint[];
  isFetching: boolean;
  isError: boolean;
  error: string | undefined;
}> {
  const primaryQuery = useMeasureTimeSeries(ctx, measureName, visible, false);

  return derived(
    [primaryQuery, showComparison, ctx.dashboardStore],
    ([$primary, _$showComparison, $dashboard]) => {
      if ($primary.isFetching || !$primary.data?.data) {
        return {
          data: [],
          isFetching: $primary.isFetching,
          isError: $primary.isError,
          error: ($primary.error as HTTPError)?.response?.data?.message,
        };
      }

      const data = transformTimeSeriesData(
        $primary.data.data,
        undefined,
        measureName,
        $dashboard.selectedTimezone,
      );

      return {
        data,
        isFetching: false,
        isError: $primary.isError,
        error: ($primary.error as HTTPError)?.response?.data?.message,
      };
    },
  );
}
