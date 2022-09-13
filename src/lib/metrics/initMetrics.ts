import { ActiveEventHandler } from "$lib/metrics/ActiveEventHandler";
import {
  config,
  metricsService,
} from "$lib/application-state-stores/application-store";
import { collectCommonUserFields } from "$lib/metrics/collectCommonUserFields";
import { NavigationEventHandler } from "./NavigationEventHandler";

export let actionEvent: ActiveEventHandler;
export let navigationEvent: NavigationEventHandler;

export async function initMetrics() {
  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(
    config,
    metricsService,
    commonUserMetrics
  );

  navigationEvent = new NavigationEventHandler(commonUserMetrics);
}
