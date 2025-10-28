import type {
  ChartDataQuery,
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

export type FunnelMode = "width" | "order";
export type FunnelColorMode = "stage" | "measure" | "name" | "value";
export type FunnelBreakdownMode = "dimension" | "measures";

export type FunnelChartSpec = {
  metrics_view: string;
  breakdownMode?: FunnelBreakdownMode;
  measure?: FieldConfig<"quantitative">;
  stage?: FieldConfig<"nominal">;
  mode?: FunnelMode;
  color?: FunnelColorMode;
};

export type FunnelChartDefaultOptions = {
  stageLimit?: number;
  sort?: ChartSortDirection;
};

const DEFAULT_STAGE_LIMIT = 15;
const DEFAULT_SORT = "-y" as ChartSortDirection;

export class FunnelChartProvider {
  private spec: Readable<FunnelChartSpec>;
  defaultStageLimit = DEFAULT_STAGE_LIMIT;
  defaultSort = DEFAULT_SORT;

  customSortStageItems: string[] = [];

  combinedWhere: Writable<V1Expression | undefined> = writable(undefined);

  constructor(
    spec: Readable<FunnelChartSpec>,
    defaultOptions?: FunnelChartDefaultOptions,
  ) {
    this.spec = spec;
    if (defaultOptions) {
      this.defaultStageLimit = defaultOptions.stageLimit || DEFAULT_STAGE_LIMIT;
      this.defaultSort = defaultOptions.sort || DEFAULT_SORT;
    }
  }

  private getMultiMeasures(measure: FieldConfig | undefined): string[] {
    if (measure?.fields?.length) {
      return measure.fields;
    } else if (measure?.field) {
      return [measure.field];
    }
    return [];
  }

  createChartDataQuery(
    runtime: Writable<Runtime>,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.spec);
    const isMultiMeasure = config.breakdownMode === "measures";

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (isMultiMeasure) {
      const measuresSet = new Set(config.measure?.fields);
      if (config.measure?.type === "quantitative" && config.measure?.field) {
        measuresSet.add(config.measure.field);
      }
      measures = Array.from(measuresSet).map((name) => ({ name }));
    } else {
      if (config.measure?.field) {
        measures = [{ name: config.measure.field }];
      }
    }

    let stageSort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;
    const stageDimensionName = isMultiMeasure ? undefined : config.stage?.field;

    if (!isMultiMeasure && config.stage?.field) {
      limit = config.stage.limit ?? this.defaultStageLimit;
      dimensions = [{ name: config.stage.field }];

      let sort = config.stage.sort;
      if (!sort || Array.isArray(sort)) {
        sort = this.defaultSort;
      }

      if (typeof sort === "string" && config.measure?.field) {
        stageSort = {
          name: config.measure.field,
          desc: sort !== "y",
        };
      }
    }

    // Create topN query for stage dimension
    const topNStageQueryOptionsStore = derived(
      [runtime, timeAndFilterStore],
      ([$runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const instanceId = $runtime.instanceId;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!stageDimensionName &&
          !isMultiMeasure &&
          !Array.isArray(config.stage?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.stage);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          instanceId,
          config.metrics_view,
          {
            measures,
            dimensions: [{ name: stageDimensionName }],
            sort: stageSort ? [stageSort] : undefined,
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

    const topNStageQuery = createQuery(topNStageQueryOptionsStore);

    const queryOptionsStore = derived(
      [runtime, timeAndFilterStore, topNStageQuery],
      ([$runtime, $timeAndFilterStore, $topNStageQuery]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const topNStageData = $topNStageQuery?.data?.data;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          (isMultiMeasure || !!dimensions?.length) &&
          (!isMultiMeasure &&
          !Array.isArray(config.stage?.sort) &&
          stageDimensionName
            ? topNStageData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          isMultiMeasure ? undefined : config.stage,
        );

        let includedStageValues: string[] = [];

        // Apply topN filter for stage dimension (only in dimension mode)
        if (!isMultiMeasure) {
          if (Array.isArray(config.stage?.sort)) {
            includedStageValues = config.stage.sort;
          } else if (topNStageData?.length && stageDimensionName) {
            includedStageValues = topNStageData.map(
              (d) => d[stageDimensionName] as string,
            );
          }

          if (stageDimensionName) {
            this.customSortStageItems = includedStageValues;
            const filterForTopStageValues = createInExpression(
              stageDimensionName,
              includedStageValues,
            );
            combinedWhere = mergeFilters(
              combinedWhere,
              filterForTopStageValues,
            );
          }
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
            sort: stageSort ? [stageSort] : undefined,
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

        return queryOptions;
      },
    );

    const query = createQuery(queryOptionsStore);
    return query;
  }

  getChartDomainValues() {
    return {}; // no-op
  }

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.spec);
    const isMultiMeasure = config.breakdownMode === "measures";

    if (isMultiMeasure) {
      const measuresLabel = this.getMultiMeasures(config.measure)
        .map((m) => fields[m]?.displayName || m)
        .join(", ");
      return `${measuresLabel} funnel`;
    } else {
      const { measure, stage } = config;
      const measureLabel = measure?.field
        ? fields[measure.field]?.displayName || measure.field
        : "";
      const stageLabel = stage?.field
        ? fields[stage.field]?.displayName || stage.field
        : "";

      return stageLabel
        ? `${measureLabel} funnel by ${stageLabel}`
        : measureLabel;
    }
  }
}
