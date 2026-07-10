import { goto } from "$app/navigation";
import { get, writable } from "svelte/store";

/**
 * Route prefix for the editing context.
 * Empty string for web-local (default), "/<org>/<project>/@<branch>/-/edit" for web-admin.
 */
export const editorRoutePrefix = writable("");

/**
 * Build a route path with the current edit prefix.
 * Example: withEditorPrefix("/files/models/foo.sql") → "/<org>/<project>/@<branch>/-/edit/files/models/foo.sql"
 */
export function withEditorPrefix(path: string): string {
  return `${get(editorRoutePrefix)}${path}`;
}

export function navigateToFile(
  filePath: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(withEditorPrefix(`/files${filePath}`), options);
}

export function getFileHref(filePath: string): string {
  return withEditorPrefix(`/files${filePath}`);
}

export function navigateToHome(options?: Parameters<typeof goto>[1]) {
  return goto(withEditorPrefix("/"), options);
}

export function getHomeHref(): string {
  return withEditorPrefix("/");
}

export function navigateToExplore(
  name: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(withEditorPrefix(`/explore/${name}`), options);
}

export function navigateToCanvas(
  name: string,
  options?: Parameters<typeof goto>[1],
) {
  return goto(withEditorPrefix(`/canvas/${name}`), options);
}
