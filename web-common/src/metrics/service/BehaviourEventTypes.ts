import type { MetricsEvent } from "./MetricsTypes";
import type { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";

export enum BehaviourEventAction {
  Navigate = "navigate",
  PublishStart = "publish-start",
  PublishSuccess = "publish-success",

  // Splash Screen Actions
  SourceModal = "source-modal",
  ExampleAdd = "example-add",
  ProjectEmpty = "project-empty",
}

export enum BehaviourEventMedium {
  Button = "button",
  Menu = "menu",
  AssetName = "asset-name",
  Card = "card",
}

export interface BehaviourEvent extends MetricsEvent {
  action: BehaviourEventAction;
  medium: BehaviourEventMedium;
  entity_id: string;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  source_screen: MetricsEventScreenName;
}
