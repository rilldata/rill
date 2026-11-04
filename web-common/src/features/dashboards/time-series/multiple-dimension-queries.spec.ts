import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_METRICS_INIT_WITH_TIME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getComparisonTopListMeasuresAndSort } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import {
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("getComparisonTopListMeasuresAndSort", () => {
  const MetricsView: V1MetricsViewSpec = {
    ...AD_BIDS_METRICS_INIT_WITH_TIME,
    measures: AD_BIDS_ADVANCED_MEASURES,
  };
  const exploreState = {
    selectedTimeRange: { interval: V1TimeGrain.TIME_GRAIN_DAY },
    sortDirection: SortDirection.DESCENDING,
  } as ExploreState;

  it("keeps a simple measure and sorts by it", () => {
    expect(
      getComparisonTopListMeasuresAndSort(
        exploreState,
        MetricsView,
        [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_IMPRESSIONS_MEASURE,
      ),
    ).toEqual({
      measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
      sort: [{ desc: true, name: AD_BIDS_IMPRESSIONS_MEASURE }],
    });
  });

  it("omits a measure that needs the time dimension and sorts by the dimension instead", () => {
    // TDD keeps such a measure because its time-grouped queries can compute it,
    // but the top list query groups only by the comparison dimension.
    expect(
      getComparisonTopListMeasuresAndSort(
        exploreState,
        MetricsView,
        [{ name: AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN }],
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      ),
    ).toEqual({
      measures: [],
      sort: [{ desc: true, name: AD_BIDS_DOMAIN_DIMENSION }],
    });
  });

  it("keeps sorting by the dimension when that is selected", () => {
    expect(
      getComparisonTopListMeasuresAndSort(
        { ...exploreState, sortDirection: SortDirection.ASCENDING },
        MetricsView,
        [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ),
    ).toEqual({
      measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
      sort: [{ desc: false, name: AD_BIDS_DOMAIN_DIMENSION }],
    });
  });

  it("keeps an ephemeral measure, which has no spec entry", () => {
    const ephemeralMeasure = "impressions_doubled";
    expect(
      getComparisonTopListMeasuresAndSort(
        {
          ...exploreState,
          ephemeralMeasures: [{ name: ephemeralMeasure }],
        } as ExploreState,
        MetricsView,
        [{ name: ephemeralMeasure }],
        AD_BIDS_DOMAIN_DIMENSION,
        ephemeralMeasure,
      ),
    ).toEqual({
      measures: [{ name: ephemeralMeasure }],
      sort: [{ desc: true, name: ephemeralMeasure }],
    });
  });
});
