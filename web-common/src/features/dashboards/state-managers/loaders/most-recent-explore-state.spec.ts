import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { getMostRecentPartialExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { beforeEach, describe, expect, it } from "vitest";

const doubled: EphemeralMeasureDef = {
  name: "doubled",
  displayName: "Doubled",
  expression: "impressions * 2",
};

describe("getMostRecentPartialExploreState", () => {
  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem(
      `rill:app:explore:${AD_BIDS_EXPLORE_NAME}`.toLowerCase(),
      JSON.stringify({
        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, doubled.name],
        allMeasuresVisible: false,
        leaderboardMeasureNames: [doubled.name],
        leaderboardSortByMeasureName: doubled.name,
      }),
    );
  });

  it("drops ad-hoc selections when their definitions are unknown", () => {
    const { mostRecentPartialExploreState } = getMostRecentPartialExploreState(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
    );
    expect(mostRecentPartialExploreState?.visibleMeasures).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
  });

  it("keeps ad-hoc selections when the library's definitions are provided", () => {
    const { mostRecentPartialExploreState } = getMostRecentPartialExploreState(
      AD_BIDS_EXPLORE_NAME,
      undefined,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
      [doubled],
    );
    expect(mostRecentPartialExploreState?.visibleMeasures).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      doubled.name,
    ]);
    expect(mostRecentPartialExploreState?.leaderboardMeasureNames).toEqual([
      doubled.name,
    ]);
    expect(mostRecentPartialExploreState?.leaderboardSortByMeasureName).toBe(
      doubled.name,
    );
    expect(mostRecentPartialExploreState?.ephemeralMeasures).toEqual([doubled]);
  });
});
