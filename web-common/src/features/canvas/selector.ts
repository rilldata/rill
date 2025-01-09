import type { CreateQueryOptions } from "@rilldata/svelte-query";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetResource,
  type RpcStatus,
  type V1CanvasSpec,
  type V1GetResourceResponse,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

export function useDefaultMetrics(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView, (data) => {
    const validMetricsViews = data?.resources?.filter(
      (res) => !!res.metricsView?.state?.validSpec,
    );

    if (validMetricsViews && validMetricsViews?.length > 0) {
      const firstMetricsView =
        validMetricsViews[0].metricsView?.state?.validSpec;
      const firstMetricName = validMetricsViews[0].meta?.name?.name as string;
      const firstMeasure = firstMetricsView?.measures?.[0]?.name as string;
      const firstDimension =
        firstMetricsView?.dimensions?.[0]?.name ||
        (firstMetricsView?.dimensions?.[0]?.column as string);

      return {
        metricsView: firstMetricName,
        measure: firstMeasure,
        dimension: firstDimension,
      };
    }

    return null;
  });
}

export function useCanvasValidSpec(
  instanceId: string,
  canvasName: string,
  queryOptions?: CreateQueryOptions<
    V1GetResourceResponse,
    ErrorType<RpcStatus>,
    V1CanvasSpec | undefined
  >,
) {
  const defaultQueryOptions: CreateQueryOptions<
    V1GetResourceResponse,
    ErrorType<RpcStatus>,
    V1CanvasSpec | undefined
  > = {
    select: (data) => data?.resource?.canvas?.state?.validSpec,
    queryClient,
    enabled: !!canvasName,
  };
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.Canvas,
      "name.name": canvasName,
    },
    {
      query: {
        ...defaultQueryOptions,
        ...queryOptions,
      },
    },
  );
}
