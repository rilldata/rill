import { type V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector.ts";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { getRangePrecision } from "@rilldata/web-common/lib/time/rill-time-grains.ts";

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

  if (exploreState.selectedTimeRange) {
    deriveIntervalFromRillTimeName(exploreState.selectedTimeRange);
  }
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

/**
 * Derives and sets the interval (time grain) on a time range from its RillTime name.
 * This is needed when the URL doesn't explicitly specify a grain.
 */
function deriveIntervalFromRillTimeName(
  selectedRange: DashboardTimeControls | undefined,
): void {
  if (!selectedRange?.name || selectedRange.interval) return;

  try {
    const parsed = parseRillTime(selectedRange.name);
    selectedRange.interval = getRangePrecision(parsed);
  } catch {
    // Parsing fails for non-rill-time names like "CUSTOM" - use undefined
  }
}
