import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { convertURLToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export async function convertURLSearchParamsToExploreState(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const errors: Error[] = [];
  const { preset, errors: errorsFromPreset } = await convertURLToExplorePreset(
    searchParams,
    metricsView,
    exploreSpec,
    defaultExplorePreset,
  );
  errors.push(...errorsFromPreset);
  const { partialExploreState, errors: errorsFromEntity } =
    convertPresetToExploreState(metricsView, exploreSpec, preset);
  errors.push(...errorsFromEntity);
  return { partialExploreState, errors };
}
