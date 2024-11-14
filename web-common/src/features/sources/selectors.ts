import {
  ResourceKind,
  useClientFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  type V1ProfileColumn,
  createQueryServiceTableColumns,
  createRuntimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult, QueryClient } from "@tanstack/svelte-query";
import { type Readable, derived } from "svelte/store";
import { parse } from "yaml";

export type SourceFromYaml = {
  type: string;
  uri?: string;
  path?: string;
};

export function useSources(instanceId: string) {
  return useClientFilteredResources(
    instanceId,
    ResourceKind.Source,
    (res) => !!res.source?.state?.table,
  );
}

export function useSourceFromYaml(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: filePath },
    {
      query: {
        select: (data) => (data.blob ? parse(data.blob) : {}),
      },
    },
  ) as CreateQueryResult<SourceFromYaml>;
}

/**
 * This client-side YAML parsing is a rudimentary hack to check if the source is a local file.
 */
export function useIsLocalFileConnector(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: filePath },
    {
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
          !!filePath &&
          (filePath.endsWith(".yaml") || filePath.endsWith(".yml")),
      },
    },
  );
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
        createTableColumnsWithName(
          queryClient,
          instanceId,
          r.source?.state?.connector ?? "",
          "",
          "",
          r.meta?.name?.name ?? "",
        ),
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
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
) {
  return createQueryServiceTableColumns(
    instanceId,
    tableName,
    {
      connector,
      database,
      databaseSchema,
    },
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
