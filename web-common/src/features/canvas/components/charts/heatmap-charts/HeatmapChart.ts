import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import { getFilterWithNullHandling } from "../query-utils";
import type { ChartDataQuery, ChartFieldsMap, FieldConfig } from "../types";

const DEFAULT_NOMINAL_LIMIT = 40;

export type HeatmapChartSpec = BaseChartConfig & {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig;
};

export class HeatmapChartComponent extends BaseChart<HeatmapChartSpec> {
  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          axisTitleSelector: true,
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    y: {
      type: "positional",
      label: "Y-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
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
          defaultLegendOrientation: "right",
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return HeatmapChartComponent.chartInputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    let measures: V1MetricsViewAggregationMeasure[] = [];

    if (config.color?.field) {
      measures = [{ name: config.color.field }];
    }

    // Create top level options store for X axis
    const xAxisQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.x?.field &&
          config?.x?.type !== "temporal";

        const xWhere = getFilterWithNullHandling(where, config.x);

        let limit = DEFAULT_NOMINAL_LIMIT.toString();
        if (config.x?.limit) {
          limit = config.x.limit.toString();
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.x?.field }],
            sort: config.color?.field
              ? [{ name: config.color.field, desc: true }]
              : [{ name: config.x?.field, desc: false }],
            where: xWhere,
            timeRange,
            limit,
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
        );
      },
    );

    // Create top level options store for Y axis
    const yAxisQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.y?.field &&
          config?.y?.type !== "temporal";

        const yWhere = getFilterWithNullHandling(where, config.y);

        let limit = DEFAULT_NOMINAL_LIMIT.toString();
        if (config.y?.limit) {
          limit = config.y.limit.toString();
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.y?.field }],
            sort: config.color?.field
              ? [{ name: config.color.field, desc: true }]
              : [{ name: config.y?.field, desc: false }],
            where: yWhere,
            timeRange,
            limit,
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
        );
      },
    );

    const xAxisQuery = createQuery(xAxisQueryOptionsStore);
    const yAxisQuery = createQuery(yAxisQueryOptionsStore);

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore, xAxisQuery, yAxisQuery],
      ([runtime, $timeAndFilterStore, $xAxisQuery, $yAxisQuery]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const xTopNData = $xAxisQuery?.data?.data;
        const yTopNData = $yAxisQuery?.data?.data;

        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          (config.x?.type === "nominal" ? !!xTopNData?.length : true) &&
          (config.y?.type === "nominal" ? !!yTopNData?.length : true);

        let combinedWhere: V1Expression | undefined = where;

        if (xTopNData?.length && config.x?.field) {
          const xField = config.x.field;
          const xTopValues = xTopNData.map((d) => d[xField] as string);
          const xFilterForTopValues = createInExpression(xField, xTopValues);
          combinedWhere = mergeFilters(combinedWhere, xFilterForTopValues);
        }

        if (yTopNData?.length && config.y?.field) {
          const yField = config.y.field;
          const yTopValues = yTopNData.map((d) => d[yField] as string);
          const yFilterForTopValues = createInExpression(yField, yTopValues);
          combinedWhere = mergeFilters(combinedWhere, yFilterForTopValues);
        }

        let dimensions: V1MetricsViewAggregationDimension[] = [
          ...(config.x?.field ? [{ name: config.x.field }] : []),
          ...(config.y?.field ? [{ name: config.y.field }] : []),
        ];

        // Update dimensions with timeGrain if temporal
        if (timeGrain) {
          dimensions = dimensions.map((d) => {
            if (
              (config.x?.type === "temporal" && d.name === config.x?.field) ||
              (config.y?.type === "temporal" && d.name === config.y?.field)
            ) {
              return { ...d, timeGrain };
            }
            return d;
          });
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort:
              config.x?.type === "nominal"
                ? [{ name: config.x?.field, desc: true }]
                : undefined,
            where: combinedWhere,
            timeRange,
            limit: "5000", // Higher limit for heatmap to show more data points
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
        );
      },
    );

    return createQuery(queryOptionsStore);
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
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y: {
        type: "nominal",
        field: randomDimension2,
        limit: DEFAULT_NOMINAL_LIMIT,
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
