import {
  createConnectorServiceOLAPGetTable,
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
  V1ListResourcesResponse,
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export enum ResourceKind {
  ProjectParser = "rill.runtime.v1.ProjectParser",
  Source = "rill.runtime.v1.Source",
  Model = "rill.runtime.v1.Model",
  MetricsView = "rill.runtime.v1.MetricsView",
  Report = "rill.runtime.v1.Report",
  Alert = "rill.runtime.v1.Alert",
  Theme = "rill.runtime.v1.Theme",
  Component = "rill.runtime.v1.Component",
  Dashboard = "rill.runtime.v1.Dashboard",
  API = "rill.runtime.v1.API",
}
export type UserFacingResourceKinds = Exclude<
  ResourceKind,
  ResourceKind.ProjectParser
>;
export const SingletonProjectParserName = "parser";
export const ResourceShortNameToKind: Record<string, ResourceKind> = {
  source: ResourceKind.Source,
  model: ResourceKind.Model,
  metricsview: ResourceKind.MetricsView,
  metrics_view: ResourceKind.MetricsView,
  component: ResourceKind.Component,
  dashboard: ResourceKind.Dashboard,
  report: ResourceKind.Report,
  alert: ResourceKind.Alert,
  theme: ResourceKind.Theme,
};

// In the UI, we shouldn't show the `rill.runtime.v1` prefix
export function prettyResourceKind(kind: string) {
  return kind.replace(/^rill\.runtime\.v1\./, "");
}

export function useResource<T = V1Resource>(
  instanceId: string,
  name: string,
  kind: ResourceKind,
  selector?: (data: V1Resource) => T,
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
        select: (data) =>
          (selector ? selector(data?.resource) : data?.resource) as T,
        enabled: !!instanceId && !!name && !!kind,
        queryClient,
      },
    },
  );
}

export function useProjectParser(queryClient: QueryClient, instanceId: string) {
  return useResource(
    instanceId,
    SingletonProjectParserName,
    ResourceKind.ProjectParser,
    undefined,
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
      kind,
    },
    {
      query: {
        select: selector,
      },
    },
  );
}

export function useFilteredResourceNames(
  instanceId: string,
  kind: ResourceKind,
) {
  return useFilteredResources<Array<string>>(instanceId, kind, (data) =>
    data.resources.map((res) => res.meta.name.name),
  );
}

export function createSchemaForTable(
  instanceId: string,
  resourceName: string,
  resourceKind: ResourceKind,
  queryClient?: QueryClient,
) {
  return derived(
    useResource(instanceId, resourceName, resourceKind, undefined, queryClient),
    (res, set) => {
      const tableSpec = res.data?.source ?? res.data?.model;
      return createConnectorServiceOLAPGetTable(
        {
          instanceId,
          table: tableSpec?.state?.table,
          connector: tableSpec?.state?.connector,
        },
        {
          query: {
            enabled: !!tableSpec?.state?.table && !!tableSpec?.state?.connector,
            queryClient,
          },
        },
      ).subscribe(set);
    },
  ) as ReturnType<typeof createConnectorServiceOLAPGetTable>;
}

export function resourceIsLoading(resource?: V1Resource) {
  return (
    !!resource &&
    resource.meta?.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE
  );
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
