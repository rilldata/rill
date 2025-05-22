import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { convertURLToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

export function convertURLSearchParamsToExploreState(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const errors: Error[] = [];
  
  const { preset, errors: errorsFromPreset } = convertURLToExplorePreset(
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
