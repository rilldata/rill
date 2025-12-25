import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  getRuntimeServiceListResourcesQueryOptions,
  type RpcStatus,
  runtimeServiceGetResource,
  runtimeServiceListResources,
  type V1ExploreSpec,
  type V1GetResourceResponse,
  type V1ListResourcesResponse,
  type V1MetricsViewSpec,
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryOptions, QueryClient } from "@tanstack/svelte-query";
import type { ErrorType } from "../../runtime-client/http-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import httpClient from "@rilldata/web-common/runtime-client/http-client";

export enum ResourceKind {
  ProjectParser = "rill.runtime.v1.ProjectParser",
  Alert = "rill.runtime.v1.Alert",
  Report = "rill.runtime.v1.Report",
  Source = "rill.runtime.v1.Source",
  Connector = "rill.runtime.v1.Connector",
  Model = "rill.runtime.v1.Model",
  MetricsView = "rill.runtime.v1.MetricsView",
  Explore = "rill.runtime.v1.Explore",
  Theme = "rill.runtime.v1.Theme",
  Component = "rill.runtime.v1.Component",
  Canvas = "rill.runtime.v1.Canvas",
  API = "rill.runtime.v1.API",
  RefreshTrigger = "rill.runtime.v1.RefreshTrigger",
}

export function displayResourceKind(kind: ResourceKind | undefined) {
  switch (kind) {
    case ResourceKind.ProjectParser:
      return "project parser";
    case ResourceKind.Alert:
      return "alert";
    case ResourceKind.Report:
      return "report";
    case ResourceKind.Source:
      return "source";
    case ResourceKind.Connector:
      return "connector";
    case ResourceKind.Model:
      return "model";
    case ResourceKind.MetricsView:
      return "metrics view";
    case ResourceKind.Explore:
      return "dashboard";
    case ResourceKind.Theme:
      return "theme";
    case ResourceKind.Component:
      return "component";
    case ResourceKind.Canvas:
      return "dashboard";
    case ResourceKind.API:
      return "API";
    case ResourceKind.RefreshTrigger:
      return "refresh trigger";
    default:
      return undefined;
  }
}

export type UserFacingResourceKinds = Exclude<
  ResourceKind,
  ResourceKind.ProjectParser | ResourceKind.RefreshTrigger
>;

export const SingletonProjectParserName = "parser";

// In the UI, we shouldn't show the `rill.runtime.v1` prefix
export function prettyResourceKind(kind: string) {
  return kind.replace(/^rill\.runtime\.v1\./, "");
}

/**
 * Coerce resource kind to match UI representation.
 * Models that are defined-as-source are displayed as Sources in the sidebar and graph.
 * This ensures consistent representation across the application.
 *
 * @param res - The resource to check
 * @returns The coerced ResourceKind, or undefined if the resource has no kind
 *
 * @example
 * // A model that is defined-as-source
 * coerceResourceKind(modelResource) // Returns ResourceKind.Source
 *
 * // A normal model
 * coerceResourceKind(normalModel) // Returns ResourceKind.Model
 */
export function coerceResourceKind(res: V1Resource): ResourceKind | undefined {
  const raw = res.meta?.name?.kind as ResourceKind | undefined;
  if (raw === ResourceKind.Model) {
    // A resource is a Source if it's a model defined-as-source and its result table matches the resource name
    const name = res.meta?.name?.name;
    const resultTable = res.model?.state?.resultTable;
    const definedAsSource = res.model?.spec?.definedAsSource;
    if (name && resultTable === name && definedAsSource === true) {
      return ResourceKind.Source;
    }
  }
  return raw;
}

export function useResource<T = V1Resource>(
  instanceId: string,
  name: string,
  kind: ResourceKind,
  queryOptions?: Partial<
    CreateQueryOptions<
      V1GetResourceResponse,
      ErrorType<RpcStatus>,
      T // T is the return type of the `select` function
    >
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": kind,
      "name.name": name,
    },
    {
      query: {
        select: (data) => data?.resource as T,
        enabled: !!instanceId && !!name && !!kind,
        ...queryOptions,
      },
    },
    queryClient,
  );
}

// FIXME: To remove this duplicate of `useResource` https://github.com/rilldata/rill/pull/5531#discussion_r1733027626
/**
 * `useResourceV2` is a more flexible version of `useResource` that accepts
 *  any `queryOptions`, not just `select` and `queryClient`.
 */
export function useResourceV2<T = V1Resource>(
  instanceId: string,
  name: string,
  kind: ResourceKind,
  queryOptions?: Partial<
    CreateQueryOptions<
      V1GetResourceResponse,
      ErrorType<RpcStatus>,
      T // T is the return type of the `select` function
    >
  >,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": kind,
      "name.name": name,
    },
    {
      query: {
        select: (data) => data?.resource as T,
        enabled: !!instanceId && !!name && !!kind,
        ...queryOptions,
      },
    },
    queryClient,
  );
}

export function useProjectParser(
  queryClient: QueryClient,
  instanceId: string,
  queryOptions?: Partial<
    CreateQueryOptions<V1GetResourceResponse, ErrorType<RpcStatus>, V1Resource>
  >,
) {
  return useResource(
    instanceId,
    SingletonProjectParserName,
    ResourceKind.ProjectParser,
    queryOptions,
    queryClient,
  );
}

export function useFilteredResources<T = Array<V1Resource>>(
  instanceId: string,
  kind: ResourceKind,
  selector: (data: V1ListResourcesResponse) => T = (data) =>
    data.resources as T,
) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: kind,
    },
    {
      query: {
        select: selector,
      },
    },
    queryClient,
  );
}

/**
 * Fetches all resources and filters them client side.
 * This is to improve network requests since we need the full list all the time as well.
 */
export function useClientFilteredResources(
  instanceId: string,
  kind: ResourceKind,
  filter: (res: V1Resource) => boolean = () => true,
) {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) =>
        data.resources?.filter(
          (res) => res.meta?.name?.kind === kind && filter(res),
        ) ?? [],
    },
  });
}

export function resourceIsLoading(resource?: V1Resource) {
  return (
    !!resource &&
    resource.meta?.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE
  );
}

export async function fetchResource(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  kind: ResourceKind,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": name,
      "name.kind": kind,
    }),
    queryFn: () =>
      runtimeServiceGetResource(instanceId, {
        "name.name": name,
        "name.kind": kind,
      }),
  });
  return resp.resource;
}

export async function fetchResources(
  queryClient: QueryClient,
  instanceId: string,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
    queryFn: () => runtimeServiceListResources(instanceId, {}),
  });
  return resp.resources ?? [];
}

export function getMetricsViewAndExploreSpecsQueryOptions() {
  const instanceId = httpClient.getInstanceId();
  return getRuntimeServiceListResourcesQueryOptions(instanceId, undefined, {
    query: {
      select: (data) => {
        const metricsViewSpecsMap = new Map<string, V1MetricsViewSpec>();
        const exploreSpecsMap = new Map<string, V1ExploreSpec>();
        const exploreForMetricViewsMap = new Map<string, string>();

        data.resources?.forEach((res) => {
          if (res.metricsView?.state?.validSpec) {
            metricsViewSpecsMap.set(
              res.meta?.name?.name ?? "",
              res.metricsView.state.validSpec,
            );
          } else if (res.explore?.state?.validSpec) {
            const metricsViewName =
              res.explore.state.validSpec.metricsView ?? "";
            const exploreName = res.meta?.name?.name ?? "";
            exploreForMetricViewsMap.set(metricsViewName, exploreName);
            exploreSpecsMap.set(exploreName, res.explore.state.validSpec);
          }
        });

        return {
          metricsViewSpecsMap,
          exploreSpecsMap,
          exploreForMetricViewsMap,
        };
      },
    },
  });
}
