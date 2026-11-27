import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import { correctExploreState } from "@rilldata/web-common/features/dashboards/stores/correct-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Converts url params to a complete explore state for the store.
 * Fields without a param on the params' view, like leaderboard fields on pivot or the scrub range,
 * take the value from the rill default state.
 */
export function convertURLSearchParamsToFullExploreState(
  urlParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  rillDefaultExploreState: ExploreState,
) {
  const { partialExploreState, errors } = convertURLSearchParamsToExploreState(
    urlParams,
    metricsViewSpec,
    exploreSpec,
    {},
  );
  const exploreState = cascadingExploreStateMerge([
    partialExploreState,
    rillDefaultExploreState,
  ]) as ExploreState;
  correctExploreState(metricsViewSpec, exploreState);

  return { exploreState, errors };
}
