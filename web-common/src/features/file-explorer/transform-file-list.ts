import { V1DirEntry } from "@rilldata/web-common/runtime-client";

export interface Directory {
  name: string; // TODO: Remove 'name' field and instead compute it downstream from 'path'
  path: string;
  directories: Directory[];
  files: string[]; // TODO: Use file 'path' instead of 'name'; compute 'name' downstream
}

export function transformFileList(files: V1DirEntry[]): Directory {
  const rootDirectory: Directory = {
    name: "",
    path: "",
    directories: [],
    files: [],
  };

  for (const file of files) {
    const parts = file.path?.split("/") ?? [];
    if (parts.length === 0) continue;

    const fileName = parts.pop();
    let currentDirectory = rootDirectory;

    parts.reduce((accPath, directoryName) => {
      const directoryPath = accPath
        ? `${accPath}/${directoryName}`
        : directoryName;
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
