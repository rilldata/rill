import type {
  ChartFieldsMap,
  FieldConfig,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import { keepPreviousData } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import type { ChartDataQuery } from "../types";

type CircularChartEncoding = {
  measure?: FieldConfig;
  color?: FieldConfig;
};
export type CircularChartSpec = BaseChartConfig & CircularChartEncoding;

export class CircularChartComponent extends BaseChart<CircularChartSpec> {
  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  protected getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return {
      measure: { type: "positional", label: "Measure" },
      color: { type: "mark", label: "Color", meta: { type: "color" } },
    };
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.measure?.field) {
      measures = [{ name: config.measure.field }];
    }

    let limit: number;
    if (config.color?.field) {
      limit = config.color.limit ?? 20;
      dimensions = [{ name: config.color.field }];
    }

    return derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore], set) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled = !!timeRange?.start && !!timeRange?.end;

        const dataQuery = createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            where,
            timeRange,
            limit: limit.toString(),
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
          ctx.queryClient,
        );

        return derived(dataQuery, ($dataQuery) => {
          return {
            isFetching: $dataQuery.isFetching,
            error: $dataQuery.error,
            data: $dataQuery?.data?.data,
          };
        }).subscribe(set);
      },
    );
  }

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
    const { measure, color } = config;
    const measureLabel = measure?.field
      ? fields[measure.field]?.displayName || measure.field
      : "";
    const colorLabel = color?.field
      ? fields[color.field]?.displayName || color.field
      : "";

    return colorLabel ? `${measureLabel} split by ${colorLabel}` : measureLabel;
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): CircularChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    const randomDimension = dimensions[
      Math.floor(Math.random() * dimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      color: {
        type: "nominal",
        field: randomDimension,
        sort: "-y",
        limit: 20,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
        zeroBasedOrigin: true,
      },
    };
  }
}
