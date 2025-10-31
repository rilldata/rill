import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { filterOutSomeAdvancedMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createTotalsForMeasure,
  createUnfilteredTotalsForMeasure,
} from "@rilldata/web-common/features/dashboards/time-series/totals-data-store";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { Period } from "@rilldata/web-common/lib/time/types";
import {
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewTimeSeriesResponse,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  keepPreviousData,
  type CreateQueryResult,
} from "@tanstack/svelte-query";
import { type Readable, type Writable, derived, writable } from "svelte/store";
import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
import { memoizeMetricsStore } from "../state-managers/memoize-metrics-store";
import {
  type DimensionDataItem,
  getDimensionValueTimeSeries,
} from "./multiple-dimension-queries";

export interface TimeSeriesDatum {
  ts?: Date;
  bin?: number;
  ts_comparison?: Date;
  ts_position?: Date;
  [key: string]: Date | string | number | undefined;
}

export type TimeSeriesDataState = {
  isFetching: boolean;
  isError: boolean;
  error: { [key: string]: string | undefined };

  // Computed prepared data for charts and table
  timeSeriesData?: TimeSeriesDatum[];
  total?: V1MetricsViewAggregationResponseDataItem;
  unfilteredTotal?: V1MetricsViewAggregationResponseDataItem;
  comparisonTotal?: V1MetricsViewAggregationResponseDataItem;
  dimensionChartData?: DimensionDataItem[];
};

export type TimeSeriesDataStore = Readable<TimeSeriesDataState>;

export function createMetricsViewTimeSeries(
  ctx: StateManagers,
  measures: string[],
  isComparison = false,
): CreateQueryResult<V1MetricsViewTimeSeriesResponse, HTTPError> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricsViewName, dashboardStore, timeControls], set) => {
      return createQueryServiceMetricsViewTimeSeries(
        runtime.instanceId,
        metricsViewName,
        {
          measureNames: measures,
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
          timeGranularity:
            timeControls.selectedTimeRange?.interval ??
            timeControls.minTimeGrain,
          timeZone: dashboardStore.selectedTimezone,
        },
        {
          query: {
            enabled:
              !!timeControls.ready &&
              !!ctx.dashboardStore &&
              // in case of comparison, we need to wait for the comparison start time to be available
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

export function createTimeSeriesDataStore(
  ctx: StateManagers,
): TimeSeriesDataStore {
  return derived(
    [ctx.validSpecStore, useTimeControlStore(ctx), ctx.dashboardStore],
    ([validSpec, timeControls, dashboardStore], set) => {
      if (!validSpec.data || !timeControls.ready || timeControls.isFetching) {
        set({
          isFetching: true,
          isError: false,
          error: {},
        });
        return;
      }

      const showComparison = timeControls.showTimeComparison;
      const interval =
        timeControls.selectedTimeRange?.interval ?? timeControls.minTimeGrain;

      const { metricsView, explore } = validSpec.data;

      const allMeasures = explore?.measures ?? [];
      let measures = allMeasures;
      const showTimeDimensionDetail = Boolean(
        dashboardStore?.activePage ===
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      );
      const expandedMeasuerName = dashboardStore?.tdd?.expandedMeasureName;
      if (showTimeDimensionDetail && expandedMeasuerName) {
        measures = allMeasures.filter(
          (measure) => measure === expandedMeasuerName,
        );
      } else {
        measures = dashboardStore?.visibleMeasures
          ? [...dashboardStore.visibleMeasures]
          : [];
      }

      const measuresForTimeSeries = filterOutSomeAdvancedMeasures(
        dashboardStore,
        metricsView ?? {},
        measures,
        true,
      );
      const measuresForTotals = filterOutSomeAdvancedMeasures(
        dashboardStore,
        metricsView ?? {},
        measures,
        false,
      );

      const primaryTimeSeries =
        measuresForTimeSeries.length > 0
          ? createMetricsViewTimeSeries(ctx, measuresForTimeSeries, false)
          : writable({
              isFetching: false,
              isError: false,
              data: null,
              error: {},
            });

      const primaryTotals =
        measuresForTotals.length > 0
          ? createTotalsForMeasure(ctx, measuresForTotals, false)
          : writable({
              isFetching: false,
              isError: false,
              data: { data: [] },
              error: undefined,
            });

      let unfilteredTotals:
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError>
        | Writable<null> = writable(null);

      if (dashboardStore?.selectedComparisonDimension) {
        unfilteredTotals = createUnfilteredTotalsForMeasure(
          ctx,
          measuresForTotals,
          dashboardStore?.selectedComparisonDimension,
        );
      }
      let comparisonTimeSeries:
        | CreateQueryResult<V1MetricsViewTimeSeriesResponse, HTTPError>
        | Writable<null> = writable(null);
      let comparisonTotals:
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError>
        | Writable<null> = writable(null);
      if (showComparison) {
        comparisonTimeSeries = createMetricsViewTimeSeries(
          ctx,
          measuresForTimeSeries,
          true,
        );
        comparisonTotals = createTotalsForMeasure(ctx, measuresForTotals, true);
      }

      let dimensionTimeSeriesCharts:
        | Readable<DimensionDataItem[]>
        | Writable<null> = writable(null);
      if (dashboardStore?.selectedComparisonDimension) {
        dimensionTimeSeriesCharts = getDimensionValueTimeSeries(
          ctx,
          measuresForTimeSeries,
          "chart",
        );
      }

      return derived(
        [
          primaryTimeSeries,
          comparisonTimeSeries,
          primaryTotals,
          unfilteredTotals,
          comparisonTotals,
          dimensionTimeSeriesCharts,
        ],
        ([
          primary,
          comparison,
          primaryTotal,
          unfilteredTotal,
          comparisonTotal,
          dimensionChart,
        ]) => {
          let preparedTimeSeriesData: TimeSeriesDatum[] = [];

          if (!primary.isFetching && interval) {
            const intervalDuration = TIME_GRAIN[interval]?.duration as Period;
            preparedTimeSeriesData = prepareTimeSeries(
              primary?.data?.data || [],
              comparison?.data?.data || [],
              intervalDuration,
              dashboardStore.selectedTimezone,
            );
          }

          let isError = false;

          const error = {};
          if (primary.error) {
            isError = true;
            error["timeseries"] = (
              primary.error as HTTPError
            ).response?.data?.message;
          }
          if (primaryTotal.error) {
            isError = true;
            error["totals"] = primaryTotal.error.response?.data?.message;
          }
          const primaryIsFetching = primary.isFetching;
          const primaryTotalIsFetching = primaryTotal.isFetching;

          return {
            isFetching: primaryIsFetching || primaryTotalIsFetching,
            isError,
            error,
            timeSeriesData: preparedTimeSeriesData,
            total: primaryTotal?.data?.data?.[0],
            unfilteredTotal: unfilteredTotal?.data?.data?.[0],
            comparisonTotal: comparisonTotal?.data?.data?.[0],
            dimensionChartData: (dimensionChart as DimensionDataItem[]) || [],
          };
        },
      ).subscribe(set);
    },
  );
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeSeriesDataStore = memoizeMetricsStore<TimeSeriesDataStore>(
  (ctx: StateManagers) => createTimeSeriesDataStore(ctx),
);
