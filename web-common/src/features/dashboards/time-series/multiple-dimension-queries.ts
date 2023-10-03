import { Readable, derived, writable } from "svelte/store";

import {
  V1MetricsViewFilter,
  createQueryServiceMetricsViewTimeSeries,
  createQueryServiceMetricsViewToplist,
} from "@rilldata/web-common/runtime-client";
import { getFilterForComparedDimension, prepareTimeSeries } from "./utils";
import {
  CHECKMARK_COLORS,
  LINE_COLORS,
} from "@rilldata/web-common/features/dashboards/config";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getFilterForDimension } from "@rilldata/web-common/features/dashboards/selectors";

/***
 * Returns a list of dimension values which for which to fetch
 * timeseries data for a given dimension.
 *
 * For Overview Page -
 * Use the included values if present,
 * otherwise fetch the top values for the dimension
 *
 * For Time Dimension Detail Page -
 * Fetch all the top n values for the dimension
 */

export function getDimensionValuesForComparison(
  ctx: StateManagers,
  measures,
  surface: "chart" | "table"
): Readable<{
  values: string[];
  filter: V1MetricsViewFilter;
}> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, name, dashboardStore, timeControls], set) => {
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const isInTimeDimensionView = dashboardStore?.expandedMeasureName;

      let includedValues = [];
      const dimensionFilters = dashboardStore?.filters?.include?.filter(
        (filter) => filter.name === dimensionName
      );
      if (surface === "chart" && dimensionFilters?.length) {
        // For TDD view max 11 allowed, for overview max 7 allowed
        includedValues =
          dimensionFilters[0]?.in.slice(0, isInTimeDimensionView ? 11 : 7) ||
          [];
      }

      if (includedValues.length && surface === "chart") {
        return derived(
          [writable(includedValues), writable(dashboardStore?.filters)],
          ([values, filter]) => {
            return {
              values,
              filter,
            };
          }
        ).subscribe(set);
      } else {
        return derived(
          createQueryServiceMetricsViewToplist(
            runtime.instanceId,
            name,
            {
              dimensionName: dimensionName,
              measureNames: measures,
              timeStart: timeControls.timeStart,
              timeEnd: timeControls.timeEnd,
              filter: getFilterForDimension(
                dashboardStore?.filters,
                dimensionName
              ),
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
                enabled:
                  timeControls.ready &&
                  !!dashboardStore?.selectedComparisonDimension,
                queryClient: ctx.queryClient,
              },
            }
          ),
          (topListData) => {
            if (topListData?.isFetching)
              return {
                values: [],
                filter: dashboardStore?.filters,
              };
            const columnName = topListData?.data?.meta[0]?.name;
            const topListValues = topListData?.data?.data.map(
              (d) => d[columnName]
            );

            const computedFilter = getFilterForComparedDimension(
              dimensionName,
              dashboardStore?.filters,
              topListValues,
              surface === "table" ? 250 : 3
            );

            return {
              values: computedFilter?.includedValues,
              filter: computedFilter?.updatedFilter,
            };
          }
        ).subscribe(set);
      }
    }
  );
}

/***
 * Fetches the timeseries data for a given dimension
 * for a infered set of dimension values and measures
 */
export function getDimensionValueTimeSeries(
  ctx: StateManagers,
  measures: string[],
  surface: "chart" | "table"
) {
  // if (!values && values.length == 0) return;

  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      getDimensionValuesForComparison(ctx, measures, surface),
    ],
    (
      [runtime, metricViewName, dashboardStore, timeStore, dimensionValues],
      set
    ) => {
      const dimensionName = dashboardStore?.selectedComparisonDimension;

      const start = timeStore?.adjustedStart;
      const end = timeStore?.adjustedEnd;
      const interval =
        timeStore?.selectedTimeRange?.interval ?? timeStore?.minTimeGrain;
      const zone = dashboardStore?.selectedTimezone;

      if (!dimensionName) return;

      return derived(
        dimensionValues?.values.map((value, i) => {
          const updatedIncludeFilter = dimensionValues?.filter.include.map(
            (filter) => {
              if (filter.name === dimensionName)
                return { name: dimensionName, in: [value] };
              else return filter;
            }
          );
          // remove excluded values
          const updatedExcludeFilter = dimensionValues?.filter.exclude.filter(
            (filter) => filter.name !== dimensionName
          );
          const updatedFilter = {
            exclude: updatedExcludeFilter,
            include: updatedIncludeFilter,
          };

          return derived(
            [
              writable(value),
              createQueryServiceMetricsViewTimeSeries(
                runtime.instanceId,
                metricViewName,
                {
                  measureNames: measures,
                  filter: updatedFilter,
                  timeStart: start,
                  timeEnd: end,
                  timeGranularity: interval,
                  timeZone: zone,
                },
                {
                  query: {
                    enabled: !!timeStore.ready && !!ctx.dashboardStore,
                    queryClient: ctx.queryClient,
                  },
                }
              ),
            ],
            ([value, timeseries]) => {
              let prepData = timeseries?.data?.data;
              if (!timeseries?.isFetching) {
                prepData = prepareTimeSeries(
                  timeseries?.data?.data,
                  undefined,
                  TIME_GRAIN[interval].duration,
                  zone
                );
              }
              return {
                value,
                strokeClass: "stroke-" + LINE_COLORS[i],
                fillClass: "fill-" + CHECKMARK_COLORS[i],
                data: prepData,
                isFetching: timeseries.isFetching,
              };
            }
          );
        }),

        (combos) => {
          return combos;
        }
      ).subscribe(set);
    }
  );
}
