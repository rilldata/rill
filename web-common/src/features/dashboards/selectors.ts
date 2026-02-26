import {
  createAndExpression,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  ResourceKind,
  useClientFilteredResources,
  useFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getExploreValidSpecQueryOptions,
  useExploreValidSpec,
} from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceListResources,
  getQueryServiceMetricsViewTimeRangeQueryOptions,
  getRuntimeServiceListResourcesQueryOptions,
  type RpcStatus,
  type V1Expression,
  type V1GetResourceResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createQuery,
  type CreateQueryOptions,
  type CreateQueryResult,
  type QueryClient,
} from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";

export function useMetricsView(
  client: RuntimeClient,
  metricsViewName: string,
  queryOptions?: CreateQueryOptions<
    V1GetResourceResponse,
    RpcStatus,
    V1Resource
  >,
) {
  return useResource(
    client,
    metricsViewName,
    ResourceKind.MetricsView,
    queryOptions,
  );
}

export function getValidMetricsViewsQueryOptions(client: RuntimeClient) {
  return getRuntimeServiceListResourcesQueryOptions(
    client,
    {
      kind: ResourceKind.MetricsView,
    },
    {
      query: {
        select: (data) =>
          data?.resources?.filter((res) => !!res.metricsView?.state?.validSpec),
      },
    },
  );
}

export function useValidExplores(client: RuntimeClient) {
  // This is used in cloud as well so do not use "useClientFilteredResources"
  return useFilteredResources(client, ResourceKind.Explore, (data) =>
    data?.resources?.filter((res) => !!res.explore?.state?.validSpec),
  );
}

export function useValidCanvases(client: RuntimeClient) {
  return useFilteredResources(client, ResourceKind.Canvas, (data) =>
    data?.resources?.filter((res) => !!res.canvas?.state?.validSpec),
  );
}

export function useValidDashboards(client: RuntimeClient) {
  return createRuntimeServiceListResources(
    client,
    {}, // TODO: it'd be nice if we could provide multiple kinds here
    {
      query: {
        select: (data) => {
          // Filter for valid Explores and Canvases
          return data?.resources?.filter(
            (res) =>
              !!res.explore?.state?.validSpec || !!res.canvas?.state?.validSpec,
          );
        },
      },
    },
  );
}

/**
 * Gets the valid metrics view spec. Only to be used in displaying a dashboard.
 * Use {@link useDashboard} in the metrics view editor and other use cases.
 */
export const useMetricsViewValidSpec = <T = V1MetricsViewSpec>(
  client: RuntimeClient,
  metricsViewName: string,
  selector?: (meta: V1MetricsViewSpec) => T,
) => {
  return useResource<T>(client, metricsViewName, ResourceKind.MetricsView, {
    select: (data) =>
      selector
        ? selector(data.resource?.metricsView?.state?.validSpec ?? {})
        : (data.resource?.metricsView?.state?.validSpec as T),
  });
};

export function useMetricsViewTimeRange(
  client: RuntimeClient,
  metricsViewName: string,
  options?: {
    query?: CreateQueryOptions<V1MetricsViewTimeRangeResponse>;
  },
  queryClient?: QueryClient,
): CreateQueryResult<V1MetricsViewTimeRangeResponse> {
  const { query: queryOptions } = options ?? {};

  const fullTimeRangeQueryOptionsStore = derived(
    useMetricsViewValidSpec(client, metricsViewName),
    (validSpecResp) => {
      const metricsViewSpec = validSpecResp.data ?? {};

      return getQueryServiceMetricsViewTimeRangeQueryOptions(
        client,
        { metricsViewName },
        {
          query: {
            ...queryOptions,
            enabled: Boolean(metricsViewSpec.timeDimension),
          },
        },
      );
    },
  );

  return createQuery(fullTimeRangeQueryOptionsStore, queryClient);
}

export function getMetricsViewTimeRangeFromExploreQueryOptions(
  client: RuntimeClient,
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(client, exploreNameStore),
    queryClient,
  );

  return derived([validSpecQuery], ([validSpecResp]) => {
    const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
    const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
    const metricsViewName = exploreSpec.metricsView ?? "";

    return getQueryServiceMetricsViewTimeRangeQueryOptions(
      client,
      { metricsViewName },
      {
        query: {
          enabled: !!metricsViewSpec.timeDimension,
        },
      },
    );
  });
}

export function hasValidMetricsViewTimeRange(
  client: RuntimeClient,
  exploreName: string,
) {
  const fullTimeRangeQueryOptionsStore = derived(
    useExploreValidSpec(client, exploreName),
    (validSpecResp) => {
      const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
      const exploreSpec = validSpecResp.data?.explore ?? {};
      const metricsViewName = exploreSpec.metricsView ?? "";

      return getQueryServiceMetricsViewTimeRangeQueryOptions(
        client,
        { metricsViewName },
        {
          query: {
            enabled: Boolean(metricsViewName && metricsViewSpec.timeDimension),
          },
        },
      );
    },
  );
  const fullTimeRangeQuery = createQuery(
    fullTimeRangeQueryOptionsStore,
    queryClient,
  );

  return derived(
    fullTimeRangeQuery,
    (fullTimeRange) => !fullTimeRange.isPending && !fullTimeRange.isError,
  );
}

export function getFiltersForOtherDimensions(
  whereFilter: V1Expression,
  dimName: string,
) {
  const exprIdx = whereFilter?.cond?.exprs?.findIndex((e) =>
    matchExpressionByName(e, dimName),
  );
  if (exprIdx === undefined || exprIdx === -1) return whereFilter;

  return createAndExpression(
    whereFilter.cond?.exprs?.filter(
      (e) => !matchExpressionByName(e, dimName),
    ) ?? [],
  );
}

export function additionalMeasures(
  activeMeasureName: string | null,
  dimensionThresholdFilters: DimensionThresholdFilter[],
) {
  const measures = new Set<string>(
    activeMeasureName ? [activeMeasureName] : [],
  );
  dimensionThresholdFilters.forEach(({ filters }) => {
    filters.forEach((filter) => {
      measures.add(filter.measure);
    });
  });
  return [...measures];
}

export const useGetMetricsViewsForModel = (
  client: RuntimeClient,
  modelName: string,
) => {
  return useClientFilteredResources(
    client,
    ResourceKind.MetricsView,
    (res) =>
      res.metricsView?.spec?.model === modelName ||
      res.metricsView?.spec?.table === modelName,
  );
};

export const useGetExploresForMetricsView = (
  client: RuntimeClient,
  metricsViewName: string,
) => {
  return useClientFilteredResources(
    client,
    ResourceKind.Explore,
    (res) => res.explore?.spec?.metricsView === metricsViewName,
  );
};
