import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/ExpressionFilterManager.svelte.ts";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  V1BuiltinMeasure,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { createReactiveQueries } from "@rilldata/web-common/lib/svelte-query/reactive-queries.svelte.ts";

type DimensionSearchArgs = {
  manager: ExpressionFilterManager;
  dimensionName: string;
  mode: DimensionFilterMode;
  searchText: string;
  values: string[];
  timeStart?: string;
  timeEnd?: string;
  timeDimension?: string;
  enabled?: boolean;
};

/**
 * Returns the search results from the search input in a dimension filter.
 *
 * A dimension can be defined by more than one metrics view, so this queries each of them and merges
 * the values. `getArgs` is read reactively, as are the specs, so the queries follow both the input
 * in the dropdown and the metrics views as they load.
 *
 * Must be called during component init, once per dimension filter.
 */
export function createDimensionSearchQuery(
  client: RuntimeClient,
  getArgs: () => DimensionSearchArgs,
) {
  return createReactiveQueries(
    () => {
      const {
        manager,
        dimensionName,
        mode,
        searchText,
        values,
        timeStart,
        timeEnd,
        timeDimension,
        enabled,
      } = getArgs();

      return getMetricsViewsForDimension(manager, dimensionName).map(
        (metricsView) =>
          getQueryServiceMetricsViewAggregationQueryOptions(
            client,
            {
              metricsView,
              dimensions: [{ name: dimensionName }],
              timeRange: { start: timeStart, end: timeEnd, timeDimension },
              limit: "250",
              offset: "0",
              sort: [{ name: dimensionName }],
              where: getFilterForSearchArgs(dimensionName, {
                mode,
                searchText,
                values,
                additionalFilter: manager.getOtherDimensionsFilter(
                  dimensionName,
                  metricsView,
                ),
              }),
            },
            {
              query: {
                enabled,
                select: (resp) =>
                  resp.data?.map((d) => d[dimensionName] as string) ?? [],
              },
            },
          ),
      );
    },
    (valuesPerMetricsView) => [
      ...new Set(valuesPerMetricsView.flatMap((values) => values ?? [])),
    ],
  );
}

/**
 * Returns the matched search results count.
 *
 * 1. For Select this will be disabled.
 * 2. For InList mode, it returns the count of values actually present in the data source.
 * 3. For Contains mode, it returns the count of values matching the search text.
 *
 * Must be called during component init, once per dimension filter.
 */
export function createDimensionSearchCountQuery(
  client: RuntimeClient,
  getArgs: () => DimensionSearchArgs,
) {
  return createReactiveQueries(
    () => {
      const {
        manager,
        dimensionName,
        mode,
        searchText,
        values,
        timeStart,
        timeEnd,
        timeDimension,
        enabled,
      } = getArgs();
      const countMeasureName = dimensionName + "__distinct_count";

      return getMetricsViewsForDimension(manager, dimensionName).map(
        (metricsView) =>
          getQueryServiceMetricsViewAggregationQueryOptions(
            client,
            {
              metricsView,
              measures: [
                {
                  name: countMeasureName,
                  builtinMeasure:
                    V1BuiltinMeasure.BUILTIN_MEASURE_COUNT_DISTINCT,
                  builtinMeasureArgs: [dimensionName],
                },
              ],
              timeRange: { start: timeStart, end: timeEnd, timeDimension },
              where: getFilterForSearchArgs(dimensionName, {
                mode,
                searchText,
                values,
                additionalFilter: manager.getOtherDimensionsFilter(
                  dimensionName,
                  metricsView,
                ),
              }),
            },
            {
              query: {
                enabled,
                select: (resp) =>
                  resp.data?.length
                    ? (resp.data[0][countMeasureName] as number)
                    : 0,
              },
            },
          ),
      );
    },
    // Absent while the queries are loading or disabled, so that the chip does not show a count yet.
    (countPerMetricsView) =>
      countPerMetricsView.some((count) => count !== undefined)
        ? countPerMetricsView.reduce(
            (total, count) => (total ?? 0) + (count ?? 0),
            0,
          )
        : undefined,
  );
}

/**
 * Metrics views that define the dimension. A dimension filter only applies to those, and querying a
 * metrics view without the dimension errors.
 */
function getMetricsViewsForDimension(
  manager: ExpressionFilterManager,
  dimensionName: string,
) {
  return Object.keys(
    manager.metricsViewsProvider.dimensionSpecs[dimensionName] ?? {},
  );
}

/**
 * Builds the filter for dimension search results or dimension search results count.
 * Note the difference, this is for the search results from the search input.
 *
 * 1. For Select mode, while the final query is an `in` filter, the search results from the search input is a `like` filter.
 * 2. For InList mode it is an `in` filter with all the selected values.
 * 3. For Contains mode it is a `like` filter.
 */
function getFilterForSearchArgs(
  dimensionName: string,
  {
    mode,
    searchText,
    values,
    additionalFilter,
  }: {
    mode: DimensionFilterMode;
    searchText: string;
    values: string[];
    additionalFilter?: V1Expression;
  },
) {
  let filter: V1Expression;
  if (mode === DimensionFilterMode.InList) {
    filter = createInExpression(dimensionName, values);
  } else {
    const addNull = searchText.length !== 0 && "null".includes(searchText);
    filter = addNull
      ? createInExpression(dimensionName, [null])
      : createLikeExpression(dimensionName, `%${searchText}%`);
  }

  if (additionalFilter) {
    return createAndExpression([filter, additionalFilter]);
  }
  return filter;
}
