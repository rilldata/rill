import type { RootConfig } from "$common/config/RootConfig";
import type { CommonUserFields, MetricsEvent } from "$common/metrics/MetricsTypes";
import type { CommonFields } from "$common/metrics/MetricsTypes";

export class MetricsEventFactory {
    public constructor(protected readonly config: RootConfig) {}

    protected getBaseMetricsEvent(commonFields: CommonFields,
                                  commonUserFields: CommonUserFields,
                                  eventType: string): MetricsEvent {
        return {
            ...commonUserFields,
            ...commonFields,
            event_datetime: Date.now(),
            event_type: eventType,
        };
    }
}
