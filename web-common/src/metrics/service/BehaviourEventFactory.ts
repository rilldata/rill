import {
  BehaviourEvent,
  BehaviourEventAction,
  BehaviourEventMedium,
} from "./BehaviourEventTypes";
import { MetricsEventFactory } from "./MetricsEventFactory";
import {
  CommonFields,
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "./MetricsTypes";
import type { SourceConnectionType, SourceFileType } from "./SourceEventTypes";

export class BehaviourEventFactory extends MetricsEventFactory {
  public navigationEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    entity_id: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName,
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      BehaviourEventAction.Navigate,
      commonFields,
      commonUserFields,
    ) as BehaviourEvent;
    event.action = BehaviourEventAction.Navigate;
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
    project_id: string,
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      action,
      commonFields,
      commonUserFields,
    ) as BehaviourEvent;
    event.action = action;
    event.entity_id = project_id;
    event.medium = medium;
    event.space = space;
    event.screen_name = MetricsEventScreenName.Splash;
    return event;
  }

  public sourceSuccess(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    medium: BehaviourEventMedium,
    screen_name: MetricsEventScreenName,
    space: MetricsEventSpace,
    connection_type: SourceConnectionType,
    file_type: SourceFileType,
    glob: boolean,
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      BehaviourEventAction.SourceSuccess,
      commonFields,
      commonUserFields,
    ) as BehaviourEvent;
    event.action = BehaviourEventAction.SourceSuccess;
    event.medium = medium;
    event.screen_name = screen_name;
    event.space = space;
    event.connection_type = connection_type;
    event.file_type = file_type;
    event.glob = glob;
    return event;
  }

  public sourceTrigger(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    action: BehaviourEventAction,
    medium: BehaviourEventMedium,
    screen_name: MetricsEventScreenName,
    space: MetricsEventSpace,
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      action,
      commonFields,
      commonUserFields,
    ) as BehaviourEvent;
    event.action = action;
    event.medium = medium;
    event.screen_name = screen_name;
    event.space = space;
    return event;
  }

  public deployIntent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
  ): BehaviourEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      BehaviourEventAction.DeployIntent,
      commonFields,
      commonUserFields,
    ) as BehaviourEvent;
    event.action = BehaviourEventAction.DeployIntent;
    event.medium = BehaviourEventMedium.Button;
    event.space = MetricsEventSpace.Workspace;
    event.screen_name = MetricsEventScreenName.Dashboard;
    return event;
  }
}
