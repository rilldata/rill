import {
  createRuntimeServiceListFiles,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetFile,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

/**
 * In dev mode we still need to get the files for left nav.
 * This is because parse errors will not have a resource.
 */
export function useMainEntityFiles(
  instanceId: string,
  prefix: "sources" | "models" | "dashboards" | "charts" | "custom-dashboards",
  transform = (name: string) => name,
) {
  let extension: string;
  switch (prefix) {
    case "sources":
    case "dashboards":
    case "charts":
    case "custom-dashboards":
      extension = ".yaml";
      break;

    case "models":
      extension = ".sql";
  }

  return createRuntimeServiceListFiles(
    instanceId,
    {
      // We still use opinionated folder names. So we still need this filter
      glob: "{sources,models,dashboards,charts,custom-dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) => {
          // Filter the list of file paths to include only those that match the given prefix and extension
          const filteredPaths = data.paths
            ?.filter((filePath) => {
              // Match the filePath against a pattern to extract the directory name
              const regexMatch = filePath.match(/\/([^/]+)\/[^/]+$/);
              // Check if the directory name exactly matches the prefix
              return regexMatch && regexMatch[1] === prefix;
            })
            .map((filePath) => {
              // Remove the directory and extension from the filePath to get the file name
              return transform(
                filePath.replace(`/${prefix}/`, "").replace(extension, ""),
              );
            })
            // Sort the file names alphabetically in a case-insensitive manner
            .sort((fileNameA, fileNameB) =>
              fileNameA.localeCompare(fileNameB, undefined, {
                sensitivity: "base",
              }),
            );

          // Return the sorted list of file names or an empty array if there were no paths
          return filteredPaths ?? [];
        },
      },
    },
  );
}

export async function fetchMainEntityFiles(
  queryClient: QueryClient,
  instanceId: string,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, {
      glob: ".{yaml,sql}",
    }),
    queryFn: () =>
      runtimeServiceListFiles(instanceId, {
        glob: ".{yaml,sql}",
      }),
  });
  return resp.paths ?? [];
}

export async function fetchFileContent(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(instanceId, filePath),
    queryFn: () => runtimeServiceGetFile(instanceId, filePath),
  });
  return resp.blob ?? "";
}

const FILE_PATH_SPLIT_REGEX = /\//;
export function splitFolderAndName(
  filePath: string,
): [folder: string, fileName: string] {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  return [
    filePath.substring(0, filePath.length - fileName.length - 1),
    fileName,
  ];
}
