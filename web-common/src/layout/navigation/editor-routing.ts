import { goto } from "$app/navigation";
import { writable, get } from "svelte/store";

/**
 * Route prefix for the editing context.
 * Empty string for web-local (default), "/<org>/<project>/@<branch>/-/edit" for web-admin.
 */
export const editRoutePrefix = writable("");

/**
 * Build a route path with the current edit prefix.
 * Example: editRoute("/files/models/foo.sql") → "/<org>/<project>/@<branch>/-/edit/files/models/foo.sql"
 */
export function editRoute(path: string): string {
  return `${get(editRoutePrefix)}${path}`;
}

export function navigateToFile(
  filePath: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(editRoute(`/files${filePath}`), options);
}

export function getFileHref(filePath: string): string {
  return editRoute(`/files${filePath}`);
}

export function navigateToHome(options?: Parameters<typeof goto>[1]) {
  return goto(editRoute("/"), options);
}

export function navigateToExplore(
  name: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(editRoute(`/explore/${name}`), options);
}

export function navigateToCanvas(
  name: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(editRoute(`/canvas/${name}`), options);
}
