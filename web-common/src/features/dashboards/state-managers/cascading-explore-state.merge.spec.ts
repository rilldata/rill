import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("cascadingExploreStateMerge", () => {
  // Not an exhaustive test, but a sanity test to make sure nothing is broken.
  // TODO: We should add exhaustive merge tests at the DashboardStateSync level
  it("sanity test", () => {
    const finalState = cascadingExploreStateMerge([
      StateFromURL,
      MostRecentState,
      YAMLConfigState,
      RillDefaultState,
    ]);
    expect(finalState).toEqual({
      selectedTimeRange: {
        // selected time range name is from StateFromURL
        name: TimeRangePreset.LAST_SIX_HOURS,
      } as DashboardTimeControls,
      // Comparison settings are from StateFromURL
      showTimeComparison: true,
      selectedComparisonTimeRange: {
        name: TimeComparisonOption.CONTIGUOUS,
      } as DashboardTimeControls,

      // measure visibility is from StateFromURL
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
      allMeasuresVisible: false,
      // dimension visibility is from MostRecentState
      visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
      allDimensionsVisible: false,
      // Sort by is from YAMLConfigState.
      // Note that validation is not done in cascadingExploreStateMerge so while this is technically invalid it is still merged as is.
      leaderboardSortByMeasureName: AD_BIDS_PUBLISHER_COUNT_MEASURE,
      // leaderboard measures is from RillDefaultState
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
    });
  });
});

const StateFromURL: Partial<ExploreState> = {
  selectedTimeRange: {
    name: TimeRangePreset.LAST_SIX_HOURS,
  } as DashboardTimeControls,
  showTimeComparison: true,
  selectedComparisonTimeRange: {
    name: TimeComparisonOption.CONTIGUOUS,
  } as DashboardTimeControls,

  visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
  allMeasuresVisible: false,
};

const MostRecentState: Partial<ExploreState> = {
  visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
  allMeasuresVisible: false,
  visibleDimensions: [AD_BIDS_DOMAIN_DIMENSION],
  allDimensionsVisible: false,
};

const YAMLConfigState: Partial<ExploreState> = {
  selectedTimeRange: {
    name: TimeRangePreset.LAST_24_HOURS,
    interval: V1TimeGrain.TIME_GRAIN_HOUR,
  } as DashboardTimeControls,
  showTimeComparison: false,
  selectedComparisonTimeRange: undefined,

  visibleMeasures: [
    AD_BIDS_IMPRESSIONS_MEASURE,
    AD_BIDS_BID_PRICE_MEASURE,
    AD_BIDS_PUBLISHER_COUNT_MEASURE,
  ],
  allMeasuresVisible: true,
  visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION, AD_BIDS_DOMAIN_DIMENSION],
  allDimensionsVisible: false,
  leaderboardSortByMeasureName: AD_BIDS_PUBLISHER_COUNT_MEASURE,
};

const RillDefaultState: Partial<ExploreState> = {
  selectedTimeRange: {
    name: TimeRangePreset.LAST_7_DAYS,
    interval: V1TimeGrain.TIME_GRAIN_DAY,
  } as DashboardTimeControls,
  showTimeComparison: false,
  selectedComparisonTimeRange: undefined,

  visibleMeasures: [
    AD_BIDS_IMPRESSIONS_MEASURE,
    AD_BIDS_BID_PRICE_MEASURE,
    AD_BIDS_PUBLISHER_COUNT_MEASURE,
  ],
  allMeasuresVisible: true,
  visibleDimensions: [
    AD_BIDS_PUBLISHER_DIMENSION,
    AD_BIDS_DOMAIN_DIMENSION,
    AD_BIDS_COUNTRY_DIMENSION,
  ],
  allDimensionsVisible: true,
  leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
  leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
};
