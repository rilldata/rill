/**
 * Returns the appropriate route based on whether preview mode is active.
 * Centralizes all preview vs developer route logic for testability.
 */

function addLeadingSlash(path: string): string {
  if (path.startsWith("/")) return path;
  return "/" + path;
}

export function getHomeRoute(isPreview: boolean): string {
  return isPreview ? "/dashboards" : "/";
}

export function getExploreRoute(
  isPreview: boolean,
  name: string,
  filePath: string,
): string {
  return isPreview ? `/explore/${name}` : `/files${addLeadingSlash(filePath)}`;
}

export function getCanvasRoute(
  isPreview: boolean,
  name: string,
  filePath: string,
): string {
  return isPreview ? `/canvas/${name}` : `/files${addLeadingSlash(filePath)}`;
}

export function getFileRoute(isPreview: boolean, filePath: string): string {
  return isPreview ? "/dashboards" : `/files${addLeadingSlash(filePath)}`;
}
