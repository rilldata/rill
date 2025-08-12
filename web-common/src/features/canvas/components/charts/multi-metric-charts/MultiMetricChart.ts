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

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SORT = "-x";

type MultiMetricChartEncoding = {
  x?: FieldConfig;
  measures?: string[];
  mark_type?: "stacked_bar" | "grouped_bar" | "stacked_area" | "line";
};

export type MultiMetricChartSpec = BaseChartConfig & MultiMetricChartEncoding;

export class MultiMetricChartComponent extends BaseChart<MultiMetricChartSpec> {
  customSortXItems: string[] = [];

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["x", "-x", "custom"],
          },
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          axisTitleSelector: true,
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    measures: {
      type: "multi_fields",
      label: "Measures",
      meta: { allowedTypes: ["measure"] },
    },
    mark_type: {
      type: "select",
      label: "Mark type",
      meta: {
        options: [
          { label: "Stacked bar", value: "stacked_bar" },
          { label: "Grouped bar", value: "grouped_bar" },
          { label: "Stacked area", value: "stacked_area" },
          { label: "Line", value: "line" },
        ],
        default: "grouped_bar",
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = MultiMetricChartComponent.chartInputParams;
    const sortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.customSortXItems;
    }
    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    const measures: V1MetricsViewAggregationMeasure[] = (
      config.measures || []
    ).map((name) => ({ name }));

    let dimensions: V1MetricsViewAggregationDimension[] = [];

    const dimensionName = config.x?.field;
    if (config.x?.type === "nominal" && dimensionName) {
      dimensions = [{ name: dimensionName }];
    } else if (config.x?.type === "temporal" && dimensionName) {
      dimensions = [{ name: dimensionName }];
    }

    const xAxisSort = vegaSortToAggregationSort("x", config, DEFAULT_SORT);
    const limit: number | undefined =
      config.x?.type === "nominal"
        ? (config.x.limit ?? DEFAULT_NOMINAL_LIMIT)
        : undefined;

    // TopN query for x dimension (only when x is nominal and sort not array)
    const topNXQueryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!dimensionName &&
          config.x?.type === "nominal" &&
          !Array.isArray(config.x?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.x);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: dimensionName ? [{ name: dimensionName }] : [],
            sort: xAxisSort ? [xAxisSort] : undefined,
            where: topNWhere,
            timeRange,
            limit: (limit ?? DEFAULT_NOMINAL_LIMIT).toString(),
          },
          {
            query: { enabled },
          },
        );
      },
    );

    const topNXQuery = createQuery(topNXQueryOptionsStore);

    const queryOptionsStore = derived(
      [ctx.runtime, timeAndFilterStore, topNXQuery],
      ([runtime, $timeAndFilterStore, $topNXQuery]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const topNXData = $topNXQuery?.data?.data;

        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          measures.length > 0 &&
          !!dimensionName &&
          (config.x?.type === "nominal" && !Array.isArray(config.x?.sort)
            ? topNXData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.x,
        );

        let includedXValues: string[] = [];
        if (config.x?.type === "nominal" && dimensionName) {
          if (Array.isArray(config.x?.sort)) {
            includedXValues = config.x.sort;
          } else if (topNXData?.length) {
            includedXValues = topNXData.map((d) => d[dimensionName] as string);
          }
          if (includedXValues.length > 0) {
            this.customSortXItems = includedXValues;
            const filterForTopXValues = createInExpression(
              dimensionName,
              includedXValues,
            );
            combinedWhere = mergeFilters(combinedWhere, filterForTopXValues);
          }
        }

        this.combinedWhere = combinedWhere;

        // Update time grain for temporal x
        if (config.x?.type === "temporal" && timeGrain && dimensionName) {
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
            limit: limit ? limit.toString() : "5000",
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
  ): MultiMetricChartSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 2).map((m) => m.name as string) || [];
    const dimensions = metricsViewSpec?.dimensions || [];
    const timeDimension = metricsViewSpec?.timeDimension;

    const xField = timeDimension
      ? timeDimension
      : (dimensions[0]?.name as string) || "";

    return {
      metrics_view: metricsViewName,
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: xField,
        sort: DEFAULT_SORT,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      measures,
      mark_type: "grouped_bar",
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
    const xLabel = config.x?.field
      ? fields[config.x.field]?.displayName || config.x.field
      : "";
    const measures = config.measures || [];
    const measuresLabel = measures
      .map((m) => fields[m]?.displayName || m)
      .join(", ");
    const preposition = xLabel === "Time" ? "over" : "by";
    return `${measuresLabel} ${preposition} ${xLabel}`;
  }

  getChartDomainValues() {
    return {
      xValues:
        this.customSortXItems.length > 0
          ? [...this.customSortXItems]
          : undefined,
    };
  }
}
