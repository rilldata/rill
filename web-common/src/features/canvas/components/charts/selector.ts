import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import { useStartEndTime } from "@rilldata/web-common/features/canvas/components/kpi/selector";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  createQueryServiceMetricsViewAggregation,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, readable, type Readable } from "svelte/store";
import {
  useMetricsViewSpecDimension,
  useMetricsViewSpecMeasure,
} from "../selectors";

export type ChartDataResult = {
  data: V1MetricsViewAggregationResponseDataItem[];
  measure?: MetricsViewSpecMeasureV2;
  dimension?: MetricsViewSpecDimensionV2;
  isFetching: boolean;
  error?: HTTPError;
};

export function getChartData(
  ctx: StateManagers,
  instanceId: string,
  config: ChartConfig,
): Readable<ChartDataResult> {
  const chartDataQuery = createChartDataQuery(ctx, instanceId, config);

  let measureQuery:
    | CreateQueryResult<MetricsViewSpecMeasureV2 | undefined, HTTPError>
    | Readable<null> = readable(null);
  let dimensionQuery:
    | CreateQueryResult<MetricsViewSpecDimensionV2 | undefined, HTTPError>
    | Readable<null> = readable(null);
  if (config.y?.field) {
    measureQuery = useMetricsViewSpecMeasure(
      instanceId,
      config.metrics_view,
      config.y.field,
    );
  }
  if (config.x?.field) {
    dimensionQuery = useMetricsViewSpecDimension(
      instanceId,
      config.metrics_view,
      config.x.field,
    );
  }

  return derived(
    [chartDataQuery, measureQuery, dimensionQuery],
    ([chartData, measure, dimension]) => {
      const isFetching =
        chartData.isFetching ||
        measure?.isFetching ||
        dimension?.isFetching ||
        false;
      const error = chartData.isError
        ? chartData.error
        : measure?.isError
          ? measure.error
          : dimension?.isError
            ? dimension.error
            : undefined;
      return {
        data: chartData?.data?.data || [],
        measure: measure?.data,
        dimension: dimension?.data,
        isFetching,
        error,
      };
    },
  );
}

export function createChartDataQuery(
  ctx: StateManagers,
  instanceId: string,
  config: ChartConfig,
  limit = "500",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  let measures: V1MetricsViewAggregationMeasure[] = [];
  if (config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.x?.field) {
    dimensions = [{ name: config.x?.field }];
  }
  if (typeof config.color === "object" && config.color?.field) {
    dimensions = [...dimensions, { name: config.color.field }];
  }

  return derived(
    [
      ctx.runtime,
      useStartEndTime(instanceId, config.metrics_view, config.time_range),
    ],
    ([runtime, timeRange], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        config.metrics_view,
        {
          measures,
          dimensions,
          where: undefined,
          timeRange: {
            start: timeRange?.data?.start?.toISOString() || undefined,
            end: timeRange?.data?.end?.toISOString() || undefined,
          },
          limit,
          offset,
        },
        {
          query: {
            enabled: !!timeRange.data,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set),
  );
}
