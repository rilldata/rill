/** Route prefixes that are exclusively preview mode */
export const PREVIEW_ROUTE_PREFIXES = [
  "/dashboards",
  "/ai",
  "/status",
] as const;

/** Route prefixes that are exclusively developer mode */
export const DEVELOPER_ROUTE_PREFIXES = ["/files"] as const;

/**
 * All route prefixes allowed in preview mode.
 * Includes preview routes plus shared routes (/explore, /canvas, /deploy).
 */
export const PREVIEW_ALLOWED_PREFIXES = [
  ...PREVIEW_ROUTE_PREFIXES,
  "/welcome",
  "/explore/",
  "/canvas/",
  "/deploy",
  "/settings",
  "/-/",
] as const;

/**
 * All route prefixes allowed in developer mode.
 * Includes developer routes, root, and shared routes (/explore, /canvas, /deploy).
 */
export const DEVELOPER_ALLOWED_PREFIXES = [
  ...DEVELOPER_ROUTE_PREFIXES,
  "/welcome",
  "/explore/",
  "/canvas/",
  "/deploy",
  "/connector/",
  "/graph",
  "/settings",
  "/-/",
] as const;

/**
 * Note: isPreviewRoute and isDeveloperRoute are intentionally not exhaustive.
 * Routes like /explore, /canvas, and /deploy are shared between modes and
 * match neither; they preserve the current mode without triggering a switch.
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
