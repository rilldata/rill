export function getPaddingFromPath(filePath: string) {
  // Root level is 0; each "/" in the path represents a level deeper
  const level = filePath === "" ? 0 : filePath.split("/").length;
  return 8 + (level - 1) * 14;
}
