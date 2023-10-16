import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { derived, type Readable } from "svelte/store";
import {
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationResponseDataItem,
  V1MetricsViewTimeSeriesResponse,
  createQueryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  DimensionDataItem,
  getDimensionValueTimeSeries,
} from "./multiple-dimension-queries";
import { createTotalsForMeasure } from "@rilldata/web-common/features/dashboards/time-series/totals-data-store";

export type TimeSeriesDataState = {
  isFetching: boolean;
  hasError: boolean;

  // Computed prepared data for charts and table
  timeSeriesData?: unknown[];
  total: V1MetricsViewAggregationResponseDataItem;
  unfilteredTotal: V1MetricsViewAggregationResponseDataItem;
  comparisonTotal: V1MetricsViewAggregationResponseDataItem;
  dimensionChartData?: DimensionDataItem[];
  dimensionTableData?: DimensionDataItem[];
};

export type TimeSeriesDataStore = Readable<TimeSeriesDataState>;

function createMetricsViewTimeSeries(
  ctx: StateManagers,
  measures,
  isComparison = false
): CreateQueryResult<V1MetricsViewTimeSeriesResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricViewName, dashboardStore, timeControls], set) =>
      createQueryServiceMetricsViewTimeSeries(
        runtime.instanceId,
        metricViewName,
        {
          measureNames: measures,
          filter: dashboardStore?.filters,
          timeStart: isComparison
            ? timeControls.comparisonAdjustedStart
            : timeControls.adjustedStart,
          timeEnd: isComparison
            ? timeControls.comparisonAdjustedEnd
            : timeControls.adjustedEnd,
          timeGranularity:
            timeControls.selectedTimeRange?.interval ??
            timeControls.minTimeGrain,
          timeZone: dashboardStore?.selectedTimezone,
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        }
      ).subscribe(set)
  );
}

export function createTimeSeriesDataStore(ctx: StateManagers) {
  return derived(
    [useMetaQuery(ctx), useTimeControlStore(ctx), ctx.dashboardStore],
    ([metricsView, timeControls, dashboardStore], set) => {
      const showComparison = timeControls.showComparison;
      const interval =
        timeControls.selectedTimeRange?.interval ?? timeControls.minTimeGrain;

      const allMeasures = metricsView.data?.measures.map(
        (measure) => measure.name
      );
      let measures = allMeasures;
      if (dashboardStore?.expandedMeasureName) {
        measures = allMeasures.filter(
          (measure) => measure === dashboardStore.expandedMeasureName
        );
      } else {
        measures = dashboardStore?.selectedMeasureNames;
      }

      const primaryTimeSeries = createMetricsViewTimeSeries(
        ctx,
        measures,
        false
      );
      const primaryTotals = createTotalsForMeasure(ctx, measures, false);

      const unfilteredTotals = createTotalsForMeasure(
        ctx,
        measures,
        false,
        true
      );

      let comparisonTimeSeries: CreateQueryResult<
        V1MetricsViewTimeSeriesResponse,
        unknown
      >;
      let comparisonTotals: CreateQueryResult<
        V1MetricsViewAggregationResponse,
        unknown
      >;
      if (showComparison) {
        comparisonTimeSeries = createMetricsViewTimeSeries(ctx, measures, true);
        comparisonTotals = createTotalsForMeasure(ctx, measures, true);
      }

      let dimensionTimeSeriesCharts;
      let dimensionTimeSeriesTable;
      if (dashboardStore?.selectedComparisonDimension) {
        dimensionTimeSeriesCharts = getDimensionValueTimeSeries(
          ctx,
          measures,
          "chart"
        );

        // Fetch table data only if in TDD view
        if (dashboardStore?.expandedMeasureName) {
          dimensionTimeSeriesTable = getDimensionValueTimeSeries(
            ctx,
            measures,
            "table"
          );
        }
      }

      return derived(
        [
          primaryTimeSeries,
          comparisonTimeSeries,
          primaryTotals,
          unfilteredTotals,
          comparisonTotals,
          dimensionTimeSeriesCharts,
          dimensionTimeSeriesTable,
        ],
        ([
          primary,
          comparison,
          primaryTotal,
          unfilteredTotal,
          comparisonTotal,
          dimensionChart,
          dimensionTable,
        ]) => {
          let timeSeriesData = primary?.data?.data;

          if (!primary.isFetching) {
            timeSeriesData = prepareTimeSeries(
              primary?.data?.data,
              comparison?.data?.data,
              TIME_GRAIN[interval].duration,
              dashboardStore.selectedTimezone
            );
          }
          return {
            isFetching: primaryTotal?.isFetching || comparisonTotal?.isFetching, // FIXME Handle fetching
            hasError: false, // FIXME Handle errors
            timeSeriesData,
            total: primaryTotal?.data?.data[0],
            unfilteredTotal: unfilteredTotal?.data?.data[0],
            comparisonTotal: comparisonTotal?.data?.data[0],
            dimensionChartData: (dimensionChart as DimensionDataItem[]) || [],
            dimensionTableData: (dimensionTable as DimensionDataItem[]) || [],
          };
        }
      ).subscribe(set);
    }
  ) as TimeSeriesDataStore;
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeSeriesDataStore = memoizeMetricsStore<TimeSeriesDataStore>(
  (ctx: StateManagers) => createTimeSeriesDataStore(ctx)
);
