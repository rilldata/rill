import type { RootConfig } from "$common/config/RootConfig";
import type { CommonUserFields, MetricsEvent } from "$common/metrics-service/MetricsTypes";
import type { CommonFields } from "$common/metrics-service/MetricsTypes";

export class MetricsEventFactory {
    public constructor(protected readonly config: RootConfig) {}

    protected getBaseMetricsEvent(eventType: string,
                                  commonFields: CommonFields,
                                  commonUserFields: CommonUserFields): MetricsEvent {
        return {
            ...commonUserFields,
            ...commonFields,
            event_datetime: Date.now(),
            event_type: eventType,
        };
    }
}
