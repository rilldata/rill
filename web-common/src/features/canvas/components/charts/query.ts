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
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, readable, type Readable } from "svelte/store";

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartConfig & ComponentFilterProperties,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<{
  isFetching: boolean;
  error: HTTPError | null;
  data: V1MetricsViewAggregationResponseDataItem[] | undefined;
}> {
  let measures: V1MetricsViewAggregationMeasure[] = [];
  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.y?.type === "quantitative" && config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  let sort: V1MetricsViewAggregationSort | undefined;
  let limit: number | undefined;
  let hasColorDimension = false;

  return derived(
    [ctx.runtime, timeAndFilterStore],
    ([runtime, $timeAndFilterStore], set) => {
      const { timeRange, where, timeGrain } = $timeAndFilterStore;

      let outerWhere = where;

      if (config.x?.type === "nominal" && config.x?.field) {
        limit = config.x.limit;
        sort = vegaSortToAggregationSort(config.x?.sort, config);
        dimensions = [{ name: config.x?.field }];

        const showNull = !!config.x.showNull;
        if (!showNull) {
          const excludeNullFilter = createInExpression(
            config.x?.field,
            [null],
            true,
          );
          outerWhere = mergeFilters(where, excludeNullFilter);
        }
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

      if (limit && hasColorDimension) {
        topNQuery = createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.x?.field }],
            sort: sort ? [sort] : undefined,
            where: outerWhere,
            timeRange,
            limit: limit.toString(),
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

        let combinedWhere: V1Expression | undefined = outerWhere;
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

        const dataQuery = createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort: sort ? [sort] : undefined,
            where: combinedWhere,
            timeRange,
            limit: hasColorDimension || !limit ? "5000" : limit.toString(),
          },
          {
            query: queryOptions,
          },
        );

        return derived(dataQuery, ($dataQuery) => {
          return {
            isFetching: $dataQuery.isFetching,
            error: $dataQuery.error,
            data: $dataQuery?.data?.data,
          };
        }).subscribe(topNSet);
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
