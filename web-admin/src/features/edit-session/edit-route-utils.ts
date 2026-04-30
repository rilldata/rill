/**
 * Path segments under `/-/edit/` that render the Preview surface
 * (read-only dashboards listing and dashboard pages). Anything not in
 * this list — including the bare `/-/edit` root and `/-/edit/files/*`
 * — is treated as Developer mode (file editor). The `dashboards` arm
 * is anchored on a trailing `/` or end-of-string so a future sibling
 * route like `/-/edit/dashboards-archive` wouldn't false-match.
 */
const EDIT_PREVIEW_PATTERN = /\/-\/edit\/(dashboards(\/|$)|explore\/|canvas\/)/;

export function isEditPreviewRoute(pathname: string): boolean {
  return EDIT_PREVIEW_PATTERN.test(pathname);
}
