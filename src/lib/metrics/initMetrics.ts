import { ActiveEventHandler } from "$lib/metrics/ActiveEventHandler";
import { config, metricsService } from "$lib/application-state-stores/application-store";
import { collectCommonUserFields } from "$lib/metrics/collectCommonUserFields";

export let actionEvent: ActiveEventHandler;

export async function initMetrics() {
    const commonUserMetrics = await collectCommonUserFields();
    actionEvent = new ActiveEventHandler(config, metricsService, commonUserMetrics);
}
