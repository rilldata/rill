/**
 * Route prefixes that are exclusively preview mode.
 * /explore and /canvas live here (not as shared routes) so any visit to a
 * dashboard URL forces preview mode — there is no "developer view" of a
 * dashboard.
 */
export const PREVIEW_ROUTE_PREFIXES = [
  "/dashboards",
  "/ai",
  "/status",
  "/explore/",
  "/canvas/",
] as const;

/** Route prefixes that are exclusively developer mode */
export const DEVELOPER_ROUTE_PREFIXES = ["/files"] as const;

/**
 * All route prefixes allowed in preview mode.
 * Includes preview routes plus shared routes (/welcome, /deploy, /-/).
 */
export const PREVIEW_ALLOWED_PREFIXES = [
  ...PREVIEW_ROUTE_PREFIXES,
  "/welcome",
  "/deploy",
  "/-/",
] as const;

/**
 * Note: isPreviewRoute and isDeveloperRoute are intentionally not exhaustive.
 * Routes like /welcome, /deploy, and /-/ are shared between modes and match
 * neither; they preserve the current mode without triggering a switch.
 * When adding new routes, decide whether they belong to a specific mode or
 * are shared, and update the appropriate prefix list if needed.
 */
export function isPreviewRoute(pathname: string): boolean {
  return PREVIEW_ROUTE_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

export function isDeveloperRoute(pathname: string): boolean {
  return (
    pathname === "/" ||
    DEVELOPER_ROUTE_PREFIXES.some((prefix) => pathname.startsWith(prefix))
  );
}

/** Whether preview-mode nav bar should be visible on this path */
export function showPreviewNav(pathname: string): boolean {
  return (
    !pathname.startsWith("/files") &&
    !pathname.startsWith("/explore") &&
    !pathname.startsWith("/canvas")
  );
}
