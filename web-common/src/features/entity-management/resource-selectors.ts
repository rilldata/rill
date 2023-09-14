import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";

export enum ResourceKind {
  Source = "source",
  Model = "model",
  MetricsView = "metricsview",
  // TODO: do a correct map based on backend code
}

function useResource(instanceId: string, name: string, kind: ResourceKind) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": kind,
      "name.name": name,
    },
    {
      query: {
        select: (data) => data?.resource,
      },
    }
  );
}
export function useSource(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Source);
}
export function useModel(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Model);
}

function useFilteredEntityNames(instanceId: string, kind: ResourceKind) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind,
    },
    {
      query: {
        select: (data) => data.resources.map((res) => res.meta.name.name),
      },
    }
  );
}
export function useSourceNames(instanceId: string) {
  return useFilteredEntityNames(instanceId, ResourceKind.Source);
}
export function useModelNames(instanceId: string) {
  return useFilteredEntityNames(instanceId, ResourceKind.Model);
}
export function useDashboardNames(instanceId: string) {
  return useFilteredEntityNames(instanceId, ResourceKind.MetricsView);
}

// TODO: replace usage of this with appropriate ones for de-duping names
export function useAllEntityNames(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select: (data) => data.resources.map((res) => res.meta.name.name),
      },
    }
  );
}
