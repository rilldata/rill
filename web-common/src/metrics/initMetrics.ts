import { page } from "$app/stores";
import { BehaviourEventHandler } from "@rilldata/web-common/metrics/BehaviourEventHandler";
import { ErrorEventHandler } from "@rilldata/web-common/metrics/ErrorEventHandler";
import { mapScreenName } from "@rilldata/web-common/metrics/mapScreenName";
import { BehaviourEventFactory } from "@rilldata/web-common/metrics/service/BehaviourEventFactory";
import { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-common/metrics/service/ProductHealthEventFactory";
import { RillIntakeClient } from "@rilldata/web-common/metrics/service/RillIntakeClient";
import { GetMetadataResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import { get } from "svelte/store";
import { ActiveEventHandler } from "./ActiveEventHandler";
import { collectCommonUserFields } from "./collectCommonUserFields";
import { ErrorEventFactory } from "./service/ErrorEventFactory";

export let metricsService: MetricsService;

export let actionEvent: ActiveEventHandler;
export let behaviourEvent: BehaviourEventHandler;
export let errorEventHandler: ErrorEventHandler;

export async function initMetrics(localConfig: GetMetadataResponse) {
  metricsService = new MetricsService(new RillIntakeClient(), [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
    new ErrorEventFactory(),
  ]);
  metricsService.loadLocalFields(localConfig);

  const commonUserMetrics = await collectCommonUserFields();
  actionEvent = new ActiveEventHandler(metricsService, commonUserMetrics);
  behaviourEvent = new BehaviourEventHandler(metricsService, commonUserMetrics);
  errorEventHandler = new ErrorEventHandler(
    metricsService,
    commonUserMetrics,
    localConfig.isDev,
    () => mapScreenName(get(page)),
  );
}

// Setters used in cloud
export function setMetricsService(ms: MetricsService) {
  metricsService = ms;
}

export function setErrorEvent(ev: ErrorEventHandler) {
  errorEventHandler = ev;
}

export function setBehaviourEvent(ev: BehaviourEventHandler) {
  behaviourEvent = ev;
}
