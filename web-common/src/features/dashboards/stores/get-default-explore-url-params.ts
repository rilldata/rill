import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlSearch } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-search";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
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
  const defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    metricsViewSpec,
    timeRangeSummary,
  );
  const { partialExploreState } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    partialExploreState as any,
  );
  return convertPartialExploreStateToUrlSearch(
    partialExploreState,
    exploreSpec,
    timeControlState,
  );
}
