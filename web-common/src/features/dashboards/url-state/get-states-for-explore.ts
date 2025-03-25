import { getCompoundQuery } from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getMostRecentExploreState } from "@rilldata/web-common/features/dashboards/url-state/most-recent-explore-state";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";

export function getStatesForExplore(
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
  extraPrefix: string | undefined,
) {
  return getCompoundQuery(
    [
      useExploreValidSpec(instanceId, exploreName),
      useMetricsViewTimeRange(instanceId, metricsViewName),
    ],
    ([exploreSpecResp, metricsViewTimeRangeResp]) => {
      const exploreSpec = exploreSpecResp.explore ?? {};
      const metricsViewSpec = exploreSpecResp.metricsView ?? {};

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
    },
  );
}
