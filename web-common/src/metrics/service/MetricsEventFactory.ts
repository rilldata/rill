import { v4 as uuidv4 } from "uuid";

import type { CommonUserFields, MetricsEvent } from "./MetricsTypes";
import type { CommonFields } from "./MetricsTypes";

export class MetricsEventFactory {
  protected getBaseMetricsEvent(
    eventType: string,
    eventName: string,
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
  ): MetricsEvent {
    return {
      ...commonUserFields,
      ...commonFields,
      event_id: uuidv4(),
      event_time: new Date().toISOString(),
      event_datetime: Date.now(),
      event_type: eventType,
      event_name: eventName,
    };
  }
}
