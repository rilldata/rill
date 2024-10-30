import type { GithubEventFields } from "./GithubEventTypes";
import type { MetricsEvent } from "./MetricsTypes";
import type { MetricsEventScreenName, MetricsEventSpace } from "./MetricsTypes";
import type { SourceEventFields } from "./SourceEventTypes";

export enum BehaviourEventAction {
  Navigate = "navigate",

  DeployIntent = "deploy-intent",
  DeploySuccess = "deploy-success",
  LoginStart = "login-start",
  LoginSuccess = "login-success",

  UserInvite = "user-invite",
  UserDomainWhitelist = "user-domain-whitelist",

  // Splash Screen Actions
  ExampleAdd = "example-add",
  ProjectEmpty = "project-empty",

  // Source Actions
  SourceSuccess = "source-success",
  SourceModal = "source-modal",
  SourceCancel = "source-cancel",
  SourceAdd = "source-add",

  // Github actions
  GithubConnectStart = "ghconnected-start",
  GithubConnectCreateRepo = "ghconnected-create-repo",
  GithubConnectSuccess = "ghconnected-success",
  GithubConnectOverwritePrompt = "ghconnected-overwrite-prompt",
  GithubConnectFailure = "ghconnected-failure",
  GithubDisconnect = "ghconnected-disconnect",
}

export enum BehaviourEventMedium {
  Button = "button",
  Menu = "menu",
  AssetName = "asset-name",
  Card = "card",
  Drag = "drag",
  Tab = "tab",
}

export interface BehaviourEvent
  extends MetricsEvent,
    SourceEventFields,
    GithubEventFields {
  action: BehaviourEventAction;
  medium: BehaviourEventMedium;
  entity_id: string;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  source_screen: MetricsEventScreenName;
  count: number;
}
