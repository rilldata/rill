import { getFilterWithNullHandling } from "@rilldata/web-common/features/canvas/components/charts/query-utils";
import type {
  ChartFieldsMap,
  FieldConfig,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import type { ChartDataQuery } from "../types";

export type FunnelMode = "width" | "order";
export type FunnelColorMode = "stage" | "measure";

type FunnelChartEncoding = {
  measure?: FieldConfig;
  stage?: FieldConfig;
  mode?: FunnelMode;
  color?: FunnelColorMode;
};

const DEFAULT_STAGE_LIMIT = 15;
const DEFAULT_SORT = "-y";

export type FunnelChartSpec = BaseChartConfig & FunnelChartEncoding;

export class FunnelChartComponent extends BaseChart<FunnelChartSpec> {
  customSortStageItems: string[] = [];

  static chartInputParams: Record<string, ComponentInputParam> = {
    stage: {
      type: "positional",
      label: "Stage",
      meta: {
        chartFieldInput: {
          type: "dimension",
          nullSelector: true,
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["y", "-y", "custom"],
          },
          limitSelector: { defaultLimit: DEFAULT_STAGE_LIMIT },
          hideTimeDimension: true,
        },
      },
    },
    measure: {
      type: "positional",
      label: "Measure",
      meta: {
        chartFieldInput: {
          type: "measure",
        },
      },
    },
    mode: {
      type: "select",
      label: "Mode",
      meta: {
        default: "width",
        options: [
          { label: "Width", value: "width" },
          { label: "Order", value: "order" },
        ],
      },
    },
    color: {
      type: "select",
      label: "Color",
      meta: {
        default: "stage",
        options: [
          { label: "Stage", value: "stage" },
          { label: "Measure", value: "measure" },
        ],
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = FunnelChartComponent.chartInputParams;
    const sortSelector = inputParams.stage.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.customSortStageItems;
    }
    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.measure?.field) {
      measures = [{ name: config.measure.field }];
    }

    let stageSort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;
    const stageDimensionName = config.stage?.field;

    if (config.stage?.field) {
      limit = config.stage.limit ?? DEFAULT_STAGE_LIMIT;
      dimensions = [{ name: config.stage.field }];

      let sort = config.stage.sort;
      if (!sort || Array.isArray(sort)) {
        sort = DEFAULT_SORT;
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
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!stageDimensionName &&
          !Array.isArray(config.stage?.sort);

        const topNWhere = getFilterWithNullHandling(where, config.stage);

        return getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
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
      [ctx.runtime, timeAndFilterStore, topNStageQuery],
      ([runtime, $timeAndFilterStore, $topNStageQuery]) => {
        const { timeRange, where } = $timeAndFilterStore;
        const topNStageData = $topNStageQuery?.data?.data;
        const enabled =
          !!timeRange?.start &&
          !!timeRange?.end &&
          !!measures?.length &&
          !!dimensions?.length &&
          (!Array.isArray(config.stage?.sort) && stageDimensionName
            ? topNStageData !== undefined
            : true);

        let combinedWhere: V1Expression | undefined = getFilterWithNullHandling(
          where,
          config.stage,
        );

        let includedStageValues: string[] = [];

        // Apply topN filter for stage dimension
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
          combinedWhere = mergeFilters(combinedWhere, filterForTopStageValues);
        }

        this.combinedWhere = combinedWhere;

        const queryOptions = getQueryServiceMetricsViewAggregationQueryOptions(
          runtime.instanceId,
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

  chartTitle(fields: ChartFieldsMap) {
    const config = get(this.specStore);
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

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): FunnelChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    const randomDimension = dimensions[
      Math.floor(Math.random() * dimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      stage: {
        type: "nominal",
        field: randomDimension,
        limit: DEFAULT_STAGE_LIMIT,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
      },
      mode: "width",
      color: "stage",
    };
  }
}
