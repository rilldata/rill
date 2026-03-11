import { goto } from "$app/navigation";
import { writable, get } from "svelte/store";

/**
 * Route prefix for workspace navigation.
 * Empty string for web-local (default), "/<org>/<project>/-/edit" for web-admin.
 */
export const workspaceRoutePrefix = writable("");

/**
 * Build a workspace route path with the current prefix.
 * Example: workspaceRoute("/files/models/foo.sql") → "/<org>/<project>/-/edit/files/models/foo.sql"
 */
export function workspaceRoute(path: string): string {
  return `${get(workspaceRoutePrefix)}${path}`;
}

export function navigateToFile(
  filePath: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(workspaceRoute(`/files${filePath}`), options);
}

export function getFileHref(filePath: string): string {
  return workspaceRoute(`/files${filePath}`);
}

export function navigateToHome(options?: Parameters<typeof goto>[1]) {
  return goto(workspaceRoute("/"), options);
}
