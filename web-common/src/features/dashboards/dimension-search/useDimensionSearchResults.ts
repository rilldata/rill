import { createLikeExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  MetricsViewSpecDimensionV2,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { StreamingQueryBatch } from "@rilldata/web-common/runtime-client/StreamingQueryBatch";
import { derived } from "svelte/store";

export type DimensionSearchResults = {
  responses: (
    | {
        dimension: string;
        values: any[];
      }
    | undefined
  )[];
  errors: Error[];
  completed: boolean;
  progress: number;
};

const batch = new StreamingQueryBatch(100);

export function useDimensionSearchResults(
  instanceId: string,
  metricsViewName: string,
  metricsView: V1MetricsViewSpec,
  searchText: string,
) {
  const dimensions = metricsView.dimensions ?? [];
  return derived(
    dimensions.map((dimension) =>
      getValuesForDimension(instanceId, metricsViewName, dimension, searchText),
    ),
    (dimensionsValues) => {
      const results: DimensionSearchResults = {
        responses: new Array(dimensionsValues.length),
        errors: [],
        completed: false,
        progress: 0,
      };

      let completedCount = 0;
      dimensionsValues.forEach((dimensionValues, index) => {
        results.responses[index] = dimensionValues.data;
        if (dimensionValues.error) {
          results.errors.push(
            new Error(
              dimensionValues.error.response?.data?.message ??
                dimensionValues.error.message ??
                "Unknown error",
            ),
          );
        }
        if (!dimensionValues.isFetching) {
          completedCount++;
        }
      });
      results.completed = completedCount === dimensionsValues.length;
      results.progress = Math.round(
        (completedCount * 100) / dimensionsValues.length,
      );

      return results;
    },
  );
}

function getValuesForDimension(
  instanceId: string,
  metricsViewName: string,
  dimension: MetricsViewSpecDimensionV2,
  searchText: string,
) {
  const dimensionName = dimension.name ?? "";
  return createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      dimensions: [{ name: dimensionName }],
      measures: [],
      where: createLikeExpression(dimensionName, `%${searchText}%`),
      limit: "100",
    },
    {
      query: {
        queryFn: ({ signal }) =>
          batch.fetch(
            "metricsViewAggregation",
            {
              instanceId,
              metricsView: metricsViewName,
              dimensions: [{ name: dimensionName }],
              measures: [],
              where: createLikeExpression(dimensionName, `%${searchText}%`),
              limit: "100",
            },
            signal,
          ),
        select: (resp) => ({
          dimension: dimensionName,
          values: resp?.data?.map((d) => d[dimensionName]) ?? [],
        }),
      },
    },
  );
}
