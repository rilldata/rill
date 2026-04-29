/**
 * Path segments under `/-/edit/` that render the Preview surface
 * (read-only dashboards, AI chat, project status). Anything not in this
 * list — including the bare `/-/edit` root and `/-/edit/files/*` — is
 * treated as Developer mode (file editor).
 */
const EDIT_PREVIEW_PATTERN =
  /\/-\/edit\/(dashboards|explore\/|canvas\/|ai|status)/;

export function isEditPreviewRoute(pathname: string): boolean {
  return EDIT_PREVIEW_PATTERN.test(pathname);
}
