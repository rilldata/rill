import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export function getExploreStateFromYAMLConfig(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  // TODO: once getDefaultExplorePreset is not needed in other places we can directly create explore state here
  const explorePreset = getDefaultExplorePreset(
    exploreSpec,
    metricsViewSpec,
    timeRangeSummary,
  );
  // ignore errors for now. issues with yaml would be thrown in rill-dev
  const { partialExploreState: exploreStateFromYAMLConfig } =
    convertPresetToExploreState(metricsViewSpec, exploreSpec, explorePreset);
  return exploreStateFromYAMLConfig;
}
