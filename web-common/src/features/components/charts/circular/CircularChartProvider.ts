import type {
  ChartDataQuery,
  ChartDomainValues,
  ChartFieldsMap,
  ChartSortDirection,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
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
import { getFilterWithNullHandling } from "../query-util";

export type CircularChartSpec = {
  metrics_view: string;
  measure?: FieldConfig<"quantitative">;
  color?: FieldConfig<"nominal">;
  innerRadius?: number;
};

export type CircularChartDefaultOptions = {
  colorLimit?: number;
  colorSort?: ChartSortDirection;
};

const DEFAULT_COLOR_LIMIT = 20;
const DEFAULT_SORT = "-measure" as ChartSortDirection;

export class CircularChartProvider {
  private spec: Readable<CircularChartSpec>;
  defaultColorLimit = DEFAULT_COLOR_LIMIT;
  defaultColorSort = DEFAULT_SORT;

  customColorValues: string[] = [];
  totalsValue: number | undefined = undefined;

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<CircularChartSpec>,
    defaultOptions?: CircularChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultColorLimit = defaultOptions.colorLimit || DEFAULT_COLOR_LIMIT;
      this.defaultColorSort = defaultOptions.colorSort || DEFAULT_SORT;
    }
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);

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
      limit = config.color?.limit || this.defaultColorLimit;
      dimensions = [{ name: colorDimensionName }];
      colorSort = this.getColorSort(config);
    }

    // Create topN query for color dimension
    const topNColorQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!colorDimensionName &&
          config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
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
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!showTotal &&
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!config.measure?.field;

        const totalWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
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
      [runtime, timeAndFilterStore, topNColorQuery, totalQuery],
      ([$runtime, $timeAndFilterStore, $topNColorQuery, $totalQuery]) => {
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

        // Store combinedWhere for use in BaseChart
        this.combinedWhere.set(combinedWhere);

        const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
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
      sort = this.defaultColorSort;
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

  getChartDomainValues(): ChartDomainValues {
    const config = get(this.spec);
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

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const { measure, color } = config;
    const measureLabel = measure?.field
      ? fields[measure.field]?.displayName || measure.field
      : "";
    const colorLabel = color?.field
      ? fields[color.field]?.displayName || color.field
      : "";

    return colorLabel ? `${measureLabel} split by ${colorLabel}` : measureLabel;
  }
}
