export interface Directory {
  name: string; // TODO: Remove 'name' field and instead compute it downstream from 'path'
  path: string;
  directories: Directory[];
  files: string[]; // TODO: Use file 'path' instead of 'name'; compute 'name' downstream
}

export function transformFileList(filePaths: string[]): Directory {
  const rootDirectory: Directory = {
    name: "",
    path: "",
    directories: [],
    files: [],
  };

  for (const filePath of filePaths) {
    const parts = filePath.split("/");
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
      currentDirectory.files.push(fileName);
    }
  }

  return rootDirectory;
}
