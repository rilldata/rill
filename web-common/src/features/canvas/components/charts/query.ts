import type {
  ChartConfig,
  ChartSortDirection,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, readable, type Readable } from "svelte/store";

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartConfig & ComponentFilterProperties,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  let measures: V1MetricsViewAggregationMeasure[] = [];
  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.y?.type === "quantitative" && config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  let sort: V1MetricsViewAggregationSort | undefined;
  let topN: number | undefined;
  let hasColorDimension = false;

  return derived(
    [ctx.runtime, timeAndFilterStore],
    ([runtime, $timeAndFilterStore], set) => {
      const { timeRange, where, timeGrain } = $timeAndFilterStore;

      if (config.x?.type === "nominal" && config.x?.field) {
        topN = config.x.topN;
        sort = vegaSortToAggregationSort(config.x?.sort, config);
        dimensions = [{ name: config.x?.field }];
      } else if (config.x?.type === "temporal" && timeGrain) {
        dimensions = [{ name: config.x?.field, timeGrain }];
      }

      if (typeof config.color === "object" && config.color?.field) {
        dimensions = [...dimensions, { name: config.color.field }];
        hasColorDimension = true;
      }

      const queryOptions = {
        enabled: !!timeRange?.start && !!timeRange?.end,
        queryClient: ctx.queryClient,
        keepPreviousData: true,
      };

      let topNQuery:
        | Readable<null>
        | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
        readable(null);

      if (topN && hasColorDimension) {
        topNQuery = createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.x?.field }],
            sort: sort ? [sort] : undefined,
            where,
            timeRange,
            limit: topN.toString(),
          },
          {
            query: queryOptions,
          },
        );
      }

      return derived(topNQuery, ($topNQuery, topNSet) => {
        if ($topNQuery !== null && !$topNQuery?.data) {
          return topNSet({
            isFetching: $topNQuery.isFetching,
            error: $topNQuery.error,
            data: undefined,
          });
        }

        const dimensionName = config.x?.field;

        let combinedWhere: V1Expression | undefined = where;
        if ($topNQuery?.data?.data?.length && dimensionName) {
          const topValues = $topNQuery?.data?.data.map(
            (d) => d[dimensionName] as string,
          );
          const filterForTopValues = createInExpression(
            dimensionName,
            topValues,
          );
          combinedWhere = mergeFilters(where, filterForTopValues);
        }

        return createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort: sort ? [sort] : undefined,
            where: combinedWhere,
            timeRange,
            limit: "5000",
            offset: "0",
          },
          {
            query: queryOptions,
          },
        ).subscribe(topNSet);
      }).subscribe(set);
    },
  );
}

function vegaSortToAggregationSort(
  sort: ChartSortDirection | undefined,
  config: ChartConfig,
): V1MetricsViewAggregationSort | undefined {
  if (!sort) return undefined;
  const field =
    sort === "x" || sort === "-x" ? config.x?.field : config.y?.field;
  if (!field) return undefined;

  return {
    name: field,
    desc: sort === "-x" || sort === "-y",
  };
}
