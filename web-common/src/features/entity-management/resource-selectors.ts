import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListFiles,
  createRuntimeServiceListResources,
  V1ListResourcesResponse,
  V1Resource,
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

export function useFilteredEntities<T = Array<V1Resource>>(
  instanceId: string,
  kind: ResourceKind,
  selector?: (data: V1ListResourcesResponse) => T
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
    }
  );
}

export function useFilteredEntityNames(instanceId: string, kind: ResourceKind) {
  return useFilteredEntities<Array<string>>(instanceId, kind, (data) =>
    data.resources.map((res) => res.meta.name.name)
  );
}

export function useAllNames(instanceId: string) {
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
