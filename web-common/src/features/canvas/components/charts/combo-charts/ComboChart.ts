import { getFilterWithNullHandling } from "@rilldata/web-common/features/canvas/components/charts/query-utils";
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
  type V1MetricsViewAggregationSort,
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
import type { ChartDataQuery, ChartFieldsMap, FieldConfig } from "../types";
import { vegaSortToAggregationSort } from "../util";

export type MarkType = "bar" | "line";

export type ComboChartSpec = BaseChartConfig & {
  x?: FieldConfig;
  y1?: FieldConfig;
  y2?: FieldConfig;
  color?: FieldConfig;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SORT = "-y";

export class ComboChartComponent extends BaseChart<ComboChartSpec> {
  customSortXItems: string[] = [];

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          axisTitleSelector: true,
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["x", "-x", "y", "-y", "custom"],
          },
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    y1: {
      type: "positional",
      label: "Left Y-Axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
          markTypeSelector: true,
        },
      },
    },

    y2: {
      type: "positional",
      label: "Right Y-Axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
          markTypeSelector: true,
        },
      },
    },

    color: {
      type: "mark",
      label: "Color",
      meta: {
        type: "color",
        chartFieldInput: {
          type: "value",
          defaultLegendOrientation: "top",
          colorMappingSelector: { enable: true },
          nullSelector: true,
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  updateProperty(
    key: keyof ComboChartSpec,
    value: ComboChartSpec[keyof ComboChartSpec],
  ) {
    const currentSpec = get(this.specStore);

    // Handle mark type mutual exclusivity
    if (key === "y1" || key === "y2") {
      const updatedField = value as FieldConfig;

      if (updatedField?.mark) {
        const otherKey = key === "y1" ? "y2" : "y1";
        const otherField = currentSpec[otherKey];

        // If the other field exists and has the same mark type, switch it
        if (otherField?.mark === updatedField.mark) {
          const oppositeMarkType = updatedField.mark === "bar" ? "line" : "bar";
          const updatedOtherField = { ...otherField, mark: oppositeMarkType };

          const newSpec = {
            ...currentSpec,
            [key]: updatedField,
            [otherKey]: updatedOtherField,
          };

          this.setSpec(newSpec);
          return;
        }
      }
    }
    super.updateProperty(key, value);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = { ...ComboChartComponent.chartInputParams };
    const config = get(this.specStore);

    const sortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.customSortXItems;
    }

    const colorMappingSelector =
      inputParams.color.meta?.chartFieldInput?.colorMappingSelector;
    if (colorMappingSelector && config.y1?.field && config.y2?.field) {
      colorMappingSelector.values = [config.y1.field, config.y2.field];
    }
    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    const measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    // Add both y1 and y2 measures
    if (config.y1?.type === "quantitative" && config.y1?.field) {
      measures.push({ name: config.y1.field });
    }
    if (config.y2?.type === "quantitative" && config.y2?.field) {
      measures.push({ name: config.y2.field });
    }

    let xAxisSort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;

    const dimensionName = config.x?.field;

    if (config.x?.type === "nominal" && dimensionName) {
      limit = config.x.limit ?? 100;
      // TODO: Use better sort config logic
      const cartesianConfig = {
        ...config,
        y: config.y1, // Use y1 as the primary y for sorting
      } as any;
      xAxisSort = vegaSortToAggregationSort("x", cartesianConfig, DEFAULT_SORT);
      dimensions = [{ name: dimensionName }];
    } else if (config.x?.type === "temporal" && dimensionName) {
      dimensions = [{ name: dimensionName }];
    }

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          !!dimensions?.length;

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.x,
        );

        let includedXValues: string[] = [];

        // Apply custom sort for x dimension
        if (Array.isArray(config.x?.sort)) {
          includedXValues = config.x.sort;
          if (dimensionName) {
            this.customSortXItems = includedXValues;
            const filterForTopXValues = createInExpression(
              dimensionName,
              includedXValues,
            );
            combinedWhere = mergeFilters(combinedWhere, filterForTopXValues);
          }
        }

        this.combinedWhere = combinedWhere;
        // Update dimensions with timeGrain if temporal
        if (config.x?.type === "temporal" && timeGrain) {
          dimensions = dimensions.map((d) =>
            d.name === dimensionName ? { ...d, timeGrain } : d,
          );
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort: xAxisSort ? [xAxisSort] : undefined,
            where: combinedWhere,
            timeRange,
            limit: limit?.toString(),
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

    const query = createQuery(queryOptionsStore);
    return query;
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): ComboChartSpec {
    // Randomly select measures and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure1 = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;
    const randomMeasure2 = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    return {
      metrics_view: metricsViewName,
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: DEFAULT_SORT,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y1: {
        type: "quantitative",
        field: randomMeasure1,
        zeroBasedOrigin: true,
        mark: "bar",
      },
      y2: {
        type: "quantitative",
        field: randomMeasure2,
        zeroBasedOrigin: true,
        mark: "line",
      },
      color: {
        type: "nominal",
        field: "measures", // This will be used for legend
        legendOrientation: "top",
      },
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
    const { x, y1, y2 } = config;
    const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
    const y1Label = y1?.field ? fields[y1.field]?.displayName || y1.field : "";
    const y2Label = y2?.field ? fields[y2.field]?.displayName || y2.field : "";

    const preposition = xLabel === "Time" ? "over" : "per";

    const measuresLabel = y2Label ? `${y1Label} & ${y2Label}` : y1Label;

    return `${measuresLabel} ${preposition} ${xLabel}`;
  }

  getChartDomainValues() {
    const config = get(this.specStore);
    const result: Record<string, string[] | undefined> = {};

    if (config.x?.field) {
      result[config.x.field] =
        this.customSortXItems.length > 0
          ? [...this.customSortXItems]
          : undefined;
    }

    return result;
  }
}
