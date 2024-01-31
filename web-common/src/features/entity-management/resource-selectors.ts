import {
  createConnectorServiceOLAPGetTable,
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
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
  Theme = "rill.runtime.v1.Theme",
}
export const SingletonProjectParserName = "parser";

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

export function useAllNames(instanceId: string) {
  const select = (data: V1ListResourcesResponse) =>
    // CAST SAFETY: must be a string[], because we filter
    // out undefined values
    (data?.resources
      ?.map((res) => res?.meta?.name?.name)
      .filter((name) => name !== undefined) ?? []) as string[];

  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select,
      },
    },
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
