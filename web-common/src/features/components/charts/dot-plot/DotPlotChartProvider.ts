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
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
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

export type DotPlotChartSpec = {
  metrics_view: string;
  y?: FieldConfig<"nominal">;
  x?: FieldConfig<"quantitative">;
  detail?: FieldConfig<"nominal">;
  color?: FieldConfig<"nominal"> | string;
  jitter?: boolean;
};

export type DotPlotChartDefaultOptions = {
  nominalLimit?: number;
  splitLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = "-x" as ChartSortDirection;

export class DotPlotChartProvider {
  private spec: Readable<DotPlotChartSpec>;
  defaultNominalLimit = DEFAULT_NOMINAL_LIMIT;
  defaultSplitLimit = DEFAULT_SPLIT_LIMIT;
  defaultSort = DEFAULT_SORT;

  customSortYItems: string[] = [];
  customColorValues: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<DotPlotChartSpec>,
    defaultOptions?: DotPlotChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultNominalLimit =
        defaultOptions.nominalLimit || DEFAULT_NOMINAL_LIMIT;
      this.defaultSplitLimit = defaultOptions.splitLimit || DEFAULT_SPLIT_LIMIT;
      this.defaultSort = defaultOptions.sort || DEFAULT_SORT;
    }
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);

    const measures: V1MetricsViewAggregationMeasure[] = [];
    const dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.x?.type === "quantitative" && config.x?.field) {
      measures.push({ name: config.x.field });
    }

    const yDimensionName = config.y?.field;
    if (config.y?.type === "nominal" && yDimensionName) {
      dimensions.push({ name: yDimensionName });
    }

    const detailDimensionName = config.detail?.field;
    if (config.detail?.type === "nominal" && detailDimensionName) {
      dimensions.push({ name: detailDimensionName });
    }

    let hasColorDimension = false;
    let colorDimensionName = "";
    let colorLimit: number | undefined;

    if (isFieldConfig(config.color)) {
      colorDimensionName = config.color.field;
      colorLimit = config.color.limit ?? this.defaultSplitLimit;
      dimensions.push({ name: colorDimensionName });
      hasColorDimension = true;
    }

    let yAxisSort: V1MetricsViewAggregationSort | undefined;
    const limit = config.y?.limit ?? this.defaultNominalLimit;

    if (config.y?.type === "nominal" && yDimensionName) {
      yAxisSort = vegaSortToAggregationSort("y", config, this.defaultSort);
    }

    const topNYQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          config.y?.type === "nominal" &&
          !Array.isArray(config.y?.sort) &&
          !!yDimensionName;

        const topNWhere = getFilterWithNullHandling(where, config.y);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: yDimensionName }],
            sort: yAxisSort ? [yAxisSort] : undefined,
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

    const topNYQuery = createQuery(topNYQueryOptionsStore);

    const topNColorQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
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
          $runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: colorDimensionName }],
            sort: config?.x?.field
              ? [{ name: config.x.field, desc: true }]
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
      [runtime, timeAndFilterStore, topNYQuery, topNColorQuery],
      ([$runtime, $timeAndFilterStore, $topNYQuery, $topNColorQuery]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const topNYData = $topNYQuery?.data?.data;
        const topNColorData = $topNColorQuery?.data?.data;

        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          !!dimensions?.length &&
          (config.y?.type === "nominal" &&
          !Array.isArray(config.y?.sort) &&
          yDimensionName
            ? topNYData !== undefined
            : true) &&
          (hasColorDimension && colorDimensionName && colorLimit
            ? topNColorData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.y,
        );

        let includedYValues: string[] = [];

        if (Array.isArray(config.y?.sort)) {
          includedYValues = config.y.sort;
        } else if (topNYData?.length && yDimensionName) {
          includedYValues = topNYData.map((d) => d[yDimensionName] as string);
        }

        if (yDimensionName) {
          this.customSortYItems = includedYValues;
          const filterForTopYValues = createInExpression(
            yDimensionName,
            includedYValues,
          );
          combinedWhere = mergeFilters(combinedWhere, filterForTopYValues);
        }

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

        this.combinedWhere.set(combinedWhere);

        const hasDetailDimension = !!detailDimensionName;
        let queryLimit: string;
        if (hasDetailDimension || hasColorDimension) {
          queryLimit = "5000";
        } else if (!limit) {
          queryLimit = "5000";
        } else {
          queryLimit = limit.toString();
        }

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            sort: yAxisSort ? [yAxisSort] : undefined,
            where: combinedWhere,
            timeRange,
            limit: queryLimit,
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

  getChartDomainValues(): ChartDomainValues {
    const config = get(this.spec);
    const result: Record<string, string[] | undefined> = {};

    if (config.y?.field) {
      result[config.y.field] =
        this.customSortYItems.length > 0
          ? [...this.customSortYItems]
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

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const { x, y, color, detail } = config;
    const xLabel = x?.field ? fields[x.field]?.displayName || x.field : "";
    const yLabel = y?.field ? fields[y.field]?.displayName || y.field : "";

    const colorLabel =
      typeof color === "object" && color?.field
        ? fields[color.field]?.displayName || color.field
        : "";

    const detailLabel = detail?.field
      ? fields[detail.field]?.displayName || detail.field
      : "";

    if (colorLabel && detailLabel) {
      return `${xLabel} by ${yLabel} and ${detailLabel} (colored by ${colorLabel})`;
    } else if (colorLabel) {
      return `${xLabel} by ${yLabel} (colored by ${colorLabel})`;
    } else if (detailLabel) {
      return `${xLabel} by ${yLabel} and ${detailLabel}`;
    } else {
      return `${xLabel} by ${yLabel}`;
    }
  }
}
