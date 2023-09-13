import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";

// This has a bunch of helpers that helps easily set telemetry params

export type TelemetryParams = {
  medium: BehaviourEventMedium;
  space: MetricsEventSpace;
  sourceScreen?: MetricsEventScreenName;
  screenName?: MetricsEventScreenName;
};

export function getLeftPanelModelParams(): TelemetryParams {
  return {
    medium: BehaviourEventMedium.Menu,
    space: MetricsEventSpace.LeftPanel,
  };
}
export function getRightPanelModelParams(): TelemetryParams {
  return {
    medium: BehaviourEventMedium.Button,
    space: MetricsEventSpace.RightPanel,
  };
}
