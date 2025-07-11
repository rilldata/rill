import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { filterOutSomeAdvancedMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createTotalsForMeasure,
  createUnfilteredTotalsForMeasure,
} from "@rilldata/web-common/features/dashboards/time-series/totals-data-store";
import { prepareTimeSeriesOffsets } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { Period } from "@rilldata/web-common/lib/time/types";
import {
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  getQueryServiceMetricsViewAggregationQueryOptions,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  type CreateQueryResult,
  createQuery,
  keepPreviousData,
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
  ts_position?: Date;
  "comparison.ts"?: Date;
  "comparison.ts_position"?: Date;
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
  dimensionChartData?: DimensionDataItem[];
};

export type TimeSeriesDataStore = Readable<TimeSeriesDataState>;

export const ComparisonTimeSuffix = "_ts_comparison";

export function createMetricsViewTimeSeriesFromAggregation(
  ctx: StateManagers,
  measureNames: string[],
  includeTimeComparison = false,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const queryOptionsStore = derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricsViewName, dashboardStore, timeControls]) => {
      const timeGrain =
        timeControls?.selectedTimeRange?.interval ?? timeControls.minTimeGrain;
      const timeZone = dashboardStore?.selectedTimezone;
      const timeDimension = timeControls?.timeDimension;

      let measures: V1MetricsViewAggregationMeasure[] = measureNames.flatMap(
        (measureName) => {
          const baseMeasure = { name: measureName };
          if (!includeTimeComparison) {
            return [baseMeasure];
          }
          return [
            baseMeasure,
            {
              name: measureName + ComparisonDeltaPreviousSuffix,
              comparisonValue: { measure: measureName },
            },
          ];
        },
      );

      if (includeTimeComparison) {
        measures = [
          ...measures,
          {
            name: timeDimension + ComparisonTimeSuffix,
            comparisonTime: { dimension: timeDimension },
          },
        ];
      }

      const enabled =
        !!timeControls.ready &&
        !!ctx.dashboardStore &&
        !!timeDimension &&
        !!timeGrain &&
        // in case of comparison, we need to wait for the comparison start time to be available
        (!includeTimeComparison || !!timeControls.comparisonAdjustedStart);

      return getQueryServiceMetricsViewAggregationQueryOptions(
        runtime.instanceId,
        metricsViewName,
        {
          measures: measures,
          dimensions: [{ name: timeDimension, timeGrain, timeZone }],
          where: sanitiseExpression(
            mergeDimensionAndMeasureFilters(
              dashboardStore.whereFilter,
              dashboardStore.dimensionThresholdFilters,
            ),
            undefined,
          ),
          sort: [{ name: timeDimension, desc: false }],
          fillMissing: true,
          timeRange: {
            start: timeControls.adjustedStart,
            end: timeControls.adjustedEnd,
          },
          comparisonTimeRange: includeTimeComparison
            ? {
                start: timeControls.comparisonAdjustedStart,
                end: timeControls.comparisonAdjustedEnd,
              }
            : undefined,
        },
        {
          query: {
            enabled,
            placeholderData: keepPreviousData,
            refetchOnMount: false,
          },
        },
      );
    },
  );

  return createQuery(queryOptionsStore, ctx.queryClient);
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
      const timeDimension = timeControls?.timeDimension;

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

      const timeSeriesDataStore =
        measuresForTimeSeries.length > 0
          ? createMetricsViewTimeSeriesFromAggregation(
              ctx,
              measuresForTimeSeries,
              showComparison,
            )
          : writable({
              isFetching: false,
              isError: false,
              data: null,
              error: {},
            });

      const totalsDataStore =
        measuresForTotals.length > 0
          ? createTotalsForMeasure(ctx, measuresForTotals, showComparison)
          : writable({
              isFetching: false,
              isError: false,
              data: null,
              error: undefined,
            });

      let unfilteredTotals:
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError>
        | Writable<null> = writable(null);

      if (dashboardStore?.selectedComparisonDimension) {
        unfilteredTotals = createUnfilteredTotalsForMeasure(
          ctx,
          measures,
          dashboardStore?.selectedComparisonDimension,
        );
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
          timeSeriesDataStore,
          totalsDataStore,
          unfilteredTotals,
          dimensionTimeSeriesCharts,
        ],
        ([timeSeriesData, totalsData, unfilteredTotal, dimensionChart]) => {
          let preparedTimeSeriesData: TimeSeriesDatum[] = [];

          if (!timeSeriesData.isFetching && interval) {
            const intervalDuration = TIME_GRAIN[interval]?.duration as Period;
            preparedTimeSeriesData = prepareTimeSeriesOffsets(
              timeSeriesData?.data?.data || [],
              timeDimension,
              intervalDuration,
              dashboardStore.selectedTimezone,
            );
          }

          let isError = false;

          const error = {};
          if (timeSeriesData.error) {
            isError = true;
            error["timeseries"] = (
              timeSeriesData.error as HTTPError
            ).response?.data?.message;
          }
          if (totalsData.error) {
            isError = true;
            error["totals"] = totalsData.error.response?.data?.message;
          }

          return {
            isFetching: timeSeriesData.isFetching || totalsData.isFetching,
            isError,
            error,
            timeSeriesData: preparedTimeSeriesData,
            total: totalsData?.data?.data?.[0],
            unfilteredTotal: unfilteredTotal?.data?.data?.[0],
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
