import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createQueryServiceResolveCanvas,
  getQueryServiceResolveCanvasQueryOptions,
  type RpcStatus,
  type V1CanvasSpec,
  type V1MetricsView,
  type V1ResolveCanvasResponse,
  type V1ResolveCanvasResponseResolvedComponents,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import type {
  CreateQueryOptions,
  CreateQueryResult,
  QueryClient,
} from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

/**
 * Returns the default metrics view for a given instance, prioritizing in order:
 * 1. A specific metrics view by name if provided
 * 2. The first metrics view with a time dimension
 * 3. The first available metrics view
 */
export function useDefaultMetrics(
  instanceId: string,
  metricsViewName?: string,
) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView, (data) => {
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
}

export function useCanvas(
  instanceId: string,
  canvasName: string,
  queryOptions?: Partial<
    CreateQueryOptions<
      V1ResolveCanvasResponse,
      ErrorType<RpcStatus>,
      CanvasResponse
    >
  >,
  queryClient?: QueryClient,
): CreateQueryResult<CanvasResponse, ErrorType<RpcStatus>> {
  return createQueryServiceResolveCanvas(
    instanceId,
    canvasName,
    {},
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
            canvas: data.canvas?.canvas?.state?.validSpec,
            components: data.resolvedComponents,
            metricsViews,
            filePath: data.canvas?.meta?.filePaths?.[0],
          };
        },

        enabled: !!canvasName,
        ...queryOptions,
      },
    },
    queryClient,
  );
}

export function getCanvasQueryOptions(canvasNameStore: Readable<string>) {
  return derived([runtime, canvasNameStore], ([{ instanceId }, canvasName]) =>
    getQueryServiceResolveCanvasQueryOptions(
      instanceId,
      canvasName,
      {},
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
              canvas: data.canvas?.canvas?.state?.validSpec,
              components: data.resolvedComponents,
              metricsViews,
              filePath: data.canvas?.meta?.filePaths?.[0],
            };
          },

          enabled: !!canvasName,
        },
      },
    ),
  );
}
