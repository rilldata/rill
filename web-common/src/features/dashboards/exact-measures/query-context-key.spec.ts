import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { describe, expect, it } from "vitest";
import { getQueryContextKey } from "./query-context-key";

function baseState() {
  return getInitExploreStateForTest(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    AD_BIDS_EXPLORE_INIT,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
}

describe("getQueryContextKey", () => {
  it("is stable for the same state", () => {
    expect(getQueryContextKey(baseState())).toBe(
      getQueryContextKey(baseState()),
    );
  });

  it("changes when the query context changes", () => {
    const key = getQueryContextKey(baseState());

    const filtered = baseState();
    filtered.whereFilter = createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
      "Google",
    ]);
    expect(getQueryContextKey(filtered)).not.toBe(key);

    const otherView = baseState();
    otherView.activePage = DashboardState_ActivePage.PIVOT;
    expect(getQueryContextKey(otherView)).not.toBe(key);

    const otherTimezone = baseState();
    otherTimezone.selectedTimezone = "Asia/Kolkata";
    expect(getQueryContextKey(otherTimezone)).not.toBe(key);

    const otherSort = baseState();
    otherSort.leaderboardSortByMeasureName = "bid_price";
    expect(getQueryContextKey(otherSort)).not.toBe(key);

    const dimensionTable = baseState();
    dimensionTable.selectedDimensionName = AD_BIDS_PUBLISHER_DIMENSION;
    expect(getQueryContextKey(dimensionTable)).not.toBe(key);
  });

  it("ignores presentation state", () => {
    const key = getQueryContextKey(baseState());

    const hidden = baseState();
    hidden.visibleMeasures = hidden.visibleMeasures.slice(0, 1);
    hidden.allMeasuresVisible = false;
    hidden.visibleDimensions = [];
    hidden.allDimensionsVisible = false;
    expect(getQueryContextKey(hidden)).toBe(key);

    const reordered = baseState();
    reordered.leaderboardMeasureNames = [
      ...reordered.leaderboardMeasureNames,
    ].reverse();
    reordered.leaderboardShowContextForAllMeasures = true;
    expect(getQueryContextKey(reordered)).toBe(key);

    const expandedPivot = baseState();
    expandedPivot.pivot = {
      ...expandedPivot.pivot,
      expanded: { "0": true },
      rowPage: 3,
      columnPage: 2,
    };
    expect(getQueryContextKey(expandedPivot)).toBe(key);
  });
});
