// Test-mode stub for the admin client. canvas-entity dynamically imports
// `@rilldata/web-admin/client` only in the cloud context; web-common unit tests cannot
// resolve that package, so this mock satisfies the import graph. See vite.config.ts.
export function getAdminServiceListBookmarksQueryOptions() {
  return {};
}
