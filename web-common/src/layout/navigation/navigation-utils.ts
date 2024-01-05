import { appScreen } from "@rilldata/web-common/layout/app-store";
import { get } from "svelte/store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";

export function getNavURLToScreenMap(href: string) {
  if (href.includes("/source/")) return MetricsEventScreenName.Source;
  if (href.includes("/model/")) return MetricsEventScreenName.Model;
  if (href.includes("/dashboard/")) return MetricsEventScreenName.Dashboard;
}

export function emitNavigationTelemetry(href) {
  const previousActiveEntity = get(appScreen)?.type;
  const screenName = getNavURLToScreenMap(href);
  behaviourEvent.fireNavigationEvent(
    name,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
    previousActiveEntity,
    screenName,
  );
}
