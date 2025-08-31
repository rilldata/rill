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
import { isFieldConfig, vegaSortToAggregationSort } from "../util";

export type CartesianChartSpec = BaseChartConfig & {
  x?: FieldConfig;
  y?: FieldConfig;
  measures?: string[];
  color?: FieldConfig | string;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = "-y";

export class CartesianChartComponent extends BaseChart<CartesianChartSpec> {
  customSortXItems: string[] = [];
  customColorValues: string[] = [];

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
    y: {
      type: "positional",
      label: "Y-axis",
      showInUI: true,
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
        },
      },
    },
    measures: {
      type: "multi_fields",
      label: "Measures",
      meta: { allowedTypes: ["measure"] },
      showInUI: true,
    },
    // TODO: Refactor to use simpler primitives
    color: {
      type: "mark",
      label: "Color",
      showInUI: true,
      meta: {
        type: "color",
        chartFieldInput: {
          type: "dimension",
          defaultLegendOrientation: "top",
          limitSelector: { defaultLimit: DEFAULT_SPLIT_LIMIT },
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
    key: keyof CartesianChartSpec,
    value: string | string[] | FieldConfig | undefined,
  ): void {
    if (key === "measures") {
      const config = get(this.specStore);
      const currentMeasures = config.measures || [];
      const updatedMeasures = value as string[];

      // Transition from single to multi-measure mode
      if (
        currentMeasures.length === 0 &&
        updatedMeasures &&
        updatedMeasures.length > 0 &&
        config.y?.field
      ) {
        const measuresSet = new Set([config.y.field, ...updatedMeasures]);
        value = Array.from(measuresSet);
      }
      // Transition from multi to single-measure mode
      else if (
        currentMeasures.length > 1 &&
        updatedMeasures &&
        updatedMeasures.length === 1
      ) {
        // When down to one measure, move it to Y-axis and clear measures array
        const singleMeasure = updatedMeasures[0];
        this.specStore.update((spec) => ({
          ...spec,
          y: {
            type: "quantitative",
            field: singleMeasure,
            zeroBasedOrigin: true,
          },
        }));
        value = [];
      }
      // Complete removal of measures
      else if (
        currentMeasures.length > 0 &&
        (!updatedMeasures || updatedMeasures.length === 0)
      ) {
        // Take the last measure and put it in Y-axis
        const lastMeasure = currentMeasures[currentMeasures.length - 1];
        if (lastMeasure) {
          this.specStore.update((spec) => ({
            ...spec,
            y: {
              type: "quantitative",
              field: lastMeasure,
              zeroBasedOrigin: true,
            },
          }));
        }
        value = [];
      }
    }

    super.updateProperty(key, value);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const config = get(this.specStore);
    const isMultiMeasure = config.measures && config.measures.length > 1;

    const inputParams = { ...CartesianChartComponent.chartInputParams };

    inputParams.y.showInUI = !isMultiMeasure;
    inputParams.color.showInUI = !isMultiMeasure;

    const sortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.customSortXItems;
    }
    const colorMappingSelector =
      inputParams.color.meta?.chartFieldInput?.colorMappingSelector;
    if (colorMappingSelector) {
      colorMappingSelector.values = this.customColorValues;
    }

    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    const isMultiMeasure = config.measures && config.measures.length > 1;

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (isMultiMeasure) {
      const measuresSet = new Set(config.measures || []);
      if (config.y?.type === "quantitative" && config.y?.field) {
        measuresSet.add(config.y.field);
      }
      measures = Array.from(measuresSet).map((name) => ({ name }));
    } else {
      if (config.y?.type === "quantitative" && config.y?.field) {
        measures = [{ name: config.y.field }];
      }
    }

    let xAxisSort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;
    let hasColorDimension = false;
    let colorDimensionName = "";
    let colorLimit: number | undefined;

    const dimensionName = config.x?.field;

    if (config.x?.type === "nominal" && dimensionName) {
      limit = config.x.limit ?? 100;
      if (isMultiMeasure) {
        const sort = config.x?.sort;
        if (sort === "y" || sort === "-y") {
          // Use first measure for y-based sorts
          const firstMeasure = config.measures?.[0];
          if (firstMeasure) {
            xAxisSort = {
              name: firstMeasure,
              desc: sort === "-y",
            };
          }
        } else if (sort === "x" || sort === "-x") {
          xAxisSort = {
            name: dimensionName,
            desc: sort === "-x",
          };
        }
      } else {
        xAxisSort = vegaSortToAggregationSort("x", config, DEFAULT_SORT);
      }
      dimensions = [{ name: dimensionName }];
    } else if (config.x?.type === "temporal" && dimensionName) {
      dimensions = [{ name: dimensionName }];
    }

    if (
      typeof config.color === "object" &&
      config.color?.field &&
      !isMultiMeasure
    ) {
      colorDimensionName = config.color.field;
      colorLimit = config.color.limit ?? DEFAULT_SPLIT_LIMIT;
      dimensions = [...dimensions, { name: colorDimensionName }];
      hasColorDimension = true;
    }

    // Create topN query for x dimension
    const topNXQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          hasColorDimension &&
          config.x?.type === "nominal" &&
          !Array.isArray(config.x?.sort) &&
          !!dimensionName;

        const topNWhere = getFilterWithNullHandling(where, config.x);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: dimensionName }],
            sort: xAxisSort ? [xAxisSort] : undefined,
            where: topNWhere,
            timeRange,
            limit: limit?.toString(),
          },
          {
            query: {
              enabled,
            },
          },
        );
      },
    );

    const topNXQuery = createQuery(topNXQueryOptionsStore);

    // Create topN query for color dimension
    const topNColorQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          hasColorDimension &&
          !!colorDimensionName &&
          !!colorLimit;

        const topNWhere = getFilterWithNullHandling(
          where,
          typeof config.color === "object" ? config.color : undefined,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: colorDimensionName }],
            sort: config?.y?.field
              ? [{ name: config.y.field, desc: true }]
              : undefined,
            where: topNWhere,
            timeRange,
            limit: colorLimit?.toString(),
          },
          {
            query: {
              enabled,
            },
          },
        );
      },
    );

    const topNColorQuery = createQuery(topNColorQueryOptionsStore);

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore, topNXQuery, topNColorQuery],
      ([runtime, $timeAndFilterStore, $topNXQuery, $topNColorQuery]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const topNXData = $topNXQuery?.data?.data;

        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          !!dimensions?.length &&
          (hasColorDimension &&
          config.x?.type === "nominal" &&
          !Array.isArray(config.x?.sort)
            ? topNXData !== undefined
            : true) &&
          (hasColorDimension && colorDimensionName && colorLimit
            ? topNColorData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.x,
        );

        let includedXValues: string[] = [];

        // Apply topN filter for x dimension
        if (Array.isArray(config.x?.sort)) {
          includedXValues = config.x.sort;
        } else if (topNXData?.length && dimensionName) {
          includedXValues = topNXData.map((d) => d[dimensionName] as string);
        }

        if (dimensionName) {
          this.customSortXItems = includedXValues;
          const filterForTopXValues = createInExpression(
            dimensionName,
            includedXValues,
          );
          combinedWhere = mergeFilters(combinedWhere, filterForTopXValues);
        }

        // Apply topN filter for color dimension
        if (topNColorData?.length && colorDimensionName) {
          const topColorValues = topNColorData.map(
            (d) => d[colorDimensionName] as string,
          );
          this.customColorValues = topColorValues;
          const filterForTopColorValues = createInExpression(
            colorDimensionName,
            topColorValues,
          );
          combinedWhere = mergeFilters(combinedWhere, filterForTopColorValues);
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
            limit: hasColorDimension || !limit ? "5000" : limit?.toString(),
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
  ): CartesianChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    return {
      metrics_view: metricsViewName,
      color: "primary",
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: DEFAULT_SORT,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y: {
        type: "quantitative",
        field: randomMeasure,
        zeroBasedOrigin: true,
      },
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
    const isMultiMeasure = config.measures && config.measures.length > 1;

    if (isMultiMeasure) {
      // Multi-measure title logic
      const xLabel = config.x?.field
        ? fields[config.x.field]?.displayName || config.x.field
        : "";
      const measuresLabel = (config.measures || [])
        .map((m) => fields[m]?.displayName || m)
        .join(", ");
      const preposition = xLabel === "Time" ? "over" : "by";
      return `${measuresLabel} ${preposition} ${xLabel}`;
    } else {
      // Existing single-measure title logic
      const { x, y, color } = config;
      const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
      const yLabel = y?.field ? fields[y.field]?.displayName || y.field : "";

      const colorLabel =
        typeof color === "object" && color?.field
          ? fields[color.field]?.displayName || color.field
          : "";

      const preposition = xLabel === "Time" ? "over" : "per";

      return colorLabel
        ? `${yLabel} ${preposition} ${xLabel} split by ${colorLabel}`
        : `${yLabel} ${preposition} ${xLabel}`;
    }
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

    if (isFieldConfig(config.color)) {
      result[config.color.field] =
        this.customColorValues.length > 0
          ? [...this.customColorValues]
          : undefined;
    }

    return result;
  }
}
