import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export function getDefaultExploreUrlParams(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const partialExploreState = getExploreStateFromYAMLConfig(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    partialExploreState as any,
  );
  return convertPartialExploreStateToUrlParams(
    exploreSpec,
    partialExploreState,
    timeControlState,
  );
}
