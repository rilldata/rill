import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlSearch } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-search";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

export function getDefaultExploreUrlParams(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  const defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    metricsViewSpec,
    fullTimeRange,
  );
  const { partialExploreState } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    fullTimeRange?.timeRangeSummary,
    partialExploreState as any,
  );
  return convertPartialExploreStateToUrlSearch(
    partialExploreState,
    exploreSpec,
    timeControlState,
    new URLSearchParams(),
  );
}
