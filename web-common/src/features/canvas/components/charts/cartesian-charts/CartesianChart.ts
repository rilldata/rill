import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationResponse,
  type V1MetricsViewAggregationSort,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  keepPreviousData,
  type CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived, get, readable, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";
import type { ChartDataQuery, ChartSortDirection, FieldConfig } from "../types";

export type CartesianChartSpec = BaseChartConfig & {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig | string;
};

export class CartesianChartComponent extends BaseChart<CartesianChartSpec> {
  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);
  }

  protected getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return {
      x: { type: "positional", label: "X-axis" },
      y: { type: "positional", label: "Y-axis" },
      color: { type: "mark", label: "Color", meta: { type: "color" } },
    };
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    const config = get(this.specStore);

    let measures: V1MetricsViewAggregationMeasure[] = [];
    let dimensions: V1MetricsViewAggregationDimension[] = [];

    if (config.y?.type === "quantitative" && config.y?.field) {
      measures = [{ name: config.y?.field }];
    }

    let sort: V1MetricsViewAggregationSort | undefined;
    let limit: number | undefined;
    let hasColorDimension = false;

    return derived(
      [ctx.runtime, timeAndFilterStore],
      ([runtime, $timeAndFilterStore], set) => {
        const { timeRange, where, timeGrain } = $timeAndFilterStore;

        let outerWhere = where;

        if (config.x?.type === "nominal" && config.x?.field) {
          limit = config.x.limit;
          sort = this.vegaSortToAggregationSort(config.x?.sort, config);
          dimensions = [{ name: config.x?.field }];

          const showNull = !!config.x.showNull;
          if (!showNull) {
            const excludeNullFilter = createInExpression(
              config.x?.field,
              [null],
              true,
            );
            outerWhere = mergeFilters(where, excludeNullFilter);
          }
        } else if (config.x?.type === "temporal" && timeGrain) {
          dimensions = [{ name: config.x?.field, timeGrain }];
        }

        if (typeof config.color === "object" && config.color?.field) {
          dimensions = [...dimensions, { name: config.color.field }];
          hasColorDimension = true;
        }

        let topNQuery:
          | Readable<null>
          | CreateQueryResult<V1MetricsViewAggregationResponse, HTTPError> =
          readable(null);

        const enabled = !!timeRange?.start && !!timeRange?.end;

        if (limit && hasColorDimension) {
          topNQuery = createQueryServiceMetricsViewAggregation(
            runtime.instanceId,
            config.metrics_view,
            {
              measures,
              dimensions: [{ name: config.x?.field }],
              sort: sort ? [sort] : undefined,
              where: outerWhere,
              timeRange,
              limit: limit.toString(),
            },
            {
              query: {
                enabled,
                placeholderData: keepPreviousData,
              },
            },
            ctx.queryClient,
          );
        }

        return derived(topNQuery, ($topNQuery, topNSet) => {
          if ($topNQuery !== null && !$topNQuery?.data) {
            return topNSet({
              isFetching: $topNQuery.isFetching,
              error: $topNQuery.error,
              data: undefined,
            });
          }

          const dimensionName = config.x?.field;

          let combinedWhere: V1Expression | undefined = outerWhere;
          if ($topNQuery?.data?.data?.length && dimensionName) {
            const topValues = $topNQuery?.data?.data.map(
              (d) => d[dimensionName] as string,
            );
            const filterForTopValues = createInExpression(
              dimensionName,
              topValues,
            );

            combinedWhere = mergeFilters(where, filterForTopValues);
          }

          const dataQuery = createQueryServiceMetricsViewAggregation(
            runtime.instanceId,
            config.metrics_view,
            {
              measures,
              dimensions,
              sort: sort ? [sort] : undefined,
              where: combinedWhere,
              timeRange,
              limit: hasColorDimension || !limit ? "5000" : limit.toString(),
            },
            {
              query: {
                enabled,
                placeholderData: keepPreviousData,
              },
            },
            ctx.queryClient,
          );

          return derived(dataQuery, ($dataQuery) => {
            return {
              isFetching: $dataQuery.isFetching,
              error: $dataQuery.error,
              data: $dataQuery?.data?.data,
            };
          }).subscribe(topNSet);
        }).subscribe(set);
      },
    );
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): CartesianChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    return {
      metrics_view: metricsViewName,
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: "-y",
        limit: 20,
      },
      y: {
        type: "quantitative",
        field: randomMeasure,
        zeroBasedOrigin: true,
      },
    };
  }

  protected vegaSortToAggregationSort(
    sort: ChartSortDirection | undefined,
    config: CartesianChartSpec,
  ): V1MetricsViewAggregationSort | undefined {
    if (!sort) return undefined;
    const field =
      sort === "x" || sort === "-x" ? config.x?.field : config.y?.field;
    if (!field) return undefined;

    return {
      name: field,
      desc: sort === "-x" || sort === "-y",
    };
  }
}
