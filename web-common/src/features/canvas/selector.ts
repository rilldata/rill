import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@rilldata/svelte-query";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceResolveCanvas,
  type RpcStatus,
  type V1CanvasSpec,
  type V1MetricsViewV2,
  type V1ResolveCanvasResponse,
  type V1ResolveCanvasResponseResolvedComponents,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

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
  metricsViews: Record<string, V1MetricsViewV2 | undefined>;
}

export function useCanvas(
  instanceId: string,
  canvasName: string,
  queryOptions?: CreateQueryOptions<
    V1ResolveCanvasResponse,
    ErrorType<RpcStatus>,
    CanvasResponse
  >,
): CreateQueryResult<CanvasResponse, ErrorType<RpcStatus>> {
  const defaultQueryOptions: CreateQueryOptions<
    V1ResolveCanvasResponse,
    ErrorType<RpcStatus>,
    CanvasResponse
  > = {
    select: (data) => {
      const metricsViews: Record<string, V1MetricsViewV2 | undefined> = {};
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
      };
    },
    queryClient,
    enabled: !!canvasName,
  };
  return createQueryServiceResolveCanvas(
    instanceId,
    canvasName,
    {},
    {
      query: {
        ...defaultQueryOptions,
        ...queryOptions,
      },
    },
  );
}
