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
) {
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
    const items = $queries.flatMap((query) => query.data?.data || []);
    const values = items.map((item) => item[dimensionName] as string);
    const seen = new Set();
    return values.filter((value) => {
      if (seen.has(value)) {
        return false;
      }
      seen.add(value);
      return true;
    });
  });
}
