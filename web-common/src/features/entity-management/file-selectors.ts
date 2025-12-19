import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  createRuntimeServiceListFiles,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
  type V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export function useAllFileNames(queryClient: QueryClient, instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
    undefined,
    {
      query: {
        select: (data) =>
          data.files
            ?.filter((f) => !f.isDir)
            .map((f) => f.path?.split("/").pop() ?? "") ?? [],
      },
    },
    queryClient,
  );
}

export function fileIsMainEntity(filePath: string) {
  return (
    filePath.endsWith(".sql") ||
    filePath.endsWith(".yml") ||
    filePath.endsWith(".yaml")
  );
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

export async function getFileNamesInDirectory(
  queryClient: QueryClient,
  instanceId: string,
  directoryPath: string,
) {
  // Ensure the directory path starts with a slash
  if (!directoryPath.startsWith("/")) {
    directoryPath = `/${directoryPath}`;
  }

  // Fetch all files in the project
  // (For now, we fetch all files at once, rather than individual requests for each directory.)
  const allFiles = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) =>
      runtimeServiceListFiles(instanceId, undefined, signal),
  });

  // Get the file names in the given directory
  return useFileNamesInDirectorySelector(allFiles, directoryPath);
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

export const GithubSizeLimitInBytes = 100 * 1024 * 1024; // 100MB limit
// export const GithubSizeLimitInBytes = 100;
export function getFilesExceedingGithubPushLimit(instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
    undefined,
    {
      query: {
        select: (data) =>
          data.files?.filter(
            (f) =>
              !f.isDir && f.size && Number(f.size) > GithubSizeLimitInBytes,
          ) ?? [],
      },
    },
    queryClient,
  );
}

export function maybeSendLargeFileNotification(file: V1WatchFilesResponse) {
  const size = Number(file.size ?? "0");
  if (size < GithubSizeLimitInBytes) return;

  eventBus.emit("notification", {
    type: "default",
    message: `A file ${file.path} (${formatMemorySize(size)}) uploaded is too large for deploy. Please upload to s3 or similar service before deploying this project.`,
    options: {
      persisted: true,
    },
  });
}
