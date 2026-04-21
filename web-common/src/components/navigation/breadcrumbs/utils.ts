const VARIABLE_SEGMENT_RE = /\[.*]/;

/**
 * Given a new URL prefix (e.g. a different org, project, or section) and
 * the current page's route id, returns the portion of the current sub-route
 * that can be safely carried over to the new prefix — so a user switching
 * projects in a breadcrumb stays on the same sub-page (e.g. settings, alerts).
 *
 * Returns "" when the sub-route shouldn't follow the user.
 */
export function getCarryOverSubRoute(
  newPrefix: string,
  currentRoute: string,
): string {
  const subRoute = currentRoute.split("/").slice(newPrefix.split("/").length);

  // No sub-route past the new prefix.
  if (subRoute.length === 0) return "";

  // Route references a specific resource (e.g. `[name]`) that likely
  // doesn't exist under the new prefix.
  if (subRoute.some((p) => VARIABLE_SEGMENT_RE.test(p))) return "";

  // Edit sessions are tied to the current project's dev deployment.
  if (subRoute[0] === "-" && subRoute[1] === "edit") return "";

  return "/" + subRoute.join("/");
}
