import type {
  ChartDataQuery,
  ChartDomainValues,
  ChartFieldsMap,
  ChartSortDirection,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import {
  getFilterWithNullHandling,
  vegaSortToAggregationSort,
} from "../query-util";

export type HeatmapChartSpec = {
  metrics_view: string;
  x?: FieldConfig<"nominal" | "time">;
  y?: FieldConfig<"nominal" | "time">;
  color?: FieldConfig<"quantitative">;
  show_data_labels?: boolean;
};

export type HeatmapChartDefaultOptions = {
  nominalLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_NOMINAL_LIMIT = 40;
const DEFAULT_SORT = "-color" as ChartSortDirection;

export class HeatmapChartProvider {
  private spec: Readable<HeatmapChartSpec>;
  defaultNominalLimit = DEFAULT_NOMINAL_LIMIT;
  defaultSort = DEFAULT_SORT;

  customSortXItems: string[] = [];
  customSortYItems: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<HeatmapChartSpec>,
    defaultOptions?: HeatmapChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultNominalLimit =
        defaultOptions.nominalLimit || DEFAULT_NOMINAL_LIMIT;
      this.defaultSort = defaultOptions.sort || DEFAULT_SORT;
    }
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);

    let measures: V1MetricsViewAggregationMeasure[] = [];

    if (config.color?.field) {
      measures = [{ name: config.color.field }];
    }

    // Create top level options store for X axis
    const xAxisQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.x?.field &&
          config?.x?.type !== "temporal" &&
          !Array.isArray(config.x?.sort);

        const xWhere = getFilterWithNullHandling(where, config.x);

        let limit = this.defaultNominalLimit.toString();
        if (config.x?.limit) {
          limit = config.x.limit.toString();
        }

        const xAxisSort = vegaSortToAggregationSort(
          "x",
          config,
          this.defaultSort,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
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
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.y?.field &&
          config?.y?.type !== "temporal" &&
          !Array.isArray(config.y?.sort);

        const yWhere = getFilterWithNullHandling(where, config.y);

        let limit = this.defaultNominalLimit.toString();
        if (config.y?.limit) {
          limit = config.y.limit.toString();
        }

        const yAxisSort = vegaSortToAggregationSort(
          "y",
          config,
          this.defaultSort,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
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
      [runtime, timeAndFilterStore, xAxisQuery, yAxisQuery],
      ([$runtime, $timeAndFilterStore, $xAxisQuery, $yAxisQuery]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const xTopNData = $xAxisQuery?.data?.data;
        const yTopNData = $yAxisQuery?.data?.data;

        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          (config.x?.type === "nominal" && !Array.isArray(config.x?.sort)
            ? xTopNData !== undefined
            : true) &&
          (config.y?.type === "nominal" && !Array.isArray(config.y?.sort)
            ? yTopNData !== undefined
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

        // Store combinedWhere for use in BaseChart
        this.combinedWhere.set(combinedWhere);

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
          $runtime.instanceId,
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

  getChartDomainValues(): ChartDomainValues {
    const config = get(this.spec);
    const result: Record<string, string[] | undefined> = {};

    if (config.x?.field) {
      result[config.x.field] =
        this.customSortXItems.length > 0
          ? [...this.customSortXItems]
          : undefined;
    }

    if (config.y?.field) {
      result[config.y.field] =
        this.customSortYItems.length > 0
          ? [...this.customSortYItems]
          : undefined;
    }
    return result;
  }

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const { x, y, color } = config;
    const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
    const yLabel = y?.field ? fields[y.field]?.displayName || y.field : "";
    const colorLabel = color?.field
      ? fields[color.field]?.displayName || color.field
      : "";

    return `${colorLabel} by ${xLabel} and ${yLabel}`;
  }
}
