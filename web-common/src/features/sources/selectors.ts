import {
  createRuntimeServiceGetFile,
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { parse } from "yaml";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";
import { useSourceStore } from "./sources-store";

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

export function useIsSourceNotSaved(instanceId: string, sourceName: string) {
  // Get serverYAML
  const file = createRuntimeServiceGetFile(
    instanceId,
    getFilePathFromNameAndType(sourceName, EntityType.Table)
  );
  const serverYAML = get(file).data?.blob;

  // Get clientYAML
  const sourceStore = useSourceStore();
  const clientYAML = get(sourceStore).clientYAML;

  // Compute difference
  // Note: if clientYAML is undefined, it means the source has not been touched
  const isContentUnsaved = clientYAML && clientYAML !== serverYAML;

  return isContentUnsaved;
}
