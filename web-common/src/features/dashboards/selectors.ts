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
import {
  getQueryServiceMetricsViewSchemaQueryOptions,
  getQueryServiceMetricsViewTimeRangeQueryOptions,
  type RpcStatus,
  type V1Expression,
  type V1GetResourceResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import {
  createQuery,
  type CreateQueryOptions,
  type CreateQueryResult,
  type QueryClient,
} from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import type { ErrorType } from "../../runtime-client/http-client";
import type { DimensionThresholdFilter } from "web-common/src/features/dashboards/stores/explore-state";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

export function useMetricsView(
  instanceId: string,
  metricsViewName: string,
  queryOptions?: CreateQueryOptions<
    V1GetResourceResponse,
    ErrorType<RpcStatus>,
    V1Resource
  >,
) {
  return useResource(
    instanceId,
    metricsViewName,
    ResourceKind.MetricsView,
    queryOptions,
  );
}

export function useValidExplores(instanceId: string) {
  // This is used in cloud as well so do not use "useClientFilteredResources"
  return useFilteredResources(instanceId, ResourceKind.Explore, (data) =>
    data?.resources?.filter((res) => !!res.explore?.state?.validSpec),
  );
}

export function useValidCanvases(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.Canvas, (data) =>
    data?.resources?.filter((res) => !!res.canvas?.state?.validSpec),
  );
}

export function useValidDashboards(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    undefined, // TODO: it'd be nice if we could provide multiple kinds here
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
  instanceId: string,
  metricsViewName: string,
  selector?: (meta: V1MetricsViewSpec) => T,
) => {
  return useResource<T>(instanceId, metricsViewName, ResourceKind.MetricsView, {
    select: (data) =>
      selector
        ? selector(data.resource?.metricsView?.state?.validSpec ?? {})
        : (data.resource?.metricsView?.state?.validSpec as T),
  });
};

export function useMetricsViewTimeRange(
  instanceId: string,
  metricsViewName: string,
  options?: {
    query?: CreateQueryOptions<V1MetricsViewTimeRangeResponse>;
  },
  queryClient?: QueryClient,
): CreateQueryResult<V1MetricsViewTimeRangeResponse> {
  const { query: queryOptions } = options ?? {};

  const fullTimeRangeQueryOptionsStore = derived(
    useMetricsViewValidSpec(instanceId, metricsViewName),
    (validSpecResp) => {
      const metricsViewSpec = validSpecResp.data ?? {};

      return getQueryServiceMetricsViewTimeRangeQueryOptions(
        instanceId,
        metricsViewName,
        {},
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
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );

  return derived(
    [runtime, validSpecQuery],
    ([{ instanceId }, validSpecResp]) => {
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const metricsViewName = exploreSpec.metricsView ?? "";

      return getQueryServiceMetricsViewTimeRangeQueryOptions(
        instanceId,
        metricsViewName,
        {},
        {
          query: {
            enabled: !!metricsViewSpec.timeDimension,
          },
        },
      );
    },
  );
}

export function hasValidMetricsViewTimeRange(
  instanceId: string,
  exploreName: string,
) {
  const fullTimeRangeQueryOptionsStore = derived(
    useExploreValidSpec(instanceId, exploreName),
    (validSpecResp) => {
      const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
      const exploreSpec = validSpecResp.data?.explore ?? {};
      const metricsViewName = exploreSpec.metricsView ?? "";

      return getQueryServiceMetricsViewTimeRangeQueryOptions(
        instanceId,
        metricsViewName,
        {},
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

export function getMetricsViewSchemaOptions(
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );

  return derived(
    [runtime, validSpecQuery],
    ([{ instanceId }, validSpecResp]) => {
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const metricsViewName = exploreSpec.metricsView ?? "";

      return getQueryServiceMetricsViewSchemaQueryOptions(
        instanceId,
        metricsViewName,
        {},
        {
          query: {
            enabled: Boolean(metricsViewName),
          },
        },
      );
    },
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
  instanceId: string,
  modelName: string,
) => {
  return useClientFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
    (res) =>
      res.metricsView?.spec?.model === modelName ||
      res.metricsView?.spec?.table === modelName,
  );
};

export const useGetExploresForMetricsView = (
  instanceId: string,
  metricsViewName: string,
) => {
  return useClientFilteredResources(
    instanceId,
    ResourceKind.Explore,
    (res) => res.explore?.spec?.metricsView === metricsViewName,
  );
};
