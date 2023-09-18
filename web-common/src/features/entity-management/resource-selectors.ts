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

export function useResource(
  instanceId: string,
  name: string,
  kind: ResourceKind
) {
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

export function useFilteredEntityNames(instanceId: string, kind: ResourceKind) {
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
