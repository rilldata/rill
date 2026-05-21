import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getRuntimeServiceGetResourceQueryOptions,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQueries } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getCombinedMeasuresAndDimensionsForMetricsViews(
  client: RuntimeClient,
  metricsViewNamesStore: Readable<string[]>,
) {
  const metricsViewQueryOptions = derived(
    metricsViewNamesStore,
    (metricsViewNames) =>
      metricsViewNames.map((metricsViewName) =>
        getRuntimeServiceGetResourceQueryOptions(client, {
          name: { kind: ResourceKind.MetricsView, name: metricsViewName },
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
      const whereSqls: string[] = [];

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

        if (spec.whereSql) {
          whereSqls.push(spec.whereSql);
        }
      });

      // For multi-metricsview views (e.g. embedded), join the locked filters with AND so the chip
      // reflects the conjunction the backend actually applies across all of them.
      const whereSql =
        whereSqls.length === 0
          ? undefined
          : whereSqls.length === 1
            ? whereSqls[0]
            : whereSqls.map((s) => `(${s})`).join(" AND ");

      return {
        measures,
        dimensions,
        whereSql,
      };
    },
  });
}
