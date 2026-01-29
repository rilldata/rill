import type {
  ChartDataQuery,
  ChartDomainValues,
  ChartFieldsMap,
  ChartSortDirection,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import {
  isFieldConfig,
  isMultiFieldConfig,
} from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationSort,
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

export type CartesianAxisKey = "x" | "y";

export type CartesianAxisRole = "dimension" | "measure";

export interface CartesianAxisRoles {
  dimensionAxis?: CartesianAxisKey;
  measureAxis?: CartesianAxisKey;
}

function isDimensionLikeAxis(field: FieldConfig | undefined): boolean {
  if (!field) return false;
  return field.type === "nominal" || field.type === "temporal";
}

function isMeasureLikeAxis(field: FieldConfig | undefined): boolean {
  if (!field) return false;
  return field.type === "quantitative";
}

export function getAxisRoles(config: CartesianChartSpec): CartesianAxisRoles {
  const x = config.x;
  const y = config.y;

  const xIsMeasure = isMeasureLikeAxis(x);
  const yIsMeasure = isMeasureLikeAxis(y);
  const xIsDimension = isDimensionLikeAxis(x);
  const yIsDimension = isDimensionLikeAxis(y);

  let dimensionAxis: CartesianAxisKey | undefined;
  let measureAxis: CartesianAxisKey | undefined;

  // Prefer a single clear measure axis when possible
  if (xIsMeasure && !yIsMeasure) {
    measureAxis = "x";
    if (yIsDimension) dimensionAxis = "y";
  } else if (!xIsMeasure && yIsMeasure) {
    measureAxis = "y";
    if (xIsDimension) dimensionAxis = "x";
  } else if (xIsMeasure && yIsMeasure) {
    // Both axes are measures – prefer y for backwards compatibility
    measureAxis = "y";
  } else {
    // No clear measure axis – fall back to dimension-like axes
    if (xIsDimension && !yIsDimension) {
      dimensionAxis = "x";
    } else if (!xIsDimension && yIsDimension) {
      dimensionAxis = "y";
    } else if (xIsDimension && yIsDimension) {
      // Both dimension-like – prefer x for backwards compatibility
      dimensionAxis = "x";
    }
  }

  return { dimensionAxis, measureAxis };
}

function getAxisRolesWithMetrics(
  config: CartesianChartSpec,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): CartesianAxisRoles {
  const baseRoles = getAxisRoles(config);
  if (!metricsViewSpec) return baseRoles;

  const measureNames = new Set(
    (metricsViewSpec.measures || []).map((m) => m.name),
  );
  const dimensionNames = new Set(
    (metricsViewSpec.dimensions || []).map((d) => d.name),
  );
  if (metricsViewSpec.timeDimension) {
    dimensionNames.add(metricsViewSpec.timeDimension);
  }

  const xField = config.x?.field;
  const yField = config.y?.field;

  const xIsMeasure = !!(xField && measureNames.has(xField));
  const yIsMeasure = !!(yField && measureNames.has(yField));
  const xIsDimension = !!(xField && dimensionNames.has(xField));
  const yIsDimension = !!(yField && dimensionNames.has(yField));

  let dimensionAxis: CartesianAxisKey | undefined;
  let measureAxis: CartesianAxisKey | undefined;

  // Prefer metrics-view-based classification
  if (xIsMeasure && !yIsMeasure) {
    measureAxis = "x";
  } else if (!xIsMeasure && yIsMeasure) {
    measureAxis = "y";
  } else if (xIsMeasure && yIsMeasure) {
    measureAxis = baseRoles.measureAxis ?? "y";
  }

  if (xIsDimension && !yIsDimension) {
    dimensionAxis = "x";
  } else if (!xIsDimension && yIsDimension) {
    dimensionAxis = "y";
  } else if (xIsDimension && yIsDimension) {
    dimensionAxis = baseRoles.dimensionAxis ?? "x";
  }

  return {
    dimensionAxis: dimensionAxis ?? baseRoles.dimensionAxis,
    measureAxis: measureAxis ?? baseRoles.measureAxis,
  };
}

export type CartesianChartSpec = {
  metrics_view: string;
  /**
   * Both positional axes can be either dimension-like (nominal/temporal)
   * or measure-like (quantitative). The actual "dimension" vs "measure"
   * role is determined dynamically via `getAxisRoles`.
   */
  x?: FieldConfig<"nominal" | "quantitative" | "time">;
  y?: FieldConfig<"nominal" | "quantitative" | "time">;
  color?: FieldConfig<"nominal"> | string;
};

export type CartesianChartDefaultOptions = {
  nominalLimit?: number;
  splitLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = "-y" as ChartSortDirection;

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

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): ChartDataQuery {
    const config = get(this.spec);

    const { dimensionAxis, measureAxis } = getAxisRolesWithMetrics(
      config,
      metricsViewSpec,
    );
    const dimensionFieldConfig =
      dimensionAxis && config[dimensionAxis]
        ? (config[dimensionAxis] as FieldConfig)
        : undefined;
    const measureFieldConfig =
      measureAxis && config[measureAxis]
        ? (config[measureAxis] as FieldConfig)
        : undefined;

    // #region agent log
    fetch("http://127.0.0.1:7242/ingest/9398fc01-29fe-493e-a10c-09e7c6bb4eaf", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "pre-fix-1",
        hypothesisId: "H1",
        location: "CartesianChartProvider.ts:createChartDataQuery:axis-selection",
        message: "Axis roles and field configs before query build",
        data: {
          metricsView: metricsViewSpec?.name,
          chartSpecMetricsView: config.metrics_view,
          x: config.x,
          y: config.y,
          color: config.color,
          dimensionAxis,
          measureAxis,
          dimensionFieldConfig,
          measureFieldConfig,
        },
        timestamp: Date.now(),
      }),
    }).catch(() => {});
    // #endregion agent log

    const isMultiMeasure = isMultiFieldConfig(measureFieldConfig);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    let measuresSet = new Set<string>();
    if (isMultiMeasure) {
      measuresSet = new Set(measureFieldConfig?.fields);
      if (measureFieldConfig?.type === "quantitative" && measureFieldConfig?.field) {
        measuresSet.add(measureFieldConfig.field);
      }
      measures = Array.from(measuresSet).map((name) => ({ name }));
    } else {
      if (measureFieldConfig?.type === "quantitative" && measureFieldConfig?.field) {
        measuresSet = new Set([measureFieldConfig.field]);
        measures = [{ name: measureFieldConfig.field }];
      }
    }

    let primaryAxisSort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;
    let hasColorDimension = false;
    let colorDimensionName = "";
    let colorLimit: number | undefined;

    const dimensionName = dimensionFieldConfig?.field;

    // #region agent log
    fetch("http://127.0.0.1:7242/ingest/9398fc01-29fe-493e-a10c-09e7c6bb4eaf", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "pre-fix-1",
        hypothesisId: "H2",
        location: "CartesianChartProvider.ts:createChartDataQuery:measures-dimensions",
        message: "Measures and dimensions after classification",
        data: {
          metricsView: metricsViewSpec?.name,
          chartSpecMetricsView: config.metrics_view,
          dimensionAxis,
          measureAxis,
          dimensionFieldConfig,
          measureFieldConfig,
          measures: Array.from(measuresSet),
          isMultiMeasure,
        },
        timestamp: Date.now(),
      }),
    }).catch(() => {});
    // #endregion agent log

    if (
      dimensionFieldConfig &&
      (dimensionFieldConfig.type === "nominal" ||
        dimensionFieldConfig.type === "temporal") &&
      dimensionName
    ) {
      if (dimensionFieldConfig.type === "nominal") {
        limit = dimensionFieldConfig.limit ?? 100;
        if (isMultiMeasure) {
          const sort = dimensionFieldConfig.sort;
          if (sort === "y" || sort === "-y" || sort === "measure" || sort === "-measure") {
            // Use first measure for measure-based sorts
            const firstMeasure = measureFieldConfig?.fields?.[0];
            if (firstMeasure) {
              primaryAxisSort = {
                name: firstMeasure,
                desc: sort.startsWith("-"),
              };
            }
          } else if (sort === "x" || sort === "-x") {
            primaryAxisSort = {
              name: dimensionName,
              desc: sort === "-x",
            };
          }
        } else {
          // When we have a clear measure axis, allow vegaSortToAggregationSort
          // to resolve measure-based sorts using the first measure field.
          const primaryEncoder: "x" | "y" =
            dimensionAxis === "y" ? "y" : "x";
          const firstMeasureField =
            measures.length > 0 ? measures[0].name : undefined;
          primaryAxisSort = vegaSortToAggregationSort(
            primaryEncoder,
            config,
            this.defaultSort,
            firstMeasureField,
          );
        }
      }

      dimensions = [{ name: dimensionName }];
    }

    if (isFieldConfig(config.color) && !isMultiMeasure && config.color.field) {
      colorDimensionName = config.color.field;
      colorLimit = config.color.limit ?? this.defaultSplitLimit;
      dimensions = [...dimensions, { name: colorDimensionName }];
      hasColorDimension = true;
    }

    // Create topN query for x dimension
    const topNXQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
          dimensionFieldConfig?.type === "nominal" &&
          !Array.isArray(dimensionFieldConfig?.sort) &&
          !!dimensionName;

        const topNWhere = getFilterWithNullHandling(where, dimensionFieldConfig);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: dimensionName }],
            sort: primaryAxisSort ? [primaryAxisSort] : undefined,
            where: topNWhere,
            timeRange: hasTimeSeries ? timeRange : undefined,
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
            sort: measureFieldConfig?.field
              ? [{ name: measureFieldConfig.field, desc: true }]
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
          dimensionFieldConfig?.type === "nominal" &&
          !Array.isArray(dimensionFieldConfig?.sort)
            ? topNXData !== undefined
            : true) &&
          (hasColorDimension && colorDimensionName && colorLimit
            ? topNColorData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          dimensionFieldConfig,
        );

        let includedXValues: string[] = [];

        // Apply topN filter for primary (dimension) axis
        if (Array.isArray(dimensionFieldConfig?.sort)) {
          includedXValues = dimensionFieldConfig.sort;
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
        if (dimensionFieldConfig?.type === "temporal" && timeGrain) {
          dimensions = dimensions.map((d) =>
            d.name === dimensionName ? { ...d, timeGrain } : d,
          );
        }

        const measuresWithComparison: V1MetricsViewAggregationMeasure[] =
          Array.from(measuresSet)
            .map((measureName) => {
              if (showTimeComparison && comparisonTimeRange?.start) {
                return [
                  { name: measureName },
                  {
                    name: measureName + ComparisonDeltaPreviousSuffix,
                    comparisonValue: {
                      measure: measureName,
                    },
                  },
                ];
              }
              return { name: measureName };
            })
            .flat();

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
          config.metrics_view,
          {
            measures: measuresWithComparison,
            dimensions,
            sort: primaryAxisSort ? [primaryAxisSort] : undefined,
            where: combinedWhere,
            timeRange,
            comparisonTimeRange:
              showTimeComparison &&
              comparisonTimeRange?.start &&
              comparisonTimeRange?.end
                ? comparisonTimeRange
                : undefined,
            fillMissing: dimensionFieldConfig?.type === "temporal",
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

    const { dimensionAxis } = getAxisRoles(config);
    const dimensionFieldConfig =
      dimensionAxis && config[dimensionAxis]
        ? (config[dimensionAxis] as FieldConfig)
        : undefined;

    if (dimensionFieldConfig?.field) {
      result[dimensionFieldConfig.field] =
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
    const { dimensionAxis, measureAxis } = getAxisRoles(config);
    const dimensionFieldConfig =
      dimensionAxis && config[dimensionAxis]
        ? (config[dimensionAxis] as FieldConfig)
        : undefined;
    const measureFieldConfig =
      measureAxis && config[measureAxis]
        ? (config[measureAxis] as FieldConfig)
        : undefined;

    const isMultiMeasure = isMultiFieldConfig(measureFieldConfig);

    if (isMultiMeasure) {
      const dimensionLabel = dimensionFieldConfig?.field
        ? fields[dimensionFieldConfig.field]?.displayName ||
          dimensionFieldConfig.field
        : "";
      const measuresLabel = (measureFieldConfig?.fields || [])
        .map((m) => fields[m]?.displayName || m)
        .join(", ");
      const preposition = dimensionLabel === "Time" ? "over" : "by";
      return `${measuresLabel} ${preposition} ${dimensionLabel}`;
    } else {
      const { color } = config;

      const dimensionLabel = dimensionFieldConfig?.field
        ? fields[dimensionFieldConfig.field]?.displayName ||
          dimensionFieldConfig.field
        : "";
      const measureLabel = measureFieldConfig?.field
        ? fields[measureFieldConfig.field]?.displayName ||
          measureFieldConfig.field
        : "";

      const colorLabel =
        typeof color === "object" && color?.field
          ? fields[color.field]?.displayName || color.field
          : "";

      const preposition = dimensionLabel === "Time" ? "over" : "per";

      return colorLabel
        ? `${measureLabel} ${preposition} ${dimensionLabel} split by ${colorLabel}`
        : `${measureLabel} ${preposition} ${dimensionLabel}`;
    }
  }
}
