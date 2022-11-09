const FILE_PATH_SPLIT_REGEX = /\//;
export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export function getTableNameFromFile(filePath: string, name?: string) {
  return name ?? sanitizeEntityName(extractTableName(filePath));
}

export function extractTableName(filePath: string): string {
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
  const lastIndexOfDot = fileName.lastIndexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(lastIndexOfDot + 1) : "";
}

export function sanitizeEntityName(entityName: string): string {
  return entityName.replace(INVALID_CHARS, "_");
}
