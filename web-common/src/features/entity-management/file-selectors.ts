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
  includeExtensions = true,
) {
  const files = await fetchAllFiles(queryClient, instanceId);
  return files
    .filter((f) => !f.isDir)
    .map((f) => f.path?.split("/").pop() ?? "")
    .map((fileName) => {
      if (!includeExtensions) {
        return fileName.split(".").slice(0, -1).join(".");
      }
      return fileName;
    })
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
  // Ensure the directory path starts with a slash
  if (!directoryPath.startsWith("/")) {
    directoryPath = `/${directoryPath}`;
  }

  return createRuntimeServiceListFiles(instanceId, undefined, {
    query: {
      select: (data) => {
        return useFileNamesInDirectorySelector(data, directoryPath);
      },
    },
  });
}

export function useFileNamesInDirectorySelector(
  data: V1ListFilesResponse,
  directoryPath: string,
) {
  if (!data.files) {
    return [];
  }

  const fileNames = data.files
    // Filter for files in the immediate directory
    .filter((file) => {
      if (!file.path) {
        return false;
      }

      const isNotDirectory = !file.isDir;
      const startsWithDirectory = file.path?.startsWith(directoryPath);
      const doesNotHaveSubdirectory =
        directoryPath === "/"
          ? file.path?.indexOf("/", 1) === -1
          : file.path?.lastIndexOf("/") === directoryPath.length;

      return isNotDirectory && startsWithDirectory && doesNotHaveSubdirectory;
    })

    // Remove the directory path from each file path
    .map((file) => {
      const startIdx = directoryPath === "/" ? 1 : directoryPath.length + 1;
      return file.path?.substring(startIdx) ?? "";
    })

    // Sort filenames alphabetically, case-insensitive
    .sort((fileNameA, fileNameB) =>
      fileNameA.localeCompare(fileNameB, undefined, {
        sensitivity: "base",
      }),
    );

  return fileNames;
}

export function useDirectoryNamesInDirectory(
  instanceId: string,
  directoryPath: string,
) {
  // Ensure the directory path starts with a slash
  if (!directoryPath.startsWith("/")) {
    directoryPath = `/${directoryPath}`;
  }

  return createRuntimeServiceListFiles(instanceId, undefined, {
    query: {
      select: (data) => {
        return useDirectoryNamesInDirectorySelector(data, directoryPath);
      },
    },
  });
}

export function useDirectoryNamesInDirectorySelector(
  data: V1ListFilesResponse,
  directoryPath: string,
) {
  if (!data.files) {
    return [];
  }

  const directoryNames = data.files
    // Filter for directories in the immediate directory
    .filter((file) => {
      if (!file.path) {
        return false;
      }

      const isDirectory = file.isDir;
      const startsWithDirectory = file.path?.startsWith(directoryPath);
      const existsAtSameDirectoryLevel =
        directoryPath === "/"
          ? file.path?.indexOf("/", 1) === -1
          : file.path?.lastIndexOf("/") === directoryPath.length;
      const isNotSameDirectory = file.path !== directoryPath;

      return (
        isDirectory &&
        startsWithDirectory &&
        existsAtSameDirectoryLevel &&
        isNotSameDirectory
      );
    })

    // Extract the directory name from the path
    .map((file) => {
      if (!file.path) {
        return "";
      }

      const startIdx = directoryPath.length + 1;
      return directoryPath === "/"
        ? file.path.substring(1)
        : file.path.substring(startIdx);
    })

    // Sort directory names alphabetically, case-insensitive
    .sort((dirNameA, dirNameB) => {
      if (!dirNameA || !dirNameB) return 0;
      return dirNameA.localeCompare(dirNameB, undefined, {
        sensitivity: "base",
      });
    });

  return directoryNames;
}
