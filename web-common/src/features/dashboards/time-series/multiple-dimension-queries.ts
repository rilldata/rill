// Query for fetching timeseries data for individual dimension values
// TODO: Replace this with MetricsViewAggregationRequest API call
import { derived, writable } from "svelte/store";

import {
  createQueryServiceMetricsViewTimeSeries,
  V1MetricsViewFilter,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { prepareTimeSeries } from "./utils";
import {
  CHECKMARK_COLORS,
  LINE_COLORS,
} from "@rilldata/web-common/features/dashboards/config";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";

/***
 * Create a dervied svelte store that fetches the
 * timeseries data for a given dimension value
 *  individually for a given set of dimension values
 */

export function getDimensionValueTimeSeries(
  values: string[],
  instanceId: string,
  metricViewName: string,
  dimensionName: string,
  selectedMeasureNames: string[],
  filters: V1MetricsViewFilter,
  start: string,
  end: string,
  interval: V1TimeGrain,
  zone: string
) {
  if (!values && values.length == 0) return;
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
          createQueryServiceMetricsViewTimeSeries(instanceId, metricViewName, {
            measureNames: selectedMeasureNames,
            filter: updatedFilter,
            timeStart: start,
            timeEnd: end,
            timeGranularity: interval,
            timeZone: zone,
          }),
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
  );
}
