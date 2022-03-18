import { metricsFactory } from "$lib/metrics/metricsFactory";
import { RootConfig } from "$common/config/RootConfig";
import { ActiveEventHandler } from "$lib/metrics/ActiveEventHandler";

export let actionEvent: ActiveEventHandler;

export function initMetrics() {
    const metricsService = metricsFactory(new RootConfig({}));
    actionEvent = new ActiveEventHandler(metricsService);
}
