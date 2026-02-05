import { page } from "$app/stores";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { Page } from "@sveltejs/kit";
import { derived } from "svelte/store";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { createQuery } from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

const exploreRouteRegex = /\/explore\/(?:\[name]|\[dashboard])/;
const canvasRouteRegex = /\/canvas\/(?:\[name]|\[dashboard])/;

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

export function getExploreNameStore() {
  return derived(page, (pageState) => {
    const dashboardResource = getDashboardResourceFromPage(pageState);
    if (dashboardResource?.kind !== ResourceKind.Explore) return "";
    return dashboardResource.name;
  });
}

export function getActiveMetricsViewNameStore() {
  const exploreNameStore = getExploreNameStore();
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
    queryClient,
  );

  return derived(
    validSpecQuery,
    (validSpecSpec) => validSpecSpec.data?.exploreSpec?.metricsView ?? "",
  );
}
