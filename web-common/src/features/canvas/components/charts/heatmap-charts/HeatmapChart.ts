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
import { vegaSortToAggregationSort } from "../util";

const DEFAULT_NOMINAL_LIMIT = 40;
const DEFAULT_SORT = "-color";

type HeatmapChartEncoding = {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig;
  show_data_labels?: boolean;
};

export type HeatmapChartSpec = BaseChartConfig & HeatmapChartEncoding;

export class HeatmapChartComponent extends BaseChart<HeatmapChartSpec> {
  customSortXItems: string[] = [];
  customSortYItems: string[] = [];

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["x", "-x", "color", "-color", "custom"],
          },
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
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["y", "-y", "color", "-color", "custom"],
          },
          axisTitleSelector: true,
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
    show_data_labels: {
      type: "boolean",
      label: "Data labels",
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = HeatmapChartComponent.chartInputParams;
    const xSortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (xSortSelector) {
      xSortSelector.customSortItems = this.customSortXItems;
    }
    const ySortSelector = inputParams.y.meta?.chartFieldInput?.sortSelector;
    if (ySortSelector) {
      ySortSelector.customSortItems = this.customSortYItems;
    }
    return inputParams;
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
          config?.x?.type !== "temporal" &&
          !Array.isArray(config.x?.sort);

        const xWhere = getFilterWithNullHandling(where, config.x);

        let limit = DEFAULT_NOMINAL_LIMIT.toString();
        if (config.x?.limit) {
          limit = config.x.limit.toString();
        }

        const xAxisSort = vegaSortToAggregationSort("x", config, DEFAULT_SORT);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.x?.field }],
            sort: xAxisSort ? [xAxisSort] : undefined,
            where: xWhere,
            timeRange,
            limit,
          },
          {
            query: {
              enabled,
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
          config?.y?.type !== "temporal" &&
          !Array.isArray(config.y?.sort);

        const yWhere = getFilterWithNullHandling(where, config.y);

        let limit = DEFAULT_NOMINAL_LIMIT.toString();
        if (config.y?.limit) {
          limit = config.y.limit.toString();
        }

        const yAxisSort = vegaSortToAggregationSort("y", config, DEFAULT_SORT);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: config.y?.field }],
            sort: yAxisSort ? [yAxisSort] : undefined,
            where: yWhere,
            timeRange,
            limit,
          },
          {
            query: {
              enabled,
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
          (config.x?.type === "nominal" && !Array.isArray(config.x?.sort)
            ? !!xTopNData?.length
            : true) &&
          (config.y?.type === "nominal" && !Array.isArray(config.y?.sort)
            ? !!yTopNData?.length
            : true);

        let combinedWhere: V1Expression | undefined = where;

        let includedXValues: string[] = [];
        let includedYValues: string[] = [];

        // Handle X axis values
        if (config.x?.field) {
          if (Array.isArray(config.x.sort)) {
            includedXValues = config.x.sort;
          } else if (xTopNData?.length) {
            const xField = config.x.field;
            includedXValues = xTopNData.map((d) => d[xField] as string);
          }

          if (includedXValues.length > 0) {
            this.customSortXItems = includedXValues;
            const xFilterForTopValues = createInExpression(
              config.x.field,
              includedXValues,
            );
            combinedWhere = mergeFilters(combinedWhere, xFilterForTopValues);
          }
        }

        // Handle Y axis values
        if (config.y?.field) {
          if (Array.isArray(config.y.sort)) {
            includedYValues = config.y.sort;
          } else if (yTopNData?.length) {
            const yField = config.y.field;
            includedYValues = yTopNData.map((d) => d[yField] as string);
          }

          if (includedYValues.length > 0) {
            this.customSortYItems = includedYValues;
            const yFilterForTopValues = createInExpression(
              config.y.field,
              includedYValues,
            );
            combinedWhere = mergeFilters(combinedWhere, yFilterForTopValues);
          }
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

        this.combinedWhere = combinedWhere;

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
