const FILE_PATH_SPLIT_REGEX = /\//;

export function extractFileName(filePath: string): string {
  let fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.lastIndexOf(".");
  fileName =
    lastIndexOfDot >= 0 ? fileName.substring(0, lastIndexOfDot) : fileName;

  // preappend underscore in case table name starts with hypen or number
  if (fileName.match(/^(\d|-)/)) {
    fileName = fileName.replace(/^-/, "");
    fileName = "_" + fileName;
  }

  return fileName;
}

export function extractFileExtension(filePath: string): string {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.indexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(lastIndexOfDot) : "";
}

export function splitFolderAndFileName(
  filePath: string,
): [folder: string, fileName: string] {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  return [
    filePath.substring(0, filePath.length - fileName.length - 1),
    fileName,
  ];
}

export function splitFolderFileNameAndExtension(
  filePath: string,
): [folder: string, fileName: string, extension: string] {
  const [folder, fullFileName] = splitFolderAndFileName(filePath);

  const lastIndexOfDot = fullFileName.lastIndexOf(".");
  const fileName =
    lastIndexOfDot >= 0
      ? fullFileName.substring(0, lastIndexOfDot)
      : fullFileName;
  const extension =
    lastIndexOfDot >= 0 ? fullFileName.substring(lastIndexOfDot) : "";

  return [folder, fileName, extension];
}

export function getTopLevelFolder(filePath: string): string {
  return "/" + (filePath.split(FILE_PATH_SPLIT_REGEX)[1] ?? "");
}
