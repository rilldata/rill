import type { BehaviourEventMedium } from "./BehaviourEventTypes";
import { BehaviourEvent, BehaviourEventAction } from "./BehaviourEventTypes";
import { MetricsEventFactory } from "./MetricsEventFactory";
import {
  CommonFields,
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "./MetricsTypes";

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

  public publishEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    entity_id: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName,
    isStart: boolean
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      commonFields,
      commonUserFields
    ) as BehaviourEvent;
    event.action = isStart
      ? BehaviourEventAction.PublishStart
      : BehaviourEventAction.PublishSuccess;
    event.entity_id = entity_id;
    event.medium = medium;
    event.space = space;
    event.screen_name = screen_name;
    event.source_screen = source_screen;
    return event;
  }

  public splashEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    action: BehaviourEventAction,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    project_id: string
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      commonFields,
      commonUserFields
    ) as BehaviourEvent;
    event.action = action;
    event.entity_id = project_id;
    event.medium = medium;
    event.space = space;
    event.screen_name = MetricsEventScreenName.Splash;
    return event;
  }
}
