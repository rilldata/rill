import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { V1DirEntry } from "@rilldata/web-common/runtime-client";
import { QueryClient } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export interface Directory {
  name: string; // TODO: Remove 'name' field and instead compute it downstream from 'path'
  path: string;
  directories: Directory[];
  files: string[]; // TODO: Use file 'path' instead of 'name'; compute 'name' downstream
}

export function transformFileList(files: V1DirEntry[]): Directory {
  const rootDirectory: Directory = {
    name: "",
    path: "/",
    directories: [],
    files: [],
  };

  for (const file of files) {
    const parts = file.path?.split("/") ?? [];
    if (parts[0] === "") parts.shift(); // remove leading empty entry
    if (parts.length === 0) continue;

    const fileName = parts.pop();
    let currentDirectory = rootDirectory;

    parts.reduce((accPath, directoryName) => {
      const directoryPath = accPath
        ? `${accPath}/${directoryName}`
        : "/" + directoryName;
      let subDirectory = currentDirectory.directories.find(
        (dir) => dir.path === directoryPath,
      );

      if (!subDirectory) {
        subDirectory = {
          name: directoryName,
          path: directoryPath,
          directories: [],
          files: [],
        };
        currentDirectory.directories.push(subDirectory);
      }

      currentDirectory = subDirectory;
      return directoryPath;
    }, "");

    if (fileName) {
      if (file.isDir) {
        currentDirectory.directories.push({
          name: fileName,
          path: file.path ?? "",
          directories: [],
          files: [],
        });
      } else {
        currentDirectory.files.push(fileName);
      }
    }
  }

  return rootDirectory;
}

export function findDirectory(root: Directory, filePath: string) {
  const folderTree = removeLeadingSlash(filePath).split("/");
  let dir: Directory | undefined = root;
  for (let i = 0; i < folderTree.length && dir; i++) {
    dir = dir.directories.find((d) => d.name === folderTree[i]);
  }
  return dir;
}

export function getDirectoryHasErrors(
  queryClient: QueryClient,
  instanceId: string,
  dir: Directory,
) {
  return derived(
    flattenDirectory(dir).map((filePath) =>
      fileArtifacts
        .getFileArtifact(filePath)
        .getAllErrors(queryClient, instanceId),
    ),
    (filesErrors) => filesErrors.some((fileErrors) => fileErrors.length > 0),
  );
}
function flattenDirectory(dir: Directory) {
  const files = dir.files.map((f) => `${dir.path}/${f}`);
  dir.directories.forEach((subDir) => files.push(...flattenDirectory(subDir)));
  return files;
}
