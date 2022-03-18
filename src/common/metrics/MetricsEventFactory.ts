import type { RootConfig } from "$common/config/RootConfig";
import type { CommonMetricsFields, MetricsEvent } from "$common/metrics/MetricsTypes";

export class MetricsEventFactory {
    public constructor(protected readonly config: RootConfig) {}

    protected getBaseMetricsEvent(commonMetricsInput: CommonMetricsFields,
                                eventType: string): MetricsEvent {
        return {
            ...commonMetricsInput,
            event_datetime: Date.now(),
            event_type: eventType,
            app_name: this.config.metrics.appName,

            // TODO
            install_id: "",
            build_id: "",
            version: "",
            project_id: "",
            model_id: "",
        };
    }
}
