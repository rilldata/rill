import type { CommonUserFields, MetricsEvent } from "./MetricsTypes";
import type { CommonFields } from "./MetricsTypes";

export class MetricsEventFactory {
  protected getBaseMetricsEvent(
    eventType: string,
    commonFields: CommonFields,
    commonUserFields: CommonUserFields
  ): MetricsEvent {
    return {
      ...commonUserFields,
      ...commonFields,
      event_datetime: Date.now(),
      event_type: eventType,
    };
  }
}
