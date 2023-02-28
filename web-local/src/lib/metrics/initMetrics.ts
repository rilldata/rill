import type { V1RuntimeGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { BehaviourEventHandler } from "@rilldata/web-local/lib/metrics/BehaviourEventHandler";
import { BehaviourEventFactory } from "@rilldata/web-local/lib/metrics/service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-local/lib/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-local/lib/metrics/service/ProductHealthEventFactory";
import { RillIntakeClient } from "@rilldata/web-local/lib/metrics/service/RillIntakeClient";
import { ActiveEventHandler } from "./ActiveEventHandler";
import { collectCommonUserFields } from "./collectCommonUserFields";

export let metricsService: MetricsService;

export let actionEvent: ActiveEventHandler;
export let behaviourEvent: BehaviourEventHandler;

export async function initMetrics(localConfig: V1RuntimeGetConfig) {
  metricsService = new MetricsService(localConfig, new RillIntakeClient(), [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
  ]);
  await metricsService.loadCommonFields();

  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(metricsService, commonUserMetrics);
  behaviourEvent = new BehaviourEventHandler(metricsService, commonUserMetrics);
}
