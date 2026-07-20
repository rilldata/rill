import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
import { getUrlForExploreYAMLDefaultState } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params.ts";
import { unorderedParamsAreEqual } from "@rilldata/web-common/lib/url-utils.ts";
import type { YAMLOnlyExploreConfig } from "@rilldata/web-common/features/dashboards/stores/yaml-only-explore-state.svelte.ts";
import { arrayUnorderedEquals } from "@rilldata/web-common/lib/arrayUtils.ts";

export function viewingExploreDefaults(
  curUrlParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  yamlOnlyConfig: YAMLOnlyExploreConfig,
) {
  const rillDefaultUrlParams = getRillDefaultExploreUrlParams(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const yamlDefaultUrlParams = getUrlForExploreYAMLDefaultState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  // No defaults defined, so nothing to save as default
  if (yamlDefaultUrlParams.size === 0) return false;

  const cleanedYamlDefaults = cleanUrlParams(
    yamlDefaultUrlParams,
    rillDefaultUrlParams,
  );
  const paramsAreEqual = unorderedParamsAreEqual(
    cleanedYamlDefaults,
    curUrlParams,
  );

  const pinnedFiltersAreEqual = arrayUnorderedEquals(
    yamlOnlyConfig.pinnedFilters,
    exploreSpec.defaultPreset?.pinnedFilters ?? [],
  );
  const requiredFiltersAreEqual = arrayUnorderedEquals(
    yamlOnlyConfig.requiredFilters,
    exploreSpec.defaultPreset?.requiredFilters ?? [],
  );

  return paramsAreEqual && pinnedFiltersAreEqual && requiredFiltersAreEqual;
}
