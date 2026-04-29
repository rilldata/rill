/**
 * Path segments under `/-/edit/` that render the Preview surface
 * (read-only dashboards, AI chat, project status). Anything not in this
 * list — including the bare `/-/edit` root and `/-/edit/files/*` — is
 * treated as Developer mode (file editor).
 */
const EDIT_PREVIEW_PATTERN =
  /\/-\/edit\/(dashboards|explore\/|canvas\/|ai|status)/;

/**
 * Subset of preview routes where the top-level preview nav (Dashboards /
 * AI / Status tabs) is shown. Hidden on individual dashboard pages
 * (explore/canvas) so the dashboard owns the full viewport, mirroring
 * web-local's `showPreviewNav`.
 */
const EDIT_PREVIEW_NAV_PATTERN = /\/-\/edit\/(dashboards|ai|status)/;

export function isEditPreviewRoute(pathname: string): boolean {
  return EDIT_PREVIEW_PATTERN.test(pathname);
}

export function showEditPreviewNav(pathname: string): boolean {
  return EDIT_PREVIEW_NAV_PATTERN.test(pathname);
}
