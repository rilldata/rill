import {
  BehaviourEvent,
  BehaviourEventAction,
  BehaviourEventMedium,
} from "./BehaviourEventTypes";
import { MetricsEventFactory } from "./MetricsEventFactory";
import type { CommonFields, CommonUserFields } from "./MetricsTypes";
import { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";

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

  public deployEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    project_id: string,
    action: BehaviourEventAction
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      commonFields,
      commonUserFields
    ) as BehaviourEvent;
    event.action = action;
    event.project_id = project_id;
    event.medium = BehaviourEventMedium.Button;
    event.space = MetricsEventSpace.Workspace;
    event.screen_name = MetricsEventScreenName.Status;
    return event;
  }
}
