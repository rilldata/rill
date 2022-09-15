import type { MetricsEvent } from "$common/metrics-service/MetricsTypes";
import type {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "$common/metrics-service/MetricsTypes";

export enum BehaviourEventAction {
  Navigate = "navigate",
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
