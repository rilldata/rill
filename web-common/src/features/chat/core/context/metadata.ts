import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, get, type Readable } from "svelte/store";
import {
  createQueryServiceResolveCanvas,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1CanvasSpec,
  type V1ComponentSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

/**
 * Metadata used to map a value to a label.
 * Currently only metrics view metadata exists but it is meant for any type.
 */
export type InlineContextMetadata = {
  metricsViewSpecs: Record<string, MetricsViewMetadata>;
  canvasSpecs: Record<string, V1CanvasSpec>;
  componentSpecs: Record<string, V1ComponentSpec>;
};
export type MetricsViewMetadata = {
  metricsViewSpec: V1MetricsViewSpec;
  measures: Record<string, MetricsViewSpecMeasure>;
  dimensions: Record<string, MetricsViewSpecDimension>;
};

/**
 * Creates a store that contains a map of metrics view names to their metadata.
 * Each metrics view metadata has a reference to its spec, and a map of for measure and dimension spec by their names.
 */
export function getInlineChatContextMetadata(): Readable<InlineContextMetadata> {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  const canvasResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(ResourceKind.Canvas, (res) =>
      Boolean(res.canvas?.state?.validSpec),
    ),
    queryClient,
  );

  const instanceId = get(runtime).instanceId;

  return derived(
    [metricsViewsQuery, canvasResourcesQuery],
    ([metricsViewsResp, canvasResourcesResp], set) => {
      const metricsViews = metricsViewsResp.data ?? [];
      const metricsViewSpecs = Object.fromEntries(
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
      );

      const canvasResources = canvasResourcesResp.data ?? [];
      const canvasSpecs = Object.fromEntries(
        canvasResources.map((r) => [
          r.meta?.name?.name ?? "",
          r.canvas?.state?.validSpec ?? {},
        ]),
      );

      const canvasQueries = canvasResources.map((r) =>
        createQueryServiceResolveCanvas(
          instanceId,
          r.meta?.name?.name ?? "",
          {},
          undefined,
          queryClient,
        ),
      );
      const canvasQueriesStore = derived(canvasQueries, (canvasResponses) => {
        const componentSpecs: Record<string, V1ComponentSpec> = {};

        canvasResponses.forEach((canvasResponse) => {
          const resolvedComponents = canvasResponse.data?.resolvedComponents;
          if (!resolvedComponents) return;
          for (const name in resolvedComponents) {
            componentSpecs[name] =
              resolvedComponents[name].component?.state?.validSpec ?? {};
          }
        });

        return componentSpecs;
      });

      return canvasQueriesStore.subscribe((componentSpecs) =>
        set({
          metricsViewSpecs,
          canvasSpecs,
          componentSpecs,
        }),
      );
    },
  );
}
