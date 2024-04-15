const FILE_PATH_SPLIT_REGEX = /\//;
export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export function getTableNameFromFile(filePath: string, name?: string) {
  return name ?? sanitizeEntityName(extractFileName(filePath));
}

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

export function sanitizeEntityName(entityName: string): string {
  return entityName.replace(INVALID_CHARS, "_");
}

export function splitFolderAndName(
  filePath: string,
): [folder: string, fileName: string] {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  return [
    filePath.substring(0, filePath.length - fileName.length - 1),
    fileName,
  ];
}

export function getTopLevelFolder(filePath: string): string {
  return "/" + filePath.split(FILE_PATH_SPLIT_REGEX)[1];
}
