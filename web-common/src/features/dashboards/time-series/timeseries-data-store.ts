import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { derived, type Readable } from "svelte/store";
import {
  V1MetricsViewTimeSeriesResponse,
  createQueryServiceMetricsViewTimeSeries,
  createQueryServiceMetricsViewToplist,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { prepareTimeSeries } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";

export type TimeSeriesDataState = {
  isFetching: boolean;

  // Computed prepared data for charts and table
  timeSeriesData?: unknown[];
  dimensionData?: unknown[];
};

export type TimeSeriesDataStore = Readable<TimeSeriesDataState>;

// TODO: Colacate with leaderboard and other toplist store
function createMetricsTopList(
  ctx: StateManagers,
  dimensionName: string,
  measures,
  filters
) {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, name, dashboardStore, timeControls], set) => {
      createQueryServiceMetricsViewToplist(
        runtime.instanceId,
        name,
        {
          dimensionName: dimensionName,
          measureNames: measures,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          filter: filters,
          limit: "250",
          offset: "0",
          sort: [
            {
              name: dashboardStore.leaderboardMeasureName,
              ascending:
                dashboardStore.sortDirection === SortDirection.ASCENDING,
            },
          ],
        },
        {
          query: {
            enabled: timeControls.ready && !!filters,
          },
        }
      ).subscribe(set);
    }
  );
}

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
          },
        }
      ).subscribe(set)
  );
}

// function getDimensionDataQuery() {
//   let includedValues;
//   let allDimQuery;

//   if (comparisonDimension && $timeControlsStore.ready) {
//     const dimensionFilters = $dashboardStore.filters.include.filter(
//       (filter) => filter.name === comparisonDimension
//     );
//     if (dimensionFilters) {
//       includedValues = dimensionFilters[0]?.in.slice(0, 7) || [];
//     }

//     if (includedValues.length === 0) {
//       // TODO: Create a central store for topList
//       // Fetch top values for the dimension
//       const filterForDimension = getFilterForDimension(
//         $dashboardStore?.filters,
//         comparisonDimension
//       );
//       topListQuery = createQueryServiceMetricsViewToplist(
//         $runtime.instanceId,
//         metricViewName,
//         {
//           dimensionName: comparisonDimension,
//           measureNames: [$dashboardStore?.leaderboardMeasureName],
//           timeStart: $timeControlsStore.timeStart,
//           timeEnd: $timeControlsStore.timeEnd,
//           filter: filterForDimension,
//           limit: "250",
//           offset: "0",
//           sort: [
//             {
//               name: $dashboardStore?.leaderboardMeasureName,
//               ascending:
//                 $dashboardStore.sortDirection === SortDirection.ASCENDING,
//             },
//           ],
//         },
//         {
//           query: {
//             enabled: $timeControlsStore.ready && !!filterForDimension,
//           },
//         }
//       );
//     }
//   }
// }

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
      let comparisonTimeSeries: CreateQueryResult<
        V1MetricsViewTimeSeriesResponse,
        unknown
      >;
      if (showComparison) {
        comparisonTimeSeries = createMetricsViewTimeSeries(ctx, measures, true);
      }

      return derived(
        [primaryTimeSeries, comparisonTimeSeries],
        ([primary, comparison]) => {
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
            isFetching: false,
            timeSeriesData,
            dimensionData: [],
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
