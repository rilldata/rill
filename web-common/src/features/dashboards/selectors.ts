import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  createAndExpression,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import {
  ResourceKind,
  useClientFilteredResources,
  useFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type RpcStatus,
  type V1Expression,
  type V1GetResourceResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  createQueryServiceMetricsViewTimeRange,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryOptions,
  CreateQueryResult,
  QueryClient,
} from "@tanstack/svelte-query";
import { Settings } from "luxon";
import { derived } from "svelte/store";
import type { ErrorType } from "../../runtime-client/http-client";
import type { DimensionThresholdFilter } from "./stores/metrics-explorer-entity";

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

  return derived(
    [useMetricsViewValidSpec(instanceId, metricsViewName)],
    ([metricsView], set) =>
      createQueryServiceMetricsViewTimeRange(
        instanceId,
        metricsViewName,
        {},
        {
          query: {
            ...queryOptions,
            enabled: !!metricsView.data?.timeDimension && queryOptions?.enabled,
          },
        },
        queryClient,
      ).subscribe(set),
  );
}

/**
 * Wrapper function that fetches full time range.
 * Uses useExploreValidSpec unlike the above useMetricsViewTimeRange since it is more widely used.
 *
 * Does an additional validation where null min and max returned throws an error instead.
 *
 * TODO: replace usages of useMetricsViewTimeRange with this eventually
 */
export function useFullTimeRangeQuery(
  instanceId: string,
  exploreName: string,
  options?: {
    query?: CreateQueryOptions<V1MetricsViewTimeRangeResponse>;
  },
  queryClient?: QueryClient,
): CompoundQueryResult<V1MetricsViewTimeRangeResponse> {
  const { query: queryOptions } = options ?? {};

  return derived(
    useExploreValidSpec(instanceId, exploreName, undefined, queryClient),
    (validSpecResp, set) => {
      const firstDayOfWeek = validSpecResp.data?.metricsView?.firstDayOfWeek;
      const metricsViewName = validSpecResp.data?.explore?.metricsView ?? "";

      Settings.defaultWeekSettings = {
        firstDay: normalizeWeekday(firstDayOfWeek),
        weekend: [6, 7],
        minimalDays: 4,
      };

      const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
      if (!metricsViewSpec.timeDimension) {
        // We return early to avoid having isLoading=true when time dimension is not present.
        // This allows us to check isLoading further down without any issues of it getting stuck.
        set({
          data: undefined,
          error: null,
          isFetching: false,
          isLoading: false,
        });
        return;
      }

      createQueryServiceMetricsViewTimeRange(
        instanceId,
        metricsViewName,
        {},
        { query: queryOptions },
        queryClient,
      ).subscribe((timeRangeResp) => {
        // The timeRangeSummary is null when there are 0 rows of data
        // Notably, this happens when a security policy fully restricts a user from reading any data
        if (
          timeRangeResp.data?.timeRangeSummary?.min === null &&
          timeRangeResp.data?.timeRangeSummary?.max === null
        ) {
          set({
            data: undefined,
            error: new Error(
              "This dashboard currently has no data to display. This may be due to access permissions.",
            ),
            isFetching: false,
            isLoading: false,
          });
          return;
        }

        set(timeRangeResp);
      });
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
