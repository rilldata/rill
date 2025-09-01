import { getFilterWithNullHandling } from "@rilldata/web-common/features/canvas/components/charts/query-utils";
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
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
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
  innerRadius?: number;
};

const DEFAULT_COLOR_LIMIT = 20;

export type CircularChartSpec = BaseChartConfig & CircularChartEncoding;

export class CircularChartComponent extends BaseChart<CircularChartSpec> {
  static chartInputParams: Record<string, ComponentInputParam> = {
    color: {
      type: "positional",
      label: "Color",
      meta: {
        chartFieldInput: {
          type: "dimension",
          nullSelector: true,
          limitSelector: { defaultLimit: DEFAULT_COLOR_LIMIT },
          hideTimeDimension: true,
          defaultLegendOrientation: "right",
        },
      },
    },
    measure: {
      type: "positional",
      label: "Measure",
      meta: {
        chartFieldInput: {
          type: "measure",
        },
      },
    },
    innerRadius: {
      type: "number",
      label: "Inner Radius (%)",
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return CircularChartComponent.chartInputParams;
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
      limit = config.color.limit ?? DEFAULT_COLOR_LIMIT;
      dimensions = [{ name: config.color.field }];
    }

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled = !!timeRange?.start && !!timeRange?.end;

        const nullHandledWhere = getFilterWithNullHandling(where, config.color);

        this.combinedWhere = nullHandledWhere;
        const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            where: nullHandledWhere,
            sort: [
              ...(config.measure?.field
                ? [{ name: config.measure.field, desc: true }]
                : []),
            ],
            timeRange,
            limit: limit.toString(),
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
        );

        return queryOptions;
      },
    );

    const query = createQuery(queryOptionsStore);
    return query;
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
      innerRadius: 50,
      color: {
        type: "nominal",
        field: randomDimension,
        limit: DEFAULT_COLOR_LIMIT,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
      },
    };
  }
}
