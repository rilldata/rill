import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  V1ProfileColumn,
  createQueryServiceTableColumns,
  createRuntimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult, QueryClient } from "@tanstack/svelte-query";
import { Readable, derived } from "svelte/store";
import { parse } from "yaml";

export type SourceFromYaml = {
  type: string;
  uri?: string;
  path?: string;
};

export function useSources(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.Source, (data) =>
    data.resources?.filter((r) => !!r.source?.state?.table),
  );
}

export function useSourceFromYaml(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => (data.blob ? parse(data.blob) : {}),
    },
  }) as CreateQueryResult<SourceFromYaml>;
}

/**
 * This client-side YAML parsing is a rudimentary hack to check if the source is a local file.
 */
export function useIsLocalFileConnector(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => {
        const serverYAML = data.blob;
        if (!serverYAML) return false;
        const yaml = parse(serverYAML);
        // Check that the `type` is `duckdb` and that the `sql` includes 'data/'
        return Boolean(
          yaml?.type === "duckdb" && yaml?.sql?.includes("'data/"),
        );
      },
      enabled:
        !!filePath && (filePath.endsWith(".yaml") || filePath.endsWith(".yml")),
    },
  });
}

export type TableColumnsWithName = {
  tableName: string;
  profileColumns: Array<V1ProfileColumn>;
};

export function useAllSourceColumns(
  queryClient: QueryClient,
  instanceId: string,
): Readable<Array<TableColumnsWithName>> {
  return derived([useSources(instanceId)], ([allSources], set) => {
    if (!allSources.data?.length) {
      set([]);
      return;
    }

    derived(
      allSources.data.map((r) =>
        createTableColumnsWithName(queryClient, instanceId, r.meta.name.name),
      ),
      (sourceColumnResponses) =>
        sourceColumnResponses
          .filter((res) => !!res.data)
          .map((res) => res.data),
    ).subscribe(set);
  });
}

/**
 * Fetches columns and adds the table name. By using the selector the results will be cached.
 */
export function createTableColumnsWithName(
  queryClient: QueryClient,
  instanceId: string,
  tableName: string,
): CreateQueryResult<TableColumnsWithName> {
  return createQueryServiceTableColumns(
    instanceId,
    tableName,
    {},
    {
      query: {
        select: (data) => {
          return {
            tableName,
            profileColumns: data.profileColumns,
          };
        },
        queryClient,
      },
    },
  );
}
