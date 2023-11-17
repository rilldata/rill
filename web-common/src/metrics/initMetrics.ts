import { BehaviourEventHandler } from "@rilldata/web-common/metrics/BehaviourEventHandler";
import { ErrorEventHandler } from "@rilldata/web-common/metrics/ErrorEventHandler";
import { BehaviourEventFactory } from "@rilldata/web-common/metrics/service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-common/metrics/service/ProductHealthEventFactory";
import { RillAdminTelemetryClient } from "@rilldata/web-common/metrics/service/RillAdminTelemetryClient";
import { RillIntakeClient } from "@rilldata/web-common/metrics/service/RillIntakeClient";
import type { V1RuntimeGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { ActiveEventHandler } from "./ActiveEventHandler";
import { collectCommonUserFields } from "./collectCommonUserFields";
import { ErrorEventFactory } from "./service/ErrorEventFactory";

export let metricsService: MetricsService;

export let actionEvent: ActiveEventHandler;
export let behaviourEvent: BehaviourEventHandler;
export let errorEvent: ErrorEventHandler;

export async function initMetrics(localConfig: V1RuntimeGetConfig) {
  metricsService = new MetricsService(new RillIntakeClient(), [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
    new ErrorEventFactory(),
  ]);
  metricsService.loadLocalFields(localConfig);

  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(metricsService, commonUserMetrics);
  behaviourEvent = new BehaviourEventHandler(metricsService, commonUserMetrics);
  errorEvent = new ErrorEventHandler(metricsService, commonUserMetrics);
}

export async function initCloudMetrics() {
  metricsService = new MetricsService(new RillAdminTelemetryClient(), [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
    new ErrorEventFactory(),
  ]);

  const commonUserMetrics = await collectCommonUserFields();
  errorEvent = new ErrorEventHandler(metricsService, commonUserMetrics);
  // TODO: add other handlers and callers
}
