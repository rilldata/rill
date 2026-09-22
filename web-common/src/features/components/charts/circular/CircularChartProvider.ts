import type { EphemeralMeasureSpec } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
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
import type {
  V1Expression,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
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
import { withEphemeralMeasures } from "../ephemeral-measures";
import {
  canQueryWithTimeRange,
  getFilterWithNullHandling,
} from "../query-util";
import {
  OTHER_VALUE,
  OTHER_VALUE_DOMAIN_KEY,
  TOTAL_DOMAIN_KEY,
  type LabelsConfig,
} from "./constants";
import type { ExpressionState } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

export type CircularChartSpec = {
  metrics_view: string;
  // Ad-hoc measures derived from existing measures via an arithmetic
  // expression; measure fields may name them.
  adhoc_measures?: EphemeralMeasureSpec[];
  measure?: FieldConfig<"quantitative">;
  color?: FieldConfig<"nominal">;
  innerRadius?: number;
  show_other?: boolean;
  labels?: LabelsConfig;
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
  otherValue: number | undefined = undefined;

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);
  isTruncated: Writable<boolean> = writable(false);

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
    filterStore: Readable<ExpressionState>,
    timeControlStore: Readable<TimeControlState>,
    visible?: Readable<boolean>,
  ): ChartDataQuery {
    const visibleStore = visible ?? readable(true);
    const config = get(this.spec);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.measure?.field) {
      measures = withEphemeralMeasures(config, [
        { name: config.measure.field },
      ]);
    }

    let colorSort: V1MetricsViewAggregationSort | undefined;
    let limit: number;
    const colorDimensionName = config.color?.field;
    const showOther = !!colorDimensionName && config.show_other === true;

    if (colorDimensionName) {
      limit = config.color?.limit ?? this.defaultColorLimit;
      dimensions = [{ name: colorDimensionName }];
      colorSort = this.getColorSort(config);
    }

    // Create topN query for color dimension
    const topNColorQueryOptionsStore = derived(
      [filterStore, timeControlStore, visibleStore],
      ([$filterStore, $timeControlStore, $visible]) => {
        const { expr } = $filterStore;
        const { apiTimeRange, hasTimeSeries } = $timeControlStore;
        const enabled =
          $visible &&
          canQueryWithTimeRange(hasTimeSeries, apiTimeRange) &&
          !!colorDimensionName &&
          config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort);

        const topNWhere = getFilterWithNullHandling(expr, config.color);

        // drives the "Other" UI affordance
        const probeLimit = limit ? limit + 1 : undefined;

        return getQueryServiceMetricsViewAggregationQueryOptions(
          client,
          {
            metricsView: config.metrics_view,
            measures,
            dimensions: [{ name: colorDimensionName }],
            sort: colorSort ? [colorSort] : undefined,
            where: topNWhere,
            timeRange: apiTimeRange,
            limit: probeLimit?.toString(),
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

    // The total query feeds the optional center-label total AND the
    // percent-of-total tooltip entry
    const totalQueryOptionsStore = derived(
      [filterStore, timeControlStore, visibleStore],
      ([$filterStore, $timeControlStore, $visible]) => {
        const { expr } = $filterStore;
        const { apiTimeRange, hasTimeSeries } = $timeControlStore;
        const enabled =
          $visible &&
          canQueryWithTimeRange(hasTimeSeries, apiTimeRange) &&
          !!config.measure?.field;

        const totalWhere = getFilterWithNullHandling(expr, config.color);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          client,
          {
            metricsView: config.metrics_view,
            measures,
            where: totalWhere,
            timeRange: apiTimeRange,
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

    const otherQueryOptionsStore = derived(
      [filterStore, timeControlStore, topNColorQuery, visibleStore],
      ([$filterStore, $timeControlStore, $topNColorQuery, $visible]) => {
        const { expr } = $filterStore;
        const { apiTimeRange, hasTimeSeries } = $timeControlStore;
        const topNData = $topNColorQuery?.data?.data;
        const customSortValues = Array.isArray(config.color?.sort)
          ? config.color.sort
          : undefined;

        const visibleValues = customSortValues
          ? customSortValues
          : topNData
            ? topNData
                .slice(0, limit)
                .map((d) => d[colorDimensionName!] as string)
            : undefined;

        const enabled =
          $visible &&
          showOther &&
          !!visibleValues &&
          visibleValues.length > 0 &&
          canQueryWithTimeRange(hasTimeSeries, apiTimeRange) &&
          !!config.measure?.field &&
          !!colorDimensionName;

        const baseWhere = getFilterWithNullHandling(expr, config.color);
        let otherWhere = baseWhere;
        if (enabled && colorDimensionName && visibleValues) {
          const notInExpr = createInExpression(
            colorDimensionName,
            visibleValues,
            true,
          );
          otherWhere = mergeFilters(baseWhere, notInExpr);
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          client,
          {
            metricsView: config.metrics_view,
            measures,
            where: otherWhere,
            timeRange: apiTimeRange,
          },
          {
            query: {
              enabled,
            },
          },
        );
      },
    );

    const otherQuery = createQuery(otherQueryOptionsStore);

    const queryOptionsStore = derived(
      [
        filterStore,
        timeControlStore,
        topNColorQuery,
        totalQuery,
        otherQuery,
        visibleStore,
      ],
      ([
        $filterStore,
        $timeControlStore,
        $topNColorQuery,
        $totalQuery,
        $otherQuery,
        $visible,
      ]) => {
        const { expr } = $filterStore;
        const { apiTimeRange, hasTimeSeries } = $timeControlStore;
        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          $visible &&
          canQueryWithTimeRange(hasTimeSeries, apiTimeRange) &&
          !!measures?.length &&
          (config.color?.type === "nominal" &&
          !Array.isArray(config.color?.sort)
            ? topNColorData !== undefined
            : true);

        let combinedWhere = expr;
        let topColorValues: string[] = [];

        // Apply topN filter for color dimension
        if (Array.isArray(config.color?.sort)) {
          topColorValues = config.color.sort;
          this.isTruncated.set(false);
        } else if (topNColorData && colorDimensionName) {
          // topN was queried with limit+1; if we got the extra row back,
          // there are more values than the limit and the chart is truncated
          this.isTruncated.set(topNColorData.length > limit);
          topColorValues = topNColorData
            .slice(0, limit)
            .map((d) => d[colorDimensionName] as string);
        }

        if (colorDimensionName) {
          this.customColorValues = topColorValues;
          const filterForTopColorValues = createInExpression(
            colorDimensionName,
            topColorValues,
          );
          combinedWhere = mergeFilters(expr, filterForTopColorValues);
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
            timeRange: apiTimeRange,
            limit: limit?.toString(),
          },
          {
            query: {
              enabled,
              placeholderData: keepPreviousData,
            },
          },
        );

        if (config.measure?.field) {
          this.totalsValue = $totalQuery?.data?.data?.[0]?.[
            config.measure?.field
          ] as number;

          const otherRaw = $otherQuery?.data?.data?.[0]?.[
            config.measure?.field
          ] as number | null | undefined;
          this.otherValue =
            showOther && typeof otherRaw === "number" ? otherRaw : undefined;
        } else {
          this.otherValue = undefined;
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

  getChartDomainValues(): ChartDomainValues {
    const config = get(this.spec);
    const result: Record<string, string[] | number[] | undefined> = {};

    if (isFieldConfig(config.color)) {
      const baseValues =
        this.customColorValues.length > 0 ? [...this.customColorValues] : [];
      if (this.otherValue !== undefined) {
        baseValues.push(OTHER_VALUE);
      }
      result[config.color.field] =
        baseValues.length > 0 ? baseValues : undefined;
    }

    if (this.totalsValue !== undefined) {
      result[TOTAL_DOMAIN_KEY] = [this.totalsValue];
    }

    if (this.otherValue !== undefined) {
      result[OTHER_VALUE_DOMAIN_KEY] = [this.otherValue];
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
