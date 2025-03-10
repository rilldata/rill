import {
  ResourceKind,
  useClientFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { QueryClient } from "@tanstack/query-core";
import type { Readable } from "svelte/motion";
import { derived } from "svelte/store";
import {
  createTableColumnsWithName,
  type TableColumnsWithName,
} from "../sources/selectors";

export function useModels(instanceId: string) {
  return useClientFilteredResources(
    instanceId,
    ResourceKind.Model,
    (res) =>
      res.meta?.name?.name === res.model?.state?.resultTable &&
      !res.model?.spec?.definedAsSource,
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
          r.model?.state?.resultConnector ?? "",
          "",
          "",
          r.meta?.name?.name ?? "",
        ),
      ),
      (modelColumnResponses) =>
        modelColumnResponses.filter((res) => !!res.data).map((res) => res.data),
    ).subscribe(set);
  });
}
