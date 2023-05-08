import type { MetricsEvent } from "./MetricsTypes";
import type { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";

export enum BehaviourEventAction {
  Navigate = "navigate",
  PublishStart = "publish-start",
  PublishSuccess = "publish-success",
  DeployStart = "deploy-start",
  GithubConnectedStart = "ghconnected-start",
  GithubConnectedSuccess = "ghconnected-success",
  DataAccessStart = "dataaccess-start",
  DataAccessSuccess = "dataaccess-success",
}

export enum BehaviourEventMedium {
  Button = "button",
  Menu = "menu",
  AssetName = "asset-name",
}

export interface BehaviourEvent extends MetricsEvent {
  action: BehaviourEventAction;
  medium: BehaviourEventMedium;
  entity_id: string;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  source_screen: MetricsEventScreenName;
}
