import type { BehaviourEventMedium } from "$web-local/common/metrics-service/BehaviourEventTypes";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "$web-local/common/metrics-service/MetricsTypes";
import { sendTelemetryEvent } from "./sendTelemetryEvent";

export class NavigationEventHandler {
  public constructor(private readonly commonUserMetrics: CommonUserFields) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireEvent(
    entity_id: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName
  ) {
    sendTelemetryEvent(
      "navigationEvent",
      this.commonUserMetrics,
      entity_id,
      medium,
      space,
      source_screen,
      screen_name
    );
  }
}
