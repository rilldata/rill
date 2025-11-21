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
  type MetricsViewSpecMeasure,
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

export type MarkType = "bar" | "line";

export type ComboChartSpec = {
  metrics_view: string;
  x?: FieldConfig<"nominal" | "time">;
  y1?: FieldConfig<"quantitative" | "mark">;
  y2?: FieldConfig<"quantitative" | "mark">;
  color?: FieldConfig<"nominal">;
};

export type ComboChartDefaultOptions = {
  nominalLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SORT = "-y" as ChartSortDirection;

export class ComboChartProvider {
  private spec: Readable<ComboChartSpec>;
  defaultNominalLimit = DEFAULT_NOMINAL_LIMIT;
  defaultSort = DEFAULT_SORT;

  customSortXItems: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<ComboChartSpec>,
    defaultOptions?: ComboChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultNominalLimit =
        defaultOptions.nominalLimit || DEFAULT_NOMINAL_LIMIT;
      this.defaultSort = defaultOptions.sort || DEFAULT_SORT;
    }
  }

  getMeasureLabels(measures: MetricsViewSpecMeasure[]): string[] | undefined {
    const config = get(this.spec);
    if (config.y1?.field && config.y2?.field) {
      return [config.y1.field, config.y2.field].map((fieldName) => {
        const measure = measures.find((m) => m.name === fieldName);
        return measure?.displayName || fieldName;
      });
    }
    return undefined;
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);

    const measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    // Add both y1 and y2 measures
    if (config.y1?.type === "quantitative" && config.y1?.field) {
      measures.push({ name: config.y1.field });
    }
    if (config.y2?.type === "quantitative" && config.y2?.field) {
      measures.push({ name: config.y2.field });
    }

    const dimensionName = config.x?.field;

    const xAxisQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!dimensionName &&
          config?.x?.type === "nominal" &&
          !Array.isArray(config.x?.sort) &&
          !!config.y1?.field;

        const xWhere = getFilterWithNullHandling(where, config.x);

        let limit = this.defaultNominalLimit.toString();
        if (config.x?.limit) {
          limit = config.x.limit.toString();
        }

        const xAxisMeasures = config.y1?.field
          ? [{ name: config.y1.field }]
          : [];

        const xAxisSort = vegaSortToAggregationSort(
          "x",
          config,
          this.defaultSort,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
          config.metrics_view,
          {
            measures: xAxisMeasures,
            dimensions: [{ name: dimensionName }],
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

    const xAxisQuery = createQuery(xAxisQueryOptionsStore);

    const queryOptionsStore = derived(
      [runtime, timeAndFilterStore, xAxisQuery],
      ([$runtime, $timeAndFilterStore, $xAxisQuery]) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;
        const xTopNData = $xAxisQuery?.data?.data;

        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          (config.x?.type === "nominal" && !Array.isArray(config.x?.sort)
            ? xTopNData !== undefined
            : !!dimensionName);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.x,
        );

        let includedXValues: string[] = [];

        // Handle X axis values
        if (dimensionName) {
          if (Array.isArray(config.x?.sort)) {
            includedXValues = config.x.sort;
          } else if (xTopNData?.length && config.x?.type === "nominal") {
            includedXValues = xTopNData.map((d) => d[dimensionName] as string);
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

        if (dimensionName) {
          dimensions = [{ name: dimensionName }];
        }

        // Store combinedWhere for use in BaseChart
        this.combinedWhere.set(combinedWhere);

        // Update dimensions with timeGrain if temporal
        if (config.x?.type === "temporal" && timeGrain) {
          dimensions = dimensions.map((d) =>
            d.name === dimensionName ? { ...d, timeGrain } : d,
          );
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            where: combinedWhere,
            timeRange,
            fillMissing: config.x?.type === "temporal",
            sort:
              config.x?.type === "temporal"
                ? [{ name: config.x?.field, desc: false }]
                : undefined,
            limit:
              config.x?.type === "temporal"
                ? "5000"
                : config.x?.limit?.toString(),
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

  getChartDomainValues(measures: MetricsViewSpecMeasure[]): ChartDomainValues {
    const config = get(this.spec);
    const result: Record<string, string[] | undefined> = {};

    if (config.x?.field) {
      result[config.x.field] =
        this.customSortXItems.length > 0
          ? [...this.customSortXItems]
          : undefined;
    }
    if (config.color?.field) {
      result[config.color?.field] = this.getMeasureLabels(measures);
    }
    return result;
  }

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const { x, y1, y2 } = config;
    const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
    const y1Label = y1?.field ? fields[y1.field]?.displayName || y1.field : "";
    const y2Label = y2?.field ? fields[y2.field]?.displayName || y2.field : "";

    const preposition = xLabel === "Time" ? "over" : "per";

    const measuresLabel = y2Label ? `${y1Label} & ${y2Label}` : y1Label;

    return `${measuresLabel} ${preposition} ${xLabel}`;
  }
}
