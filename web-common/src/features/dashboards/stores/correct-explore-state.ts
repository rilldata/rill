import { type V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector.ts";

/**
 * Corrects the final merged explore state.
 * Fixes invalid values (advanced measures, leaderboard sync) rather than removing them.
 * Called after cascading merge to fix mismatches from combining different sources.
 */
export function correctExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreState: ExploreState,
) {
  // Resuse code for now. We might want to consolidate more in the future.
  AdvancedMeasureCorrector.correct(exploreState, metricsViewSpec);

  correctLeaderboardMeasures(exploreState);
}

function correctLeaderboardMeasures(exploreState: ExploreState) {
  const sortMeasureIsInvalid = Boolean(
    exploreState.leaderboardMeasureNames?.length &&
      !exploreState.leaderboardMeasureNames.includes(
        exploreState.leaderboardSortByMeasureName,
      ),
  );
  const leaderboardMeasuresAreInvalid = Boolean(
    !exploreState.leaderboardMeasureNames?.length &&
      exploreState.leaderboardSortByMeasureName,
  );

  if (sortMeasureIsInvalid) {
    exploreState.leaderboardSortByMeasureName =
      exploreState.leaderboardMeasureNames[0];
  } else if (leaderboardMeasuresAreInvalid) {
    exploreState.leaderboardMeasureNames = [
      exploreState.leaderboardSortByMeasureName,
    ];
  }
}
