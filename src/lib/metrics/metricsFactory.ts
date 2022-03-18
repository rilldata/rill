import { ProductHealthEventFactory } from "$common/metrics/ProductHealthEventFactory";
import type { RootConfig } from "$common/config/RootConfig";
import { MetricsService } from "$common/metrics/MetricsService";
import { RillIntakeClient } from "$common/metrics/RillIntakeClient";
import { collectCommonMetricsFields } from "$lib/metrics/collectCommonMetricsFields";

export function metricsFactory(config: RootConfig) {
    const metricsService = new MetricsService(
        config, null,
        new RillIntakeClient(config), [
            new ProductHealthEventFactory(config),
        ]);
    collectCommonMetricsFields()
        .then(fields => metricsService.setCommonMetricsInput(fields));
    return metricsService;
}
