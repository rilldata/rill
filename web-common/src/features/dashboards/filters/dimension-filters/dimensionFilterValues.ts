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

export function useDimensionSearch(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  searchText: string,
  timeStart?: string,
  timeEnd?: string,
  enabled?: boolean,
): CompoundQueryResult<string[]> {
  const addNull = searchText.length !== 0 && "null".includes(searchText);

  const queries = metricsViewNames.map((mvName) =>
    createQueryServiceMetricsViewAggregation(
      instanceId,
      mvName,
      {
        dimensions: [{ name: dimensionName }],
        timeRange: { start: timeStart, end: timeEnd },
        limit: "100",
        offset: "0",
        sort: [{ name: dimensionName }],
        where: addNull
          ? createInExpression(dimensionName, [null])
          : createLikeExpression(dimensionName, `%${searchText}%`),
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

export function useSearchMatchedCount(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  searchText: string,
  timeStart?: string,
  timeEnd?: string,
  enabled?: boolean,
): CompoundQueryResult<number | undefined> {
  const addNull = searchText.length !== 0 && "null".includes(searchText);

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
        limit: limit.toString(),
        offset: "0",
        where: addNull
          ? createInExpression(dimensionName, [null])
          : createLikeExpression(dimensionName, `%${searchText}%`),
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

const limit = 250;
export function useBulkSearchResults(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  values: string[],
  timeStart?: string,
  timeEnd?: string,
  enabled?: boolean,
): CompoundQueryResult<string[]> {
  const queries = metricsViewNames.map((mvName) =>
    createQueryServiceMetricsViewAggregation(
      instanceId,
      mvName,
      {
        dimensions: [{ name: dimensionName }],
        timeRange: { start: timeStart, end: timeEnd },
        limit: limit.toString(),
        offset: "0",
        sort: [{ name: dimensionName }],
        where: createInExpression(dimensionName, values),
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

export function useBulkSearchMatchedCount(
  instanceId: string,
  metricsViewNames: string[],
  dimensionName: string,
  values: string[],
  timeStart?: string,
  timeEnd?: string,
  enabled?: boolean,
): CompoundQueryResult<number | undefined> {
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
        limit: limit.toString(),
        offset: "0",
        where: createInExpression(dimensionName, values),
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
