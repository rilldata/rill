import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { Page } from "@sveltejs/kit";

const exploreRouteRegex = /\/explore\/(?:\[name]|\[dashboard])/;
const canvasRouteRegex = /\/explore\/(?:\[name]|\[dashboard])/;

export function getDashboardResourceFromPage(pageLike: {
  params: Page["params"] | null;
  route: Page["route"] | null;
}) {
  const dashboardName =
    pageLike.params?.dashboard ?? pageLike.params?.name ?? "";
  const isExplore = pageLike.route?.id
    ? exploreRouteRegex.test(pageLike.route.id)
    : false;
  const isCanvas = pageLike.route?.id
    ? canvasRouteRegex.test(pageLike.route.id)
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
