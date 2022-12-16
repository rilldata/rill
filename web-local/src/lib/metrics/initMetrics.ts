import type { V1RuntimeGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { BehaviourEventFactory } from "@rilldata/web-local/lib/metrics/service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-local/lib/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-local/lib/metrics/service/ProductHealthEventFactory";
import { RillIntakeClient } from "@rilldata/web-local/lib/metrics/service/RillIntakeClient";
import { ActiveEventHandler } from "./ActiveEventHandler";
import { collectCommonUserFields } from "./collectCommonUserFields";
import { NavigationEventHandler } from "./NavigationEventHandler";

export let metricsService: MetricsService;

export let actionEvent: ActiveEventHandler;
export let navigationEvent: NavigationEventHandler;

export async function initMetrics(localConfig: V1RuntimeGetConfig) {
  metricsService = new MetricsService(localConfig, new RillIntakeClient(), [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
  ]);
  await metricsService.loadCommonFields();

  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(metricsService, commonUserMetrics);
  navigationEvent = new NavigationEventHandler(
    metricsService,
    commonUserMetrics
  );
}
