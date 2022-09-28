import { MetricsEventFactory } from "./MetricsEventFactory";
import type { CommonFields, CommonUserFields } from "./MetricsTypes";
import { BehaviourEvent, BehaviourEventAction } from "./BehaviourEventTypes";
import type { BehaviourEventMedium } from "./BehaviourEventTypes";
import type { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";

export class BehaviourEventFactory extends MetricsEventFactory {
  public navigationEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    entity_id: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      commonFields,
      commonUserFields
    ) as BehaviourEvent;
    event.action = BehaviourEventAction.Navigate;
    event.entity_id = entity_id;
    event.medium = medium;
    event.space = space;
    event.screen_name = screen_name;
    event.source_screen = source_screen;
    return event;
  }
}
