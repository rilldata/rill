import {
  metricsExplorerStore,
  updateMetricsExplorerByName,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";
import { toggleExactMeasure } from "../state-managers/actions/exact-measures";
import { exploreExactSurface } from "./explore-surfaces";
import { selectedExactMeasureNames } from "../state-managers/selectors/exact-measures";
import { setMeasureVisibility } from "../state-managers/actions/measures";

function exploreState() {
  return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME];
}

function toggle(
  measureName: string,
  surface: string = exploreExactSurface.measures,
) {
  updateMetricsExplorerByName(AD_BIDS_EXPLORE_NAME, (dashboard) =>
    toggleExactMeasure({ dashboard }, surface, measureName),
  );
}

function selected(surface: string = exploreExactSurface.measures) {
  return [...selectedExactMeasureNames({ dashboard: exploreState() }, surface)];
}

describe("exact measure selection in the explore store", () => {
  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    metricsExplorerStore.init(
      AD_BIDS_EXPLORE_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      ),
    );
  });

  it("toggles a measure on and off", () => {
    expect(selected()).toEqual([]);

    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);
    expect(selected()).toEqual([AD_BIDS_PUBLISHER_COUNT_MEASURE]);

    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);
    expect(exploreState().exactMeasures).toBeUndefined();
  });

  it("resets when the query context changes", () => {
    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);

    updateMetricsExplorerByName(AD_BIDS_EXPLORE_NAME, (dashboard) => {
      dashboard.whereFilter = createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
        "Google",
      ]);
    });

    expect(exploreState().exactMeasures).toBeUndefined();
    expect(selected()).toEqual([]);
  });

  it("resets when the sort changes", () => {
    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);
    updateMetricsExplorerByName(AD_BIDS_EXPLORE_NAME, (dashboard) => {
      dashboard.leaderboardSortByMeasureName = AD_BIDS_BID_PRICE_MEASURE;
    });
    expect(exploreState().exactMeasures).toBeUndefined();
  });

  it("keeps the surfaces independent", () => {
    const leaderboard = exploreExactSurface.leaderboard(
      AD_BIDS_PUBLISHER_DIMENSION,
    );
    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);
    expect(selected(leaderboard)).toEqual([]);

    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE, leaderboard);
    expect(exploreState().exactMeasures?.surfaces).toEqual({
      [exploreExactSurface.measures]: [AD_BIDS_PUBLISHER_COUNT_MEASURE],
      [leaderboard]: [AD_BIDS_PUBLISHER_COUNT_MEASURE],
    });

    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);
    expect(selected()).toEqual([]);
    expect(selected(leaderboard)).toEqual([AD_BIDS_PUBLISHER_COUNT_MEASURE]);
  });

  it("survives presentation changes", () => {
    toggle(AD_BIDS_PUBLISHER_COUNT_MEASURE);

    updateMetricsExplorerByName(AD_BIDS_EXPLORE_NAME, (dashboard) =>
      setMeasureVisibility(
        { dashboard },
        [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_PUBLISHER_COUNT_MEASURE],
        [
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
          AD_BIDS_PUBLISHER_COUNT_MEASURE,
        ],
      ),
    );

    expect(selected()).toEqual([AD_BIDS_PUBLISHER_COUNT_MEASURE]);
  });
});
