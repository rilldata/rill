import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationSort,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { keepPreviousData } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import type { ChartDataQuery, ChartFieldsMap, FieldConfig } from "../types";

export type HeatmapChartSpec = BaseChartConfig & {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig;
};

export class HeatmapChartComponent extends BaseChart<HeatmapChartSpec> {
  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  protected getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return {
      x: {
        type: "positional",
        label: "X-axis",
        meta: {
          chartFieldInput: {
            type: "dimension",
            axisTitleSelector: true,
            sortSelector: true,
            limitSelector: true,
            nullSelector: true,
          },
        },
      },
      y: {
        type: "positional",
        label: "Y-axis",
        meta: {
          chartFieldInput: {
            type: "dimension",
            axisTitleSelector: true,
            sortSelector: true,
            limitSelector: true,
            nullSelector: true,
          },
        },
      },
      color: {
        type: "positional",
        label: "Color",
        meta: {
          chartFieldInput: {
            type: "measure",
            axisTitleSelector: true,
          },
        },
      },
    };
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.color?.field) {
      measures = [{ name: config.color.field }];
    }

    let xSort: V1MetricsViewAggregationSort | undefined;
    let ySort: V1MetricsViewAggregationSort | undefined;

    if (config.x?.field) {
      dimensions = [...dimensions, { name: config.x.field }];
    }

    if (config.y?.field) {
      dimensions = [...dimensions, { name: config.y.field }];
    }

    return derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore], set) => {
        const { timeRange, where } = $timeAndFilterStore;

        let outerWhere = where;

        // Handle null filtering for both x and y dimensions
        if (config.x?.field && !config.x.showNull) {
          const excludeNullFilter = createInExpression(
            config.x.field,
            [null],
            true,
          );
          outerWhere = mergeFilters(outerWhere, excludeNullFilter);
        }

        if (config.y?.field && !config.y.showNull) {
          const excludeNullFilter = createInExpression(
            config.y.field,
            [null],
            true,
          );
          outerWhere = mergeFilters(outerWhere, excludeNullFilter);
        }

        const enabled = !!timeRange?.start && !!timeRange?.end;

        const dataQuery = createQueryServiceMetricsViewAggregation(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort: [...(xSort ? [xSort] : []), ...(ySort ? [ySort] : [])],
            where: outerWhere,
            timeRange,
            limit: "5000", // Higher limit for heatmap to show more data points
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

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): HeatmapChartSpec {
    // Select two dimensions and one measure if available
    const measures = metricsViewSpec?.measures || [];
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    // Get two random dimensions
    const availableDimensions = [...dimensions];
    const randomDimension1 = availableDimensions.splice(
      Math.floor(Math.random() * availableDimensions.length),
      1,
    )[0]?.name as string;
    const randomDimension2 = availableDimensions[
      Math.floor(Math.random() * availableDimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      x: {
        type: "nominal",
        field: randomDimension1,
        limit: 20,
      },
      y: {
        type: "nominal",
        field: randomDimension2,
        limit: 20,
      },
      color: {
        type: "quantitative",
        field: randomMeasure,
      },
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
    const { x, y, color } = config;
    const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
    const yLabel = y?.field ? fields[y.field]?.displayName || y.field : "";
    const colorLabel = color?.field
      ? fields[color.field]?.displayName || color.field
      : "";

    return `${colorLabel} by ${xLabel} and ${yLabel}`;
  }
}
