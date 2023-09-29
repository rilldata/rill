import {
  createRuntimeServiceGetFile,
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { parse } from "yaml";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";

/**
 * Calls {@link createRuntimeServiceListFiles} using glob to select only sources.
 * Returns just the source names from the files.
 */
export function useSourceNames(instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        // refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("sources/"))
            .map((path) => path.replace("/sources/", "").replace(".yaml", ""))
            // sort alphabetically case-insensitive
            .sort((a, b) =>
              a.localeCompare(b, undefined, { sensitivity: "base" })
            ),
      },
    }
  );
}

export type SourceFromYaml = {
  type: string;
  uri?: string;
  path?: string;
};

export function useSourceFromYaml(instanceId: string, filePath: string) {
  return createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => (data.blob ? parse(data.blob) : {}),
    },
  }) as CreateQueryResult<SourceFromYaml>;
}

export function useEmbeddedSources(instanceId: string) {
  return createRuntimeServiceListCatalogEntries(
    instanceId,
    {},
    {
      query: {
        select: (data) =>
          data?.entries?.filter(
            (catalog) => catalog.embedded && catalog.source
          ) ?? [],
      },
    }
  );
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
/**
 * This client-side YAML parsing is a rudimentary hack to check if the source is a local file.
 */
export function useIsLocalFileConnector(
  instanceId: string,
  sourceName: string
) {
  return createRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table),
    {
      query: {
        select: (data) => {
          const serverYAML = data.blob;
          const yaml = parse(serverYAML);
          // Check that the `type` is `duckdb` and that the `sql` includes 'data/'
          return yaml?.type === "duckdb" && yaml?.sql?.includes("data/");
        },
      },
    }
  );
}
