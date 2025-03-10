import {
  type CompoundQueryResult,
  getCompoundAggregationQuery,
} from "@rilldata/web-common/features/compound-query-result";
import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  V1BuiltinMeasure,
} from "@rilldata/web-common/runtime-client";

type DimensionSearchArgs = {
  searchText?: string;
  values?: string[];
  timeStart?: string;
  timeEnd?: string;
  enabled?: boolean;
};
export function useDimensionSearch(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  { searchText, values, timeStart, timeEnd, enabled }: DimensionSearchArgs,
): CompoundQueryResult<string[]> {
  const where = getFilterForSearchArgs(dimensionName, { searchText, values });

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

  return getCompoundAggregationQuery(queries, (responses) => {
    const values = responses
      .filter((r) => !!r?.data)
      .map((r) => r!.data!.map((i) => i[dimensionName] as string))
      .flat();
    const dedupedValues = new Set(values);
    return [...dedupedValues];
  });
}

export function useAllSearchResultsCount(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  { searchText, values, timeStart, timeEnd, enabled }: DimensionSearchArgs,
): CompoundQueryResult<number | undefined> {
  const where = getFilterForSearchArgs(dimensionName, { searchText, values });

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

  return getCompoundAggregationQuery(queries, (responses) => {
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

function getFilterForSearchArgs(
  dimensionName: string,
  { searchText, values }: DimensionSearchArgs,
) {
  if (searchText) {
    const addNull = searchText.length !== 0 && "null".includes(searchText);
    return addNull
      ? createInExpression(dimensionName, [null])
      : createLikeExpression(dimensionName, `%${searchText}%`);
  } else if (values?.length) {
    return createInExpression(dimensionName, values);
  }

  return undefined;
}
