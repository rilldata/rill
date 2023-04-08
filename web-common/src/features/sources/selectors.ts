import {
  createRuntimeServiceGetFile,
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { parse } from "yaml";

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
