import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { derived, type Readable } from "svelte/store";
import type { ChartSpec } from "./";
import { createChartDataQuery } from "./query";
import type { ChartConfig } from "./types";
import { timeGrainToVegaTimeUnitMap } from "./util";

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
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<ChartDataResult> {
  const chartDataQuery = createChartDataQuery(ctx, config, timeAndFilterStore);
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
      return getTimeDimensionDefinition(field.name, timeAndFilterStore);
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

export function getTimeDimensionDefinition(
  field: string,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<TimeDimensionDefinition> {
  return derived(timeAndFilterStore, ($timeAndFilterStore) => {
    const grain = $timeAndFilterStore?.timeGrain;
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

export function validateChartSchema(
  ctx: StateManagers,
  chartSpec: ChartSpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view, x, y, color } = chartSpec;
  let measures: string[] = [];
  let dimensions: string[] = [];

  if (y?.field) measures = [y.field];
  if (typeof color === "object" && color?.field)
    dimensions = [...dimensions, color.field];

  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

      const timeDimension = metricsView.timeDimension;
      if (x?.field && x.field !== timeDimension) dimensions = [x.field];

      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure ${invalidMeasures} selected`,
        };
      }

      const validateDimensionsRes = validateDimensions(metricsView, dimensions);

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) ${invalidDimensions} selected`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
