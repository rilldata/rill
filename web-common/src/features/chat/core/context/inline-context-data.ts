import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived } from "svelte/store";
import {
  type InlineContextMetadata,
  type MetricsViewMetadata,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

/**
 * Creates a store that contains a map of metrics view names to their metadata.
 * Each metrics view metadata has a reference to its spec, and a map of for measure and dimension spec by their names.
 */
export function getInlineChatContextMetadata() {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  return derived(metricsViewsQuery, (metricsViewsResp) => {
    const metricsViews = metricsViewsResp.data ?? [];
    return Object.fromEntries(
      metricsViews.map((mv) => {
        const mvName = mv.meta?.name?.name ?? "";
        const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};

        const measures = Object.fromEntries(
          metricsViewSpec?.measures?.map((m) => [m.name!, m]) ?? [],
        );

        const dimensions = Object.fromEntries(
          metricsViewSpec?.dimensions?.map((d) => [d.name!, d]) ?? [],
        );

        return [
          mvName,
          <MetricsViewMetadata>{
            metricsViewSpec,
            measures,
            dimensions,
          },
        ];
      }),
    ) as InlineContextMetadata;
  });
}
