import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getMostRecentExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

export function getExploreStatesFromYaml(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  metricsViewTimeRangeResp: V1MetricsViewTimeRangeResponse,
  exploreName: string,
  extraPrefix: string | undefined,
) {
  const errors: Error[] = [];

  const defaultExplorePreset = getDefaultExplorePreset(
    {
      ...exploreSpec,
      defaultPreset: {},
    },
    metricsViewSpec,
    metricsViewTimeRangeResp,
  );
  const {
    partialExploreState: defaultExploreState,
    errors: errorsFromDefaultState,
  } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
  errors.push(...errorsFromDefaultState);

  const explorePresetFromYAMLConfig = getDefaultExplorePreset(
    exploreSpec,
    metricsViewSpec,
    metricsViewTimeRangeResp,
  );
  const {
    partialExploreState: exploreStateFromYAMLConfig,
    errors: errorsFromYAMLConfig,
  } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    explorePresetFromYAMLConfig,
  );
  errors.push(...errorsFromYAMLConfig);

  const {
    partialExploreState: mostRecentPartialExploreState,
    errors: errorsFormRecentState,
  } = getMostRecentExploreState(
    exploreName,
    extraPrefix,
    metricsViewSpec,
    exploreSpec,
  );
  errors.push(...errorsFormRecentState);

  return {
    defaultExploreState,
    explorePresetFromYAMLConfig,
    exploreStateFromYAMLConfig,
    mostRecentPartialExploreState,
    errors,
  };
}
