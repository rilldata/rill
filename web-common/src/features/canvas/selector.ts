import type { ConnectError } from "@connectrpc/connect";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createQueryServiceResolveCanvas,
  type V1CanvasSpec,
  type V1MetricsView,
  type V1ResolveCanvasResponse,
  type V1ResolveCanvasResponseResolvedComponents,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type {
  CreateQueryOptions,
  CreateQueryResult,
  QueryClient,
} from "@tanstack/svelte-query";
/**
 * Returns the default metrics view for a given instance, prioritizing in order:
 * 1. A specific metrics view by name if provided
 * 2. The first metrics view with a time dimension
 * 3. The first available metrics view
 */
export function useDefaultMetrics(
  client: RuntimeClient,
  metricsViewName?: string,
) {
  return useFilteredResources(client, ResourceKind.MetricsView, (data) => {
    const validMetricsViews = data?.resources?.filter(
      (res) => !!res.metricsView?.state?.validSpec,
    );

    if (validMetricsViews && validMetricsViews?.length > 0) {
      if (metricsViewName) {
        const matchingMetricsView = validMetricsViews.find(
          (res) => res.meta?.name?.name === metricsViewName,
        );

        if (matchingMetricsView) {
          const metricsViewSpec =
            matchingMetricsView.metricsView?.state?.validSpec;

          return {
            metricsViewName,
            metricsViewSpec,
          };
        }
      }

      const metricsViewWithTimeDimension = validMetricsViews.find((res) => {
        const spec = res.metricsView?.state?.validSpec;
        return !!spec?.timeDimension;
      });

      if (metricsViewWithTimeDimension) {
        const metricsViewSpec =
          metricsViewWithTimeDimension.metricsView?.state?.validSpec;
        const metricsViewName = metricsViewWithTimeDimension.meta?.name
          ?.name as string;

        return {
          metricsViewName,
          metricsViewSpec,
        };
      }

      const firstMetricsView =
        validMetricsViews[0].metricsView?.state?.validSpec;
      const firstMetricName = validMetricsViews[0].meta?.name?.name as string;

      return {
        metricsViewName: firstMetricName,
        metricsViewSpec: firstMetricsView,
      };
    }

    return null;
  });
}

export interface CanvasResponse {
  canvas: V1CanvasSpec | undefined;
  components: V1ResolveCanvasResponseResolvedComponents | undefined;
  metricsViews: Record<string, V1MetricsView | undefined>;
  filePath: string | undefined;
  /** YAML paths of the canvas elements hidden from the current user, such as "rows.0.items.1". Only set for the editor. */
  hiddenPaths?: string[];
}

/**
 * The ResolveCanvas request for a canvas.
 * The editor (allowUnvalidatedSpec) gets the full canvas, including the content hidden from the current user,
 * because it maps canvas elements to their position in the YAML. Viewers get the canvas without that content.
 * Use the same request everywhere a canvas is fetched so the queries share a cache entry.
 */
export function getResolveCanvasRequest(
  canvasName: string,
  allowUnvalidatedSpec: boolean,
) {
  return {
    canvas: canvasName,
    unsafe: allowUnvalidatedSpec,
    includeHidden: allowUnvalidatedSpec,
  };
}

export function useCanvas(
  client: RuntimeClient,
  canvasName: string,
  queryOptions?: Partial<
    CreateQueryOptions<V1ResolveCanvasResponse, ConnectError, CanvasResponse>
  >,
  queryClient?: QueryClient,
  allowUnvalidatedSpec = false,
): CreateQueryResult<CanvasResponse, ConnectError> {
  return createQueryServiceResolveCanvas(
    client,
    getResolveCanvasRequest(canvasName, allowUnvalidatedSpec),
    {
      query: {
        select: (data) => {
          const metricsViews: Record<string, V1MetricsView | undefined> = {};
          const refMetricsViews = data?.referencedMetricsViews;
          if (refMetricsViews) {
            Object.keys(refMetricsViews).forEach((key) => {
              metricsViews[key] = refMetricsViews?.[key]?.metricsView;
            });
          }

          return {
            canvas:
              data.canvas?.canvas?.state?.validSpec ??
              (allowUnvalidatedSpec ? data.canvas?.canvas?.spec : undefined),
            components: data.resolvedComponents,
            metricsViews,
            filePath: data.canvas?.meta?.filePaths?.[0],
            hiddenPaths: data.hiddenPaths,
          };
        },

        enabled: !!canvasName && !!client?.instanceId,
        ...queryOptions,
      },
    },
    queryClient,
  );
}
