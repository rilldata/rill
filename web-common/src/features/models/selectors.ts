import {
  ResourceKind,
  useClientFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/query-core";
import type { Readable } from "svelte/motion";
import { derived } from "svelte/store";
import {
  createTableColumnsWithName,
  type TableColumnsWithName,
} from "../sources/selectors";

export function useModels(client: RuntimeClient) {
  return useClientFilteredResources(
    client,
    ResourceKind.Model,
    (res) =>
      res.meta?.name?.name === res.model?.state?.resultTable &&
      !res.model?.spec?.definedAsSource,
  );
}

export function useModel(client: RuntimeClient, name: string) {
  return useResource(client, name, ResourceKind.Model);
}

export function useAllModelColumns(
  queryClient: QueryClient,
  client: RuntimeClient,
): Readable<Array<TableColumnsWithName>> {
  return derived([useModels(client)], ([allModels], set) => {
    if (!allModels.data?.length) {
      set([]);
      return;
    }

    derived(
      allModels.data.map((r) =>
        createTableColumnsWithName(
          queryClient,
          client,
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
