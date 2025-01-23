import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/features/canvas/components/charts/util";
import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  createQueryServiceMetricsViewAggregation,
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationResponseDataItem,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type ChartDataResult = {
  data: V1MetricsViewAggregationResponseDataItem[];
  isFetching: boolean;
  fields: Record<
    string,
    | MetricsViewSpecMeasureV2
    | MetricsViewSpecDimensionV2
    | TimeDimensionDefinition
    | undefined
  >;
  error?: HTTPError | null;
};

export interface TimeDimensionDefinition {
  field: string;
  displayName: string;
  timeUnit?: string;
  format?: string;
}

export function getChartData(
  ctx: StateManagers,
  config: ChartConfig,
): Readable<ChartDataResult> {
  const chartDataQuery = createChartDataQuery(ctx, config);
  const { spec } = ctx.canvasEntity;

  const fields: { name: string; type: "measure" | "dimension" | "time" }[] = [];
  if (config.y?.field) fields.push({ name: config.y.field, type: "measure" });
  if (config.x?.field)
    fields.push({
      name: config.x.field,
      type: config.x.type === "temporal" ? "time" : "dimension",
    });
  if (typeof config.color === "object" && config.color?.field) {
    fields.push({ name: config.color.field, type: "dimension" });
  }

  // Match each field to its corresponding measure or dimension spec.
  const fieldReadableMap = fields.map((field) => {
    if (field.type === "measure") {
      return spec.getMeasureForMetricView(field.name, config.metrics_view);
    } else if (field.type === "dimension") {
      return spec.getDimensionForMetricView(field.name, config.metrics_view);
    } else {
      return getTimeDimensionDefinition(ctx, field.name);
    }
  });

  return derived(
    [chartDataQuery, ...fieldReadableMap],
    ([chartData, ...fieldMap]) => {
      const fieldSpecMap = fields.reduce(
        (acc, field, index) => {
          acc[field.name] = fieldMap?.[index];
          return acc;
        },
        {} as Record<
          string,
          | MetricsViewSpecMeasureV2
          | MetricsViewSpecDimensionV2
          | TimeDimensionDefinition
          | undefined
        >,
      );
      return {
        data: chartData?.data?.data || [],
        isFetching: chartData.isFetching,
        error: chartData.error,
        fields: fieldSpecMap,
      };
    },
  );
}

export function createChartDataQuery(
  ctx: StateManagers,
  config: ChartConfig & ComponentFilterProperties,
  limit = "500",
  offset = "0",
): CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> {
  const {
    timeControls: { selectedTimeRange },
  } = ctx.canvasEntity;

  let measures: V1MetricsViewAggregationMeasure[] = [];
  let dimensions: V1MetricsViewAggregationDimension[] = [];

  if (config.y?.type === "quantitative" && config.y?.field) {
    measures = [{ name: config.y?.field }];
  }

  return derived(
    [ctx.runtime, selectedTimeRange],
    ([runtime, $selectedTimeRange], set) => {
      let timeRange: V1TimeRange = {
        start: $selectedTimeRange?.start?.toISOString(),
        end: $selectedTimeRange?.end?.toISOString(),
      };

      const timeGrain = $selectedTimeRange?.interval;

      if (config.x?.type === "nominal" && config.x?.field) {
        dimensions = [{ name: config.x?.field }];
      } else if (config.x?.type === "temporal" && timeGrain) {
        dimensions = [{ name: config.x?.field, timeGrain }];
      }

      if (typeof config.color === "object" && config.color?.field) {
        dimensions = [...dimensions, { name: config.color.field }];
      }

      if (config.time_range) {
        timeRange = { isoDuration: config.time_range };
      }

      return createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        config.metrics_view,
        {
          measures,
          dimensions,
          where: undefined,
          timeRange,
          limit,
          offset,
        },
        {
          query: {
            enabled: !!$selectedTimeRange?.start && !!$selectedTimeRange?.end,
            queryClient: ctx.queryClient,
            keepPreviousData: true,
          },
        },
      ).subscribe(set);
    },
  );
}

export function getTimeDimensionDefinition(
  ctx: StateManagers,
  field: string,
): Readable<TimeDimensionDefinition> {
  const {
    timeControls: { selectedTimeRange },
  } = ctx.canvasEntity;
  return derived([selectedTimeRange], ([$selectedTimeRange]) => {
    const grain = $selectedTimeRange?.interval;
    const displayName = "Time";

    if (grain) {
      const timeUnit = timeGrainToVegaTimeUnitMap[grain];
      const format = TIME_GRAIN[grain]?.d3format as string;
      return {
        field,
        timeUnit,
        displayName,
        format,
      };
    }
    return {
      field,
      displayName,
    };
  });
}
