import {
  ResourceKind,
  useFilteredEntities,
  useFilteredEntityNames,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createQueryServiceTableColumns,
  createRuntimeServiceGetFile,
  V1ProfileColumn,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, Readable } from "svelte/store";
import { parse } from "yaml";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";

export type SourceFromYaml = {
  type: string;
  uri?: string;
  path?: string;
};

export function useSources(instanceId: string) {
  return useFilteredEntities(instanceId, ResourceKind.Source);
}

export function useSourceNames(instanceId: string) {
  return useFilteredEntityNames(instanceId, ResourceKind.Source);
}

export function useSource(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Source);
}

export function useSourceFromYaml(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => (data.blob ? parse(data.blob) : {}),
    },
  }) as CreateQueryResult<SourceFromYaml>;
}

export function useIsSourceUnsaved(
  instanceId: string,
  sourceName: string,
  // Include clientYAML in the function call to force the selector to recompute when it changes
  clientYAML: string
) {
  return createRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
    {
      query: {
        select: (data) => {
          const serverYAML = data.blob;
          return clientYAML !== serverYAML;
        },
      },
    }
  );
}

type TableColumnsWithName = {
  tableName: string;
  profileColumns: Array<V1ProfileColumn>;
};

export function useAllSourceColumns(
  instanceId: string
): Readable<Array<TableColumnsWithName>> {
  return derived([useSources(instanceId)], ([allSources], set) => {
    if (!allSources.data?.length) {
      set([]);
    }

    derived(
      allSources.data.map((r) =>
        createTableColumnsWithName(instanceId, r.meta.name.name)
      ),
      (sourceColumnResponses) =>
        sourceColumnResponses.filter((res) => !!res.data).map((res) => res.data)
    ).subscribe(set);
  });
}

/**
 * Fetches columns and adds the table name. By using the selector the results will be cached.
 */
function createTableColumnsWithName(
  instanceId: string,
  tableName: string
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
      },
    }
  );
}
