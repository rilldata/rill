const FILE_PATH_SPLIT_REGEX = /\//;
export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export function getSourceNameFromFile(filePath: string, name: string) {
  return name ?? sanitizeSourceName(extractSourceName(filePath));
}

export function extractSourceName(filePath: string): string {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.lastIndexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(0, lastIndexOfDot) : fileName;
}

export function extractFileExtension(filePath: string): string {
  const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
  const lastIndexOfDot = fileName.lastIndexOf(".");
  return lastIndexOfDot >= 0 ? fileName.substring(lastIndexOfDot + 1) : "";
}

export function sanitizeSourceName(sourceName: string): string {
  return sourceName.replace(INVALID_CHARS, "_");
}
