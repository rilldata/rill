import { BehaviourEventFactory } from "@rilldata/web-local/common/metrics-service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-local/common/metrics-service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-local/common/metrics-service/ProductHealthEventFactory";
import { RillIntakeClient } from "@rilldata/web-local/common/metrics-service/RillIntakeClient";
import { ActiveEventHandler } from "./ActiveEventHandler";
import { config } from "../application-state-stores/application-store";
import { collectCommonUserFields } from "./collectCommonUserFields";
import { NavigationEventHandler } from "./NavigationEventHandler";

export let metricsService: MetricsService;

export let actionEvent: ActiveEventHandler;
export let navigationEvent: NavigationEventHandler;

export async function initMetrics() {
  metricsService = new MetricsService(config, new RillIntakeClient(config), [
    new ProductHealthEventFactory(config),
    new BehaviourEventFactory(config),
  ]);
  await metricsService.loadCommonFields();

  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(
    config,
    metricsService,
    commonUserMetrics
  );
  navigationEvent = new NavigationEventHandler(
    metricsService,
    commonUserMetrics
  );
}
