import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  useFilteredResourceNames,
  useFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";
import type { Readable } from "svelte/motion";
import { derived } from "svelte/store";
import {
  createTableColumnsWithName,
  type TableColumnsWithName,
} from "../sources/selectors";

export function useModels(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.Model, (data) =>
    data.resources?.filter((r) => !!r.model?.state?.table),
  );
}

export function useModelNames(instanceId: string) {
  return useFilteredResourceNames(instanceId, ResourceKind.Model);
}

export function useModelFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "models");
}

export function useModelRoutes(instanceId: string) {
  return useMainEntityFiles(instanceId, "models", (name) =>
    getRouteFromName(name, EntityType.Model),
  );
}

export function useModel(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Model);
}

export function useAllModelColumns(
  queryClient: QueryClient,
  instanceId: string,
): Readable<Array<TableColumnsWithName>> {
  return derived([useModels(instanceId)], ([allModels], set) => {
    if (!allModels.data?.length) {
      set([]);
      return;
    }

    derived(
      allModels.data.map((r) =>
        createTableColumnsWithName(
          queryClient,
          instanceId,
          r.meta?.name?.name ?? "",
        ),
      ),
      (modelColumnResponses) =>
        modelColumnResponses.filter((res) => !!res.data).map((res) => res.data),
    ).subscribe(set);
  });
}

export async function getModelNames(
  queryClient: QueryClient,
  instanceId: string,
) {
  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    }),
    queryFn: () => {
      return runtimeServiceListFiles(instanceId, {
        glob: "{sources,models,dashboards}/*.{yaml,sql}",
      });
    },
  });
  const modelNames = files.files
    ?.filter((f) => !f.isDir && f.path?.includes("models/"))
    .map((f) => f.path?.replace("/models/", "").replace(".sql", "") ?? "")
    // sort alphabetically case-insensitive
    .sort((a, b) => a.localeCompare(b, undefined, { sensitivity: "base" }));
  return modelNames ?? [];
}
