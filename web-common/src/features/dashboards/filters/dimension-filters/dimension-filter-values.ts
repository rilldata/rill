import { getCompoundQuery } from "@rilldata/web-common/features/compound-query-result";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-mode";
import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  V1BuiltinMeasure,
} from "@rilldata/web-common/runtime-client";

type DimensionSearchArgs = {
  mode: DimensionFilterMode;
  searchText: string;
  values: string[];
  timeStart?: string;
  timeEnd?: string;
  enabled?: boolean;
};
/**
 * Returns the search results from the search input in a dimension filter.
 *
 * 1. For Select and Contains mode, it returns the result from the search text using a `like` filter.
 * 2. For InList mode, it returns values from selection that is actually in the data source.
 */
export function useDimensionSearch(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  {
    mode,
    searchText,
    values,
    timeStart,
    timeEnd,
    enabled,
  }: DimensionSearchArgs,
) {
  const where = getFilterForSearchArgs(dimensionName, {
    mode,
    searchText,
    values,
  });

  const queries = metricsViewNames.map((mvName) =>
    createQueryServiceMetricsViewAggregation(
      instanceId,
      mvName,
      {
        dimensions: [{ name: dimensionName }],
        timeRange: { start: timeStart, end: timeEnd },
        limit: "250",
        offset: "0",
        sort: [{ name: dimensionName }],
        where,
      },
      {
        query: { enabled },
      },
    ),
  );

  return getCompoundQuery(queries, (responses) => {
    const values = responses
      .filter((r) => !!r?.data)
      .map((r) => r!.data!.map((i) => i[dimensionName] as string))
      .flat();
    const dedupedValues = new Set(values);
    return [...dedupedValues];
  });
}

/**
 * Returns the matched search results count.
 *
 * 1. For Select this will be disabled.
 * 2. For InList mode, it returns the count of values actually present in the data source.
 * 3. For Contains mode, it returns the count of values matching the search text.
 */
export function useAllSearchResultsCount(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  {
    mode,
    searchText,
    values,
    timeStart,
    timeEnd,
    enabled,
  }: DimensionSearchArgs,
) {
  const where = getFilterForSearchArgs(dimensionName, {
    mode,
    searchText,
    values,
  });

  const queries = metricsViewNames.map((mvName) =>
    createQueryServiceMetricsViewAggregation(
      instanceId,
      mvName,
      {
        measures: [
          {
            name: dimensionName + "__distinct_count",
            builtinMeasure: V1BuiltinMeasure.BUILTIN_MEASURE_COUNT_DISTINCT,
            builtinMeasureArgs: [dimensionName],
          },
        ],
        timeRange: { start: timeStart, end: timeEnd },
        limit: "250",
        offset: "0",
        where,
      },
      {
        query: { enabled },
      },
    ),
  );

  return getCompoundQuery(queries, (responses) => {
    if (!enabled) return undefined;

    const values = responses
      .filter((r) => !!r?.data)
      .map((r) =>
        r!.data!.map((i) => i[dimensionName + "__distinct_count"] as number),
      )
      .flat();
    return values.reduce((s, v) => s + v, 0);
  });
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
  { mode, searchText, values }: DimensionSearchArgs,
) {
  if (mode === DimensionFilterMode.InList) {
    return createInExpression(dimensionName, values);
  }

  const addNull = searchText.length !== 0 && "null".includes(searchText);
  return addNull
    ? createInExpression(dimensionName, [null])
    : createLikeExpression(dimensionName, `%${searchText}%`);
}
