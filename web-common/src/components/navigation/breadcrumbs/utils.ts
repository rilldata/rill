import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types.ts";

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

export function makePath(
  current: (string | undefined)[],
  depth: number,
  id: string,
  option: PathOption,
  route: string,
  hasSelect: boolean,
) {
  if (hasSelect) return undefined;
  if (option?.href) return option.href;

  const newPath = current
    .slice(0, option?.depth ?? depth)
    .filter((p): p is string => !!p);

  if (option?.section) newPath.push(option.section);

  newPath.push(id);
  const path = `/${newPath.join("/")}`;
  return path + getNonVariableSubRoute(path, route);
}
