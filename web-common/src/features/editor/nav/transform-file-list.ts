export interface Directory {
  name: string; // TODO: Remove 'name' field and instead compute it downstream from 'path'
  path: string;
  directories: Directory[];
  files: string[]; // TODO: Use file 'path' instead of 'name'; compute 'name' downstream
}

export function transformFileList(fileList: string[]): Directory {
  const rootDirectory: Directory = {
    name: "",
    path: "",
    directories: [],
    files: [],
  };

  for (const filePath of fileList) {
    console.log("filePath:", filePath);
    const directoryPath = filePath.split("/");
    const fileName = directoryPath.pop();
    console.log("fileName:", fileName);

    let currentDirectory = rootDirectory;
    for (const directoryName of directoryPath) {
      let subDirectory = currentDirectory.directories.find(
        (dir) => dir.name === directoryName
      );
      if (!subDirectory) {
        subDirectory = {
          name: directoryName,
          path: directoryPath.join("/"),
          directories: [],
          files: [],
        };
        currentDirectory.directories.push(subDirectory);
      }
      currentDirectory = subDirectory;
    }

    if (fileName) {
      currentDirectory.files.push(fileName);
    }
  }

  return rootDirectory;
}
