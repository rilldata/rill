import type { Page } from "@sveltejs/kit";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export type DashboardInfo = {
  name: string;
  kind: ResourceKind;
};

export function getDashboardFromEmbedRoute(
  routeId: string | null,
  params: Record<string, string> | undefined,
): DashboardInfo | null {
  if (!routeId || !params?.name) return null;

  if (routeId.includes("/explore/")) {
    return { name: params.name, kind: ResourceKind.Explore };
  }

  if (routeId.includes("/canvas/")) {
    return { name: params.name, kind: ResourceKind.Canvas };
  }

  return null;
}

export function isEmbedDashboardPage(page: Page | null): boolean {
  if (!page) return false;
  return getDashboardFromEmbedRoute(page.route.id, page.params) !== null;
}

export function isDifferentDashboard(
  from: DashboardInfo | null,
  to: DashboardInfo | null,
): boolean {
  if (!from || !to) return false;
  return from.name !== to.name || from.kind !== to.kind;
}
