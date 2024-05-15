import { createLikeExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { createQueryServiceStreamingQueryBatch } from "@rilldata/web-common/runtime-client/createQueryServiceStreamingQueryBatch";

export function useDimensionSearchResults(
  instanceId: string,
  metricsViewName: string,
  metricsView: V1MetricsViewSpec,
  searchText: string,
) {
  const dimensions = metricsView.dimensions ?? [];
  return createQueryServiceStreamingQueryBatch(
    instanceId,
    dimensions?.map((d) => ({
      metricsViewAggregationRequest: {
        instanceId,
        metricsView: metricsViewName,
        dimensions: [{ name: d.name }],
        measures: [],
        where: createLikeExpression(d.name ?? "", `%${searchText}%`),
        limit: "100",
      },
    })) ?? [],
    (resp, index) => ({
      dimension: dimensions[index]?.name ?? "",
      values:
        resp.metricsViewAggregationResponse?.data?.map(
          (d) => d[dimensions[index]?.name ?? ""],
        ) ?? [],
    }),
  );
}
