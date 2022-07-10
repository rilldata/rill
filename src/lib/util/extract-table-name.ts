const FILE_PATH_SPLIT_REGEX = /\//;
export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export function getTableNameFromFile(filePath: string, name?: string) {
  return name ?? sanitizeTableName(extractTableName(filePath));
}

export function extractTableName(filePath: string): string {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.lastIndexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(0, lastIndexOfDot) : fileName;
}

export function extractFileExtension(filePath: string): string {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.lastIndexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(lastIndexOfDot + 1) : "";
}

export function sanitizeTableName(tableName: string): string {
  return tableName.replace(INVALID_CHARS, "_");
}
