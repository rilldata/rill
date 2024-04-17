import {
  createRuntimeServiceListFiles,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetFile,
  runtimeServiceListFiles,
  V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

/**
 * In dev mode we still need to get the files for left nav.
 * This is because parse errors will not have a resource.
 */
export function useMainEntityFiles(
  instanceId: string,
  prefix:
    | "sources"
    | "models"
    | "dashboards"
    | "charts"
    | "custom-dashboards"
    | "apis"
    | "themes"
    | "alerts"
    | "reports",
  transform = (name: string) => name,
) {
  let extension: string;
  switch (prefix) {
    case "apis":
    case "themes":
    case "alerts":
    case "reports":
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
      glob: "{apis,themes,alerts,reports,sources,models,dashboards,charts,custom-dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) => {
          // Filter the list of file paths to include only those that match the given prefix and extension
          const filteredPaths = data.files
            ?.filter((file) => {
              if (file.isDir) return false;
              // Match the filePath against a pattern to extract the directory name
              const regexMatch = file.path?.match(/\/([^/]+)\/[^/]+$/);
              // Check if the directory name exactly matches the prefix
              return regexMatch && regexMatch[1] === prefix;
            })
            .map((file) => {
              // Remove the directory and extension from the filePath to get the file name
              return transform(
                file.path?.replace(`/${prefix}/`, "").replace(extension, "") ??
                  "",
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

export async function fetchAllFiles(
  queryClient: QueryClient,
  instanceId: string,
) {
  const filesResp = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: () => {
      return runtimeServiceListFiles(instanceId, undefined);
    },
  });
  return filesResp.files ?? [];
}

export function useAllFileNames(queryClient: QueryClient, instanceId: string) {
  return createRuntimeServiceListFiles(instanceId, undefined, {
    query: {
      queryClient,
      select: (data) =>
        data.files
          ?.filter((f) => !f.isDir)
          .map((f) => f.path?.split("/").pop() ?? "") ?? [],
    },
  });
}

export async function fetchAllFileNames(
  queryClient: QueryClient,
  instanceId: string,
) {
  const files = await fetchAllFiles(queryClient, instanceId);
  return files
    .filter((f) => !f.isDir)
    .map((f) => f.path?.split("/").pop() ?? "")
    .filter(Boolean);
}

export async function fetchMainEntityFiles(
  queryClient: QueryClient,
  instanceId: string,
) {
  const files = await fetchAllFiles(queryClient, instanceId);
  return files
    .filter((f) => !f.isDir && fileIsMainEntity(f.path ?? ""))
    .map((f) => f.path ?? "");
}

export function fileIsMainEntity(filePath: string) {
  return (
    filePath.endsWith(".sql") ||
    filePath.endsWith(".yml") ||
    filePath.endsWith(".yaml")
  );
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

export function useFileNamesInDirectory(
  instanceId: string,
  directoryPath: string,
) {
  return createRuntimeServiceListFiles(
    instanceId,
    {
      glob: `${directoryPath}/*`,
    },
    {
      query: {
        select: (data) => {
          const files = data.files?.filter((file) => !file.isDir);
          const fileNames = files?.map((file) => {
            return file.path?.replace(`${directoryPath}/`, "") ?? "";
          });
          const sortedFileNames = fileNames?.sort((fileNameA, fileNameB) =>
            fileNameA.localeCompare(fileNameB, undefined, {
              sensitivity: "base",
            }),
          );
          return sortedFileNames ?? [];
        },
      },
    },
  );
}

export function useDirectoryNamesInDirectory(
  instanceId: string,
  directoryPath: string,
) {
  return createRuntimeServiceListFiles(
    instanceId,
    {
      glob: `${directoryPath}/*`,
    },
    {
      query: {
        select: (data) => {
          const files = data.files?.filter((file) => file.isDir);
          const directoryNames = files?.map((file) => {
            return file.path?.replace(`${directoryPath}/`, "") ?? "";
          });
          const sortedDirectoryNames = directoryNames?.sort(
            (dirNameA, dirNameB) =>
              dirNameA.localeCompare(dirNameB, undefined, {
                sensitivity: "base",
              }),
          );
          return sortedDirectoryNames ?? [];
        },
      },
    },
  );
}
