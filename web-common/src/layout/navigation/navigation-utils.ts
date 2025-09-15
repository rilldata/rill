import { appScreen } from "@rilldata/web-common/layout/app-store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";
import { get } from "svelte/store";

export function getNavURLToScreenMap(href: string) {
  if (href.includes("/source/")) return MetricsEventScreenName.Source;
  if (href.includes("/model/")) return MetricsEventScreenName.Model;
  if (href.includes("/dashboard/")) return MetricsEventScreenName.Dashboard;
}

export async function emitNavigationTelemetry(href: string, name: string) {
  const previousActiveEntity = get(appScreen).type;
  const screenName = getNavURLToScreenMap(href);

  if (!screenName) return;
  await behaviourEvent?.fireNavigationEvent(
    name,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
    previousActiveEntity,
    screenName,
  );
}

export function isEmbedPage(page: Page): boolean {
  if (!page.route.id) return false;
  return page.route.id.startsWith("/-/embed");
}
