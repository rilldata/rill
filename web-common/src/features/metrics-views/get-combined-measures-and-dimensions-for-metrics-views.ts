import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getRuntimeServiceGetResourceQueryOptions,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { createQueries } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getCombinedMeasuresAndDimensionsForMetricsViews(
  metricsViewNamesStore: Readable<string[]>,
) {
  const metricsViewQueryOptions = derived(
    [runtime, metricsViewNamesStore],
    ([{ instanceId }, metricsViewNames]) =>
      metricsViewNames.map((metricsViewName) =>
        getRuntimeServiceGetResourceQueryOptions(instanceId, {
          "name.kind": ResourceKind.MetricsView,
          "name.name": metricsViewName,
        }),
      ),
  );

  return createQueries({
    queries: metricsViewQueryOptions,
    combine: (metricsViewQueryResponses) => {
      const seenMeasureNames = new Set<string>();
      const measures: MetricsViewSpecMeasure[] = [];
      const seenDimensionNames = new Set<string>();
      const dimensions: MetricsViewSpecDimension[] = [];

      metricsViewQueryResponses.forEach((metricsViewQueryResponse) => {
        const spec =
          metricsViewQueryResponse.data?.resource?.metricsView?.state
            ?.validSpec;
        if (!spec?.measures || !spec?.dimensions) return;

        measures.push(
          ...spec.measures.filter((measure) => {
            const measureName = measure.name!;
            if (seenMeasureNames.has(measureName)) return false;
            seenMeasureNames.add(measureName);
            return true;
          }),
        );

        dimensions.push(
          ...spec.dimensions.filter((dimension) => {
            const dimensionName = dimension.name!;
            if (seenDimensionNames.has(dimensionName)) return false;
            seenDimensionNames.add(dimensionName);
            return true;
          }),
        );
      });

      return {
        measures,
        dimensions,
      };
    },
  });
}
