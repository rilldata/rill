import type { MetricsEvent } from "./MetricsTypes";
import type { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";
import type { SourceEventFields } from "./SourceEventTypes";

export enum BehaviourEventAction {
  Navigate = "navigate",
  PublishStart = "publish-start",
  PublishSuccess = "publish-success",

  // Splash Screen Actions
  ExampleAdd = "example-add",
  ProjectEmpty = "project-empty",

  // Source Actions
  SourceSuccess = "source-success",
  SourceModal = "source-modal",
  SourceCancel = "source-cancel",
  SourceAdd = "source-add",
}

export enum BehaviourEventMedium {
  Button = "button",
  Menu = "menu",
  AssetName = "asset-name",
  Card = "card",
  Drag = "drag",
}

export interface BehaviourEvent extends MetricsEvent, SourceEventFields {
  action: BehaviourEventAction;
  medium: BehaviourEventMedium;
  entity_id: string;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  source_screen: MetricsEventScreenName;
}
