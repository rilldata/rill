import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { Page } from "@sveltejs/kit";

const exploreRouteRegex = /\/explore\/(?:\[name]|\[dashboard])/;
const canvasRouteRegex = /\/explore\/(?:\[name]|\[dashboard])/;

export function getDashboardResourceFromPage(pageState: Page) {
  const dashboardName =
    pageState.params.dashboard ?? pageState.params.name ?? "";
  const isExplore = pageState.route?.id
    ? exploreRouteRegex.test(pageState.route.id)
    : false;
  const isCanvas = pageState.route?.id
    ? canvasRouteRegex.test(pageState.route.id)
    : false;

  if (isExplore) {
    return {
      name: dashboardName,
      kind: ResourceKind.Explore,
    };
  } else if (isCanvas) {
    return {
      name: dashboardName,
      kind: ResourceKind.Canvas,
    };
  } else {
    return null;
  }
}
