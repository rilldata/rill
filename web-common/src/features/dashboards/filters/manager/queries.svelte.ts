import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";
import type { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/dimension-filter-manager.svelte.ts";
import { createQuery } from "@tanstack/svelte-query";
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

type DimensionSearchArgs = {
  mode: DimensionFilterMode;
  searchText: string;
  values: string[];
  timeStart?: string;
  timeEnd?: string;
  timeDimension?: string;
  enabled?: boolean;
};

export function getDimensionSearchQuery(
  client: RuntimeClient,
  manager: ExpressionFilterManager,
  dimensionManager: DimensionFilterManager,
  {
    mode,
    searchText,
    values,
    timeStart,
    timeEnd,
    timeDimension,
    enabled,
  }: DimensionSearchArgs,
) {
  const dimensionName = dimensionManager.name;
  const mvName = [...dimensionManager.dimensions.keys()][0] ?? "";

  const otherDimsFilter = $derived.by(() => {
    manager.expr; // Force rederivation when expr changes.
    return manager.getOtherDimensionsFilter(dimensionManager.name);
  });
  const where = getFilterForSearchArgs(dimensionManager.name, {
    mode,
    searchText,
    values,
    additionalFilter: otherDimsFilter,
  });

  const optionsStore = $derived(
    getQueryServiceMetricsViewAggregationQueryOptions(
      client,
      {
        metricsView: mvName,
        dimensions: [{ name: dimensionName }],
        timeRange: { start: timeStart, end: timeEnd, timeDimension },
        limit: "250",
        offset: "0",
        sort: [{ name: dimensionName }],
        where,
      },
      {
        query: {
          select: (resp) =>
            resp.data?.map((d) => d[dimensionName] as string) ?? [],
          enabled,
        },
      },
    ),
  );
  return createQuery(optionsStore);
}

export function getAllSearchResultsCount(
  client: RuntimeClient,
  manager: ExpressionFilterManager,
  dimensionManager: DimensionFilterManager,
  {
    mode,
    searchText,
    values,
    timeStart,
    timeEnd,
    timeDimension,
    enabled,
  }: DimensionSearchArgs,
) {
  const dimensionName = dimensionManager.name;
  const mvName = [...dimensionManager.dimensions.keys()][0] ?? "";
  const countMeasureName = dimensionName + "__distinct_count";

  const otherDimsFilter = $derived.by(() => {
    manager.expr; // Force rederivation when expr changes.
    return manager.getOtherDimensionsFilter(dimensionManager.name);
  });
  const where = getFilterForSearchArgs(dimensionManager.name, {
    mode,
    searchText,
    values,
    additionalFilter: otherDimsFilter,
  });
  const optionsStore = $derived(
    getQueryServiceMetricsViewAggregationQueryOptions(
      client,
      {
        metricsView: mvName,
        measures: [
          {
            name: countMeasureName,
            builtinMeasure: V1BuiltinMeasure.BUILTIN_MEASURE_COUNT_DISTINCT,
            builtinMeasureArgs: [dimensionName],
          },
        ],
        timeRange: { start: timeStart, end: timeEnd, timeDimension },
        where,
      },
      {
        query: {
          enabled,
          select: (resp) => {
            if (!resp.data?.length) return 0;
            return resp.data[0][countMeasureName] as number;
          },
        },
      },
    ),
  );
  return createQuery(optionsStore);
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
