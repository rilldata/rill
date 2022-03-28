import { ActiveEventHandler } from "$lib/metrics/ActiveEventHandler";
import { config, metricsService } from "$lib/app-store";
import { collectCommonUserFields } from "$lib/metrics/collectCommonUserFields";

export let actionEvent: ActiveEventHandler;

export async function initMetrics() {
    const commonUserMetrics = await collectCommonUserFields();
    actionEvent = new ActiveEventHandler(config, metricsService, commonUserMetrics);
}
