import {
  ChartSortType,
  type ChartDataQuery,
  type ChartDomainValues,
  type ChartFieldsMap,
  type ChartSortDirection,
  type FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type {
  V1Expression,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponseDataItem,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import { getQueryServiceMetricsViewAggregationQueryOptions } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import {
  derived,
  get,
  readable,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import { getFilterWithNullHandling } from "../query-util";
import {
  computeOtherGrouping,
  OTHER_SLICE_LABEL,
  type OtherGroupResult,
} from "./other-grouping";

export type CircularChartSpec = {
  metrics_view: string;
  measure?: FieldConfig<"quantitative">;
  color?: FieldConfig<"nominal">;
  innerRadius?: number;
  showOther?: boolean;
};

export type CircularChartDefaultOptions = {
  colorLimit?: number;
  colorSort?: ChartSortDirection;
};

const DEFAULT_COLOR_LIMIT = 20;
const DEFAULT_SORT = ChartSortType.MEASURE_DESC as ChartSortDirection;

export class CircularChartProvider {
  private spec: Readable<CircularChartSpec>;
  defaultColorLimit = DEFAULT_COLOR_LIMIT;
  defaultColorSort = DEFAULT_SORT;

  customColorValues: string[] = [];
  totalsValue: number | undefined = undefined;
  otherGroupResult: OtherGroupResult | undefined = undefined;

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
    client: RuntimeClient,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
    visible?: Readable<boolean>,
  ): ChartDataQuery {
    const visibleStore = visible ?? readable(true);
    const config = get(this.spec);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.measure?.field) {
      measures = [{ name: config.measure.field }];
    }

    let colorSort: V1MetricsViewAggregationSort | undefined;
    let queryLimit: number = this.defaultColorLimit;
    const colorDimensionName = config.color?.field;
    const showTotal = config.measure?.showTotal;
    const userLimit = config.color?.limit;

    if (colorDimensionName) {
      const showOtherEnabled = config.showOther !== false;
      if (showOtherEnabled) {
        queryLimit = Math.max(
          userLimit ?? this.defaultColorLimit,
          this.defaultColorLimit,
        );
      } else {
        queryLimit = userLimit || this.defaultColorLimit;
      }
      dimensions = [{ name: colorDimensionName }];
      colorSort = this.getColorSort(config);
    }

    // Create topN query for color dimension
    const topNColorQueryOptionsStore = derived(
      [timeAndFilterStore, visibleStore],
      ([$timeAndFilterStore, $visible]) => {
        const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;
        const enabled =
          $visible &&
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
          !!colorDimensionName &&
          config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          client,
          {
            metricsView: config.metrics_view,
            measures,
            dimensions: [{ name: colorDimensionName }],
            sort: colorSort ? [colorSort] : undefined,
            where: topNWhere,
            timeRange,
            limit: queryLimit?.toString(),
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

    const showOther = config.showOther !== false;
    const needsTotal = showTotal || showOther;

    const totalQueryOptionsStore = derived(
      [timeAndFilterStore, visibleStore],
      ([$timeAndFilterStore, $visible]) => {
        const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;
        const enabled =
          $visible &&
          !!needsTotal &&
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
          !!config.measure?.field;

        const totalWhere = getFilterWithNullHandling(where, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          client,
          {
            metricsView: config.metrics_view,
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
      [timeAndFilterStore, topNColorQuery, totalQuery, visibleStore],
      ([$timeAndFilterStore, $topNColorQuery, $totalQuery, $visible]) => {
        const { timeRange, where, hasTimeSeries } = $timeAndFilterStore;
        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          $visible &&
          (!hasTimeSeries || (!!timeRange?.start && !!timeRange?.end)) &&
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
          client,
          {
            metricsView: config.metrics_view,
            measures,
            dimensions,
            where: combinedWhere,
            sort: colorSort ? [colorSort] : undefined,
            timeRange,
            limit: queryLimit?.toString(),
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
      case ChartSortType.COLOR_ASC:
      case ChartSortType.COLOR_DESC:
        field = config.color.field;
        desc = sort === ChartSortType.COLOR_DESC;
        break;
      case ChartSortType.MEASURE_ASC:
      case ChartSortType.MEASURE_DESC:
        field = config.measure?.field;
        desc = sort === ChartSortType.MEASURE_DESC;
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

  /**
   * Transforms raw query data to apply "Other" grouping for pie/donut charts.
   * Called by the data provider pipeline before data reaches the chart spec.
   */
  transformData(
    data: V1MetricsViewAggregationResponseDataItem[],
  ): V1MetricsViewAggregationResponseDataItem[] {
    const config = get(this.spec);
    const measureField = config.measure?.field;
    const colorField = config.color?.field;

    if (!measureField || !colorField) {
      this.otherGroupResult = undefined;
      return data;
    }

    const showOther = config.showOther !== false;
    const userLimit = config.color?.limit;
    const isExplicitLimit =
      userLimit !== undefined && userLimit !== this.defaultColorLimit;

    const result = computeOtherGrouping(data, measureField, colorField, {
      limit: isExplicitLimit ? userLimit : undefined,
      showOther,
      grandTotal: this.totalsValue,
    });

    this.otherGroupResult = result;

    if (result.hasOther) {
      this.customColorValues = result.visibleData
        .map((d) => String(d[colorField] ?? ""))
        .filter((v) => v !== OTHER_SLICE_LABEL);
      this.customColorValues.push(OTHER_SLICE_LABEL);
    }

    if (this.totalsValue === undefined && result.total > 0) {
      this.totalsValue = result.total;
    }

    return result.visibleData;
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

    if (this.otherGroupResult) {
      result["__otherTotal"] = [this.otherGroupResult.total];
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
