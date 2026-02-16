import {
  ChartSortType,
  type ChartDataQuery,
  type ChartDomainValues,
  type ChartFieldsMap,
  type ChartSortDirection,
  type FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import {
  isFieldConfig,
  isMultiFieldConfig,
} from "@rilldata/web-common/features/components/charts/util";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type MetricsViewSpecMeasure,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1TimeRange,
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
  isSortByDelta,
  vegaSortToAggregationSort,
} from "../query-util";

export type CartesianChartSpec = {
  metrics_view: string;
  x?: FieldConfig<"nominal" | "time">;
  y?: FieldConfig<"quantitative">;
  color?: FieldConfig<"nominal"> | string;
};

export type CartesianChartDefaultOptions = {
  nominalLimit?: number;
  splitLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = ChartSortType.Y_DESC as ChartSortDirection;

export class CartesianChartProvider {
  private spec: Readable<CartesianChartSpec>;
  defaultNominalLimit = DEFAULT_NOMINAL_LIMIT;
  defaultSplitLimit = DEFAULT_SPLIT_LIMIT;
  defaultSort = DEFAULT_SORT;

  customSortXItems: string[] = [];
  customColorValues: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<CartesianChartSpec>,
    defaultOptions?: CartesianChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultNominalLimit =
        defaultOptions.nominalLimit || DEFAULT_NOMINAL_LIMIT;
      this.defaultSplitLimit = defaultOptions.splitLimit || DEFAULT_SPLIT_LIMIT;
      this.defaultSort = defaultOptions.sort || DEFAULT_SORT;
    }
  }

  getMeasureLabels(measures: MetricsViewSpecMeasure[]): string[] | undefined {
    const config = get(this.spec);

    if (isMultiFieldConfig(config.y)) {
      return config.y.fields?.map((fieldName) => {
        const measure = measures.find((m) => m.name === fieldName);
        return measure?.displayName || fieldName;
      });
    }
    return undefined;
  }

  /**
   * Resolves the x-axis sort for the aggregation query.
   * For multi-measure charts, uses the first measure for y-based sorts.
   * Delegates to vegaSortToAggregationSort which handles y_delta -> y
   * fallback when comparison is not active.
   */
  private resolveXAxisSort(
    config: CartesianChartSpec,
    isMultiMeasure: boolean,
    isComparisonActive: boolean,
  ) {
    if (config.x?.type !== "nominal" || !config.x?.field) {
      return undefined;
    }

    if (isMultiMeasure) {
      const sort = config.x?.sort;
      const sortIsDelta = isSortByDelta(sort);
      if (
        sort === ChartSortType.Y_ASC ||
        sort === ChartSortType.Y_DESC ||
        sortIsDelta
      ) {
        // Use first measure for y-based sorts
        const firstMeasure = config.y?.fields?.[0];
        if (!firstMeasure) return undefined;

        const isDesc =
          sort === ChartSortType.Y_DESC || sort === ChartSortType.Y_DELTA_DESC;
        return {
          name:
            sortIsDelta && isComparisonActive
              ? firstMeasure + ComparisonDeltaAbsoluteSuffix
              : firstMeasure,
          desc: isDesc,
        };
      } else if (
        sort === ChartSortType.X_ASC ||
        sort === ChartSortType.X_DESC
      ) {
        return {
          name: config.x.field,
          desc: sort === ChartSortType.X_DESC,
        };
      }
      return undefined;
    }

    return vegaSortToAggregationSort(
      "x",
      config,
      this.defaultSort,
      isComparisonActive,
    );
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);

    const isMultiMeasure = isMultiFieldConfig(config.y);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    let measuresSet = new Set<string>();
    if (isMultiMeasure) {
      measuresSet = new Set(config.y?.fields);
      if (config.y?.type === "quantitative" && config.y?.field) {
        measuresSet.add(config.y.field);
      }
      measures = Array.from(measuresSet).map((name) => ({ name }));
    } else {
      if (config.y?.type === "quantitative" && config.y?.field) {
        measuresSet = new Set([config.y.field]);
        measures = [{ name: config.y.field }];
      }
    }

    let limit: number | undefined;
    let hasColorDimension = false;
    let colorDimensionName = "";
    let colorLimit: number | undefined;

    const dimensionName = config.x?.field;
    const rawSort = config.x?.sort;
    const sortIsDelta = isSortByDelta(rawSort);

    if (config.x?.type === "nominal" && dimensionName) {
      limit = config.x.limit ?? 100;
      dimensions = [{ name: dimensionName }];
    } else if (config.x?.type === "temporal" && dimensionName) {
      dimensions = [{ name: dimensionName }];
    }

    if (isFieldConfig(config.color) && !isMultiMeasure) {
      colorDimensionName = config.color.field;
      colorLimit = config.color.limit ?? this.defaultSplitLimit;
      dimensions = [...dimensions, { name: colorDimensionName }];
      hasColorDimension = true;
    }

    // Create topN query for x dimension
    const topNXQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const {
          timeRange,
          where,
          hasTimeSeries,
          comparisonTimeRange,
          showTimeComparison,
        } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
          config.x?.type === "nominal" &&
          !Array.isArray(config.x?.sort) &&
          !!dimensionName;

        const topNWhere = getFilterWithNullHandling(where, config.x);

        const isComparisonActive =
          showTimeComparison &&
          !!comparisonTimeRange?.start &&
          !!comparisonTimeRange?.end;

        const xAxisSort = this.resolveXAxisSort(
          config,
          isMultiMeasure,
          isComparisonActive,
        );

        // For delta sort with active comparison, add delta computed measure
        let topNMeasures = measures;
        let topNComparisonTimeRange: V1TimeRange | undefined = undefined;

        if (sortIsDelta && isComparisonActive) {
          const sortMeasureName = isMultiMeasure
            ? config.y?.fields?.[0]
            : config.y?.field;

          if (sortMeasureName) {
            const deltaFieldName =
              sortMeasureName + ComparisonDeltaAbsoluteSuffix;
            topNMeasures = [
              ...measures,
              {
                name: deltaFieldName,
                comparisonDelta: {
                  measure: sortMeasureName,
                },
              },
            ];
            topNComparisonTimeRange = comparisonTimeRange;
          }
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
          config.metrics_view,
          {
            measures: topNMeasures,
            dimensions: [{ name: dimensionName }],
            sort: xAxisSort ? [xAxisSort] : undefined,
            where: topNWhere,
            timeRange: hasTimeSeries ? timeRange : undefined,
            comparisonTimeRange: topNComparisonTimeRange,
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
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;
        const enabled =
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
          hasColorDimension &&
          !!colorDimensionName &&
          !!colorLimit;

        const topNWhere = getFilterWithNullHandling(
          where,
          typeof config.color === "object" ? config.color : undefined,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
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
      [runtime, timeAndFilterStore, topNXQuery, topNColorQuery],
      ([$runtime, $timeAndFilterStore, $topNXQuery, $topNColorQuery]) => {
        const {
          timeRange,
          where,
          timeGrain,
          comparisonTimeRange,
          showTimeComparison,
          hasTimeSeries,
        } = $timeAndFilterStore;
        const topNXData = $topNXQuery?.data?.data;

        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
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

        // Store combinedWhere for use in BaseChart
        this.combinedWhere.set(combinedWhere);

        // Update dimensions with timeGrain if temporal
        if (config.x?.type === "temporal" && timeGrain) {
          dimensions = dimensions.map((d) =>
            d.name === dimensionName ? { ...d, timeGrain } : d,
          );
        }

        const isComparisonActive =
          showTimeComparison &&
          !!comparisonTimeRange?.start &&
          !!comparisonTimeRange?.end;

        const measuresWithComparison: V1MetricsViewAggregationMeasure[] =
          Array.from(measuresSet)
            .map((measureName) => {
              if (showTimeComparison && comparisonTimeRange?.start) {
                const result: V1MetricsViewAggregationMeasure[] = [
                  { name: measureName },
                  {
                    name: measureName + ComparisonDeltaPreviousSuffix,
                    comparisonValue: {
                      measure: measureName,
                    },
                  },
                ];
                // Include delta computed measure when delta sort is active
                if (sortIsDelta) {
                  result.push({
                    name: measureName + ComparisonDeltaAbsoluteSuffix,
                    comparisonDelta: {
                      measure: measureName,
                    },
                  });
                }
                return result;
              }
              return { name: measureName };
            })
            .flat();

        const mainQuerySort = this.resolveXAxisSort(
          config,
          isMultiMeasure,
          isComparisonActive,
        );

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
          config.metrics_view,
          {
            measures: measuresWithComparison,
            dimensions,
            sort: mainQuerySort ? [mainQuerySort] : undefined,
            where: combinedWhere,
            timeRange,
            comparisonTimeRange:
              showTimeComparison &&
              comparisonTimeRange?.start &&
              comparisonTimeRange?.end
                ? comparisonTimeRange
                : undefined,
            fillMissing: config.x?.type === "temporal",
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

  getChartDomainValues(measures: MetricsViewSpecMeasure[]): ChartDomainValues {
    const config = get(this.spec);
    const result: Record<string, string[] | undefined> = {};

    if (config.x?.field) {
      result[config.x.field] =
        this.customSortXItems.length > 0
          ? [...this.customSortXItems]
          : undefined;
    }

    if (isFieldConfig(config.color)) {
      if (isMultiFieldConfig(config.y)) {
        result[config.color.field] = this.getMeasureLabels(measures);
      } else {
        result[config.color.field] =
          this.customColorValues.length > 0
            ? [...this.customColorValues]
            : undefined;
      }
    }

    return result;
  }

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const isMultiMeasure = isMultiFieldConfig(config.y);

    if (isMultiMeasure) {
      const xLabel = config.x?.field
        ? fields[config.x.field]?.displayName || config.x.field
        : "";
      const measuresLabel = (config.y?.fields || [])
        .map((m) => fields[m]?.displayName || m)
        .join(", ");
      const preposition = xLabel === "Time" ? "over" : "by";
      return `${measuresLabel} ${preposition} ${xLabel}`;
    } else {
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
}
