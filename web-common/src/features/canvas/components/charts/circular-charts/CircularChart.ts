import { getFilterWithNullHandling } from "@rilldata/web-common/features/canvas/components/charts/query-utils";
import type {
  ChartFieldsMap,
  FieldConfig,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import type { ChartDataQuery } from "../types";
import { isFieldConfig } from "../util";

type CircularChartEncoding = {
  measure?: FieldConfig;
  color?: FieldConfig;
  innerRadius?: number;
};

const DEFAULT_COLOR_LIMIT = 20;
const DEFAULT_SORT = "-measure";

export type CircularChartSpec = BaseChartConfig & CircularChartEncoding;

export class CircularChartComponent extends BaseChart<CircularChartSpec> {
  customColorValues: string[] = [];
  totalsValue: number | undefined = undefined;

  static chartInputParams: Record<string, ComponentInputParam> = {
    measure: {
      type: "positional",
      label: "Measure",
      meta: {
        chartFieldInput: {
          type: "measure",
          totalSelector: true,
        },
      },
    },
    innerRadius: {
      type: "number",
      label: "Inner Radius (%)",
    },
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
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["color", "-color", "measure", "-measure", "custom"],
          },
          colorMappingSelector: { enable: true },
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = CircularChartComponent.chartInputParams;
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

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.measure?.field) {
      measures = [{ name: config.measure.field }];
    }

    let colorSort: V1MetricsViewAggregationSort | undefined;
    let limit: number;
    const colorDimensionName = config.color?.field;
    const showTotal = config.measure?.showTotal;

    if (colorDimensionName) {
      limit = config.color?.limit || DEFAULT_COLOR_LIMIT;
      dimensions = [{ name: colorDimensionName }];
      colorSort = this.getColorSort(config);
    }

    // Create topN query for color dimension
    const topNColorQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!colorDimensionName &&
          config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: colorDimensionName }],
            sort: colorSort ? [colorSort] : undefined,
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

    const topNColorQuery = createQuery(topNColorQueryOptionsStore);

    const totalQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!showTotal &&
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.measure?.field;

        const totalWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            where: totalWhere,
            timeRange,
          },
          {
            query: {
              enabled,
            },
          },
        );
      },
    );

    const totalQuery = createQuery(totalQueryOptionsStore);

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore, topNColorQuery, totalQuery],
      ([runtime, $timeAndFilterStore, $topNColorQuery, $totalQuery]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          (config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort)
            ? topNColorData !== undefined
            : true);

        let combinedWhere = where;
        let topColorValues: string[] = [];

        // Apply topN filter for color dimension
        if (Array.isArray(config.color?.sort)) {
          topColorValues = config.color.sort;
        } else if (topNColorData?.length && colorDimensionName) {
          topColorValues = topNColorData.map(
            (d) => d[colorDimensionName] as string,
          );
        }

        if (colorDimensionName) {
          this.customColorValues = topColorValues;
          const filterForTopColorValues = createInExpression(
            colorDimensionName,
            topColorValues,
          );
          combinedWhere = mergeFilters(where, filterForTopColorValues);
        }

        const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            where: combinedWhere,
            sort: colorSort ? [colorSort] : undefined,
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

        if (showTotal && config.measure?.field) {
          this.totalsValue = $totalQuery?.data?.data?.[0]?.[
            config.measure?.field
          ] as number;
        }

        return queryOptions;
      },
    );

    const query = createQuery(queryOptionsStore);
    return query;
  }

  private getColorSort(
    config: CircularChartSpec,
  ): V1MetricsViewAggregationSort | undefined {
    if (!config.color?.field) return undefined;

    let sort = config.color.sort;
    if (!sort || Array.isArray(sort)) {
      sort = DEFAULT_SORT;
    }

    let field: string | undefined;
    let desc: boolean = false;

    switch (sort) {
      case "color":
      case "-color":
        field = config.color.field;
        desc = sort === "-color";
        break;
      case "measure":
      case "-measure":
        field = config.measure?.field;
        desc = sort === "-measure";
        break;
      default:
        return undefined;
    }

    if (!field) return undefined;

    return {
      name: field,
      desc,
    };
  }

  getChartDomainValues() {
    const config = get(this.specStore);
    const result: Record<string, string[] | number[] | undefined> = {};

    if (isFieldConfig(config.color)) {
      result[config.color.field] =
        this.customColorValues.length > 0
          ? [...this.customColorValues]
          : undefined;
    }

    if (config.measure?.showTotal && this.totalsValue) {
      result["total"] = [this.totalsValue];
    }

    return result;
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
        sort: DEFAULT_SORT,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
        showTotal: true,
      },
    };
  }
}
