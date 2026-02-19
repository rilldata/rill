/** Route prefixes that are exclusively preview mode */
export const PREVIEW_ROUTE_PREFIXES = [
  "/home",
  "/ai",
  "/preview",
  "/reports",
  "/alerts",
  "/status",
  "/settings",
] as const;

/** Route prefixes that are exclusively developer mode */
export const DEVELOPER_ROUTE_PREFIXES = ["/files"] as const;

/**
 * All route prefixes allowed in previewer mode (locked preview).
 * Includes preview routes plus shared routes (/explore, /canvas, /deploy).
 */
export const PREVIEWER_ALLOWED_PREFIXES = [
  ...PREVIEW_ROUTE_PREFIXES,
  "/explore/",
  "/canvas/",
  "/deploy",
] as const;

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
