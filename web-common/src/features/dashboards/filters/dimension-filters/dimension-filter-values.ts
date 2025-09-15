import { getCompoundQuery } from "@rilldata/web-common/features/compound-query-result";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import {
  createInExpression,
  createLikeExpression,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewAggregation,
  V1BuiltinMeasure,
} from "@rilldata/web-common/runtime-client";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

type DimensionSearchArgs = {
  mode: DimensionFilterMode;
  searchText: string;
  values: string[];
  timeStart?: string;
  timeEnd?: string;
  enabled?: boolean;
  additionalFilter?: V1Expression;
};
/**
 * Returns the search results from the search input in a dimension filter.
 *
 * 1. For Select and Contains mode, it returns the result from the search text using a `like` filter.
 * 2. For InList mode, it returns values from selection that is actually in the data source.
 *
 * Uses a "below the fold" strategy to ensure selected values always appear in results,
 * even if they're not in the top 250.
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
    additionalFilter,
  }: DimensionSearchArgs,
) {
  const where = getFilterForSearchArgs(dimensionName, {
    mode,
    searchText,
    values,
    additionalFilter,
  });

  // Main query: Get top 250 results (above the fold)
  const mainQueries = metricsViewNames.map((mvName) =>
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
      queryClient,
    ),
  );

  return getCompoundQuery(mainQueries, (responses) => {
    // Get main results (above the fold)
    const mainValues = responses
      .filter((r) => !!r?.data)
      .map((r) => r!.data!.map((i) => i[dimensionName]))
      .flat();

    // For Select and InList modes, ensure selected values are included
    // This is the "below the fold" behavior - no query needed, just merge the values
    const shouldIncludeSelectedValues =
      (mode === DimensionFilterMode.InList && values.length > 0) ||
      (mode === DimensionFilterMode.Select && values.length > 0);

    if (shouldIncludeSelectedValues) {
      // Merge results: main results first, then any selected values not already included
      const mainSet = new Set(mainValues);
      const actualSelectedValues = values.filter(
        (value) => !mainSet.has(value),
      );
      const combinedValues = [...mainValues, ...actualSelectedValues];

      const dedupedValues = new Set(combinedValues);
      return [...dedupedValues] as string[];
    }

    // For Contains mode or when no selected values, just return main results
    const dedupedValues = new Set(mainValues);
    return [...dedupedValues] as string[];
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
    additionalFilter,
  }: DimensionSearchArgs,
) {
  const where = getFilterForSearchArgs(dimensionName, {
    mode,
    searchText,
    values,
    additionalFilter,
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
      queryClient,
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
  { mode, searchText, values, additionalFilter }: DimensionSearchArgs,
) {
  let filter;
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
