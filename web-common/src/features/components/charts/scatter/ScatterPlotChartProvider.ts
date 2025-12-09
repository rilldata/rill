import type {
  ChartDataQuery,
  ChartDomainValues,
  ChartFieldsMap,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1Expression,
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

export type ScatterPlotChartSpec = {
  metrics_view: string;
  x?: FieldConfig<"quantitative">;
  y?: FieldConfig<"quantitative">;
  dimension?: FieldConfig<"nominal">;
  size?: FieldConfig<"quantitative">;
  color?: FieldConfig<"nominal"> | string;
};

const DEFAULT_SPLIT_LIMIT = 10;

export class ScatterPlotChartProvider {
  private spec: Readable<ScatterPlotChartSpec>;
  defaultSplitLimit = DEFAULT_SPLIT_LIMIT;

  customColorValues: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(spec: Readable<ScatterPlotChartSpec>) {
    this.spec = spec;
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
    if (config.y?.type === "quantitative" && config.y?.field) {
      measures.push({ name: config.y.field });
    }
    if (config.size?.type === "quantitative" && config.size?.field) {
      measures.push({ name: config.size.field });
    }

    if (config.dimension?.type === "nominal" && config.dimension?.field) {
      dimensions.push({ name: config.dimension.field });
    }

    let hasColorDimension = false;
    let colorDimensionName = "";
    let colorLimit: number | undefined;

    if (isFieldConfig(config.color) && config.color?.field) {
      colorDimensionName = config.color.field;
      colorLimit = config.color.limit ?? this.defaultSplitLimit;
      dimensions.push({ name: colorDimensionName });
      hasColorDimension = true;
    }

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
          isFieldConfig(config.color) ? config.color : undefined,
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
      [runtime, timeAndFilterStore, topNColorQuery],
      ([$runtime, $timeAndFilterStore, $topNColorQuery]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const topNColorData = $topNColorQuery?.data?.data;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          !!dimensions?.length &&
          (hasColorDimension && colorDimensionName && colorLimit
            ? topNColorData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.dimension,
        );

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

        return getQueryServiceMetricsViewAggregationQueryOptions(
          $runtime.instanceId,
          config.metrics_view,
          {
            measures,
            dimensions,
            where: combinedWhere,
            timeRange,
            limit: "9999",
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
    const xField = fields[config.x?.field || ""];
    const yField = fields[config.y?.field || ""];
    const dimensionField = fields[config.dimension?.field || ""];
    const sizeField = fields[config.size?.field || ""];

    const xTitle = xField?.displayName || config.x?.field || "X";
    const yTitle = yField?.displayName || config.y?.field || "Y";
    const dimensionTitle =
      dimensionField?.displayName || config.dimension?.field || "";
    const sizeTitle = sizeField?.displayName || config.size?.field || "";

    let title = `${xTitle} vs ${yTitle}`;

    if (dimensionTitle) {
      title += ` for ${dimensionTitle}`;
    }

    if (sizeTitle) {
      title += ` sized by ${sizeTitle}`;
    }

    return title;
  }
}
