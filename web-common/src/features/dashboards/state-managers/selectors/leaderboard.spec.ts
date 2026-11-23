import { leaderboardMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/leaderboard";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
  AD_BIDS_METRICS_INIT_WITH_TIME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { describe, expect, it } from "vitest";

describe("leaderboardMeasures", () => {
  const dataSources = (dashboard: Partial<ExploreState>) =>
    ({
      dashboard,
      validMetricsView: {
        ...AD_BIDS_METRICS_INIT_WITH_TIME,
        measures: AD_BIDS_ADVANCED_MEASURES,
      },
      validExplore: {
        ...AD_BIDS_EXPLORE_INIT,
        measures: AD_BIDS_ADVANCED_MEASURES.map((m) => m.name!),
      },
    }) as DashboardDataSources;

  it("excludes measures that require a dimension", () => {
    const measures = leaderboardMeasures(
      dataSources({
        visibleMeasures: [
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
          AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
        ],
        leaderboardMeasureNames: [
          AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
        ],
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      }),
    );
    expect(measures.map((m) => m.name)).toEqual([AD_BIDS_IMPRESSIONS_MEASURE]);
  });

  it("falls back to the sort measure when none of the leaderboard measures can be queried", () => {
    const measures = leaderboardMeasures(
      dataSources({
        visibleMeasures: [
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
        ],
        leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE_WINDOW],
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      }),
    );
    expect(measures.map((m) => m.name)).toEqual([AD_BIDS_IMPRESSIONS_MEASURE]);
  });

  it("does not fall back to a sort measure that requires a dimension", () => {
    const measures = leaderboardMeasures(
      dataSources({
        visibleMeasures: [
          AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
          AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
        ],
        leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN],
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      }),
    );
    expect(measures).toEqual([]);
  });
});
