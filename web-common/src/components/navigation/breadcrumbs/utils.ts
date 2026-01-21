const variableRouteRegex = /\[.*]/;

/**
 * Gets sub route with prefix of path if it does not have any variables.
 * This helps to easily switch to another org/project but keep the same sub route like settings/alerts
 * @param path
 * @param route
 */
export function getNonVariableSubRoute(path: string, route: string) {
  const pathParts = path.split("/");
  const routeParts = route.split("/").slice(pathParts.length);
  if (
    // if any sub route has a variable then do not return sub route
    routeParts.some((p) => p.match(variableRouteRegex)) ||
    routeParts.length === 0
  ) {
    return "";
  }

  return "/" + routeParts.join("/");
}
