import { derived, writable } from "svelte/store";

import {
  createQueryServiceMetricsViewTimeSeries,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { prepareTimeSeries } from "./utils";
import {
  CHECKMARK_COLORS,
  LINE_COLORS,
} from "@rilldata/web-common/features/dashboards/config";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getFilterForDimension } from "@rilldata/web-common/features/dashboards/selectors";

/***
 * Create a dervied svelte store that fetches the
 * timeseries data for a given dimension value
 *  individually for a given set of dimension values
 */
// TODO: Replace this with MetricsViewAggregationRequest API call
export function getDimensionValueTimeSeries(
  ctx: StateManagers,
  values: string[],
  measures: string[],
  filters: V1MetricsViewFilter
) {
  if (!values && values.length == 0) return;

  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
    ],
    ([runtime, metricViewName, dashboardStore, timeStore], set) => {
      // let values = [];
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      // const dimensionFilters = dashboardStore?.filters.include.filter(
      //   (filter) => filter.name === dimensionName
      // );
      // if (dimensionFilters) {
      //   values = dimensionFilters[0]?.in.slice(0, 7) || [];
      // }
      // if (!values?.length) {
      //   const filterForDimension = getFilterForDimension(
      //     dashboardStore?.filters,
      //     dimensionName
      //   );
      // }

      const start = timeStore?.adjustedStart;
      const end = timeStore?.adjustedEnd;
      const interval =
        timeStore?.selectedTimeRange?.interval ?? timeStore?.minTimeGrain;
      const zone = dashboardStore?.selectedTimezone;

      return derived(
        values.map((value, i) => {
          const updatedIncludeFilter = filters.include.map((filter) => {
            if (filter.name === dimensionName)
              return { name: dimensionName, in: [value] };
            else return filter;
          });
          // remove excluded values
          const updatedExcludeFilter = filters.exclude.filter(
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
