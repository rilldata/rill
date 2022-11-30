import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type { CommonUserFields, MetricsEvent } from "./MetricsTypes";
import type { CommonFields } from "./MetricsTypes";

export class MetricsEventFactory {
  public constructor(protected readonly config: RootConfig) {}

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
