import type { SearchableItemType } from "./types";

export function buildRoute(
  type: SearchableItemType,
  orgName: string,
  projectName: string,
  resourceName: string,
): string {
  switch (type) {
    case "project":
      return `/${orgName}/${projectName}`;
    case "explore":
      return `/${orgName}/${projectName}/explore/${resourceName}`;
    case "canvas":
      return `/${orgName}/${projectName}/canvas/${resourceName}`;
    case "report":
      return `/${orgName}/${projectName}/-/reports/${resourceName}`;
    case "alert":
      return `/${orgName}/${projectName}/-/alerts/${resourceName}`;
  }
}
