import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

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

  return derived(queries, ($queries) => {
    const someQueryFetching = $queries.some((q) => q.isFetching);
    if (someQueryFetching) {
      return {
        data: undefined,
        error: undefined,
        isFetching: true,
      };
    }
    const errors = $queries.filter((q) => q.isError).map((q) => q.error);
    if (errors.length > 0) {
      return {
        data: undefined,
        // TODO: merge multiple errors
        error: errors[0]?.response?.data.message,
        isFetching: false,
      };
    }

    const items = $queries.flatMap((query) => query.data?.data || []);
    const values = items.map((item) => item[dimensionName] as string);
    const dedupedValues = new Set(values);
    return {
      data: [...dedupedValues],
      error: undefined,
      isFetching: false,
    };
  });
}
