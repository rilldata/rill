import { getDimensionTableAggregationRequestForTime } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("getDimensionTableAggregationRequestForTime", () => {
  const MeasurePerPublisher = "impressions_per_publisher";
  // The advanced measures require the timestamp dimension, which is the metrics view's time dimension.
  const MetricsView: V1MetricsViewSpec = {
    ...AD_BIDS_METRICS_INIT_WITH_TIME,
    measures: [
      ...AD_BIDS_ADVANCED_MEASURES,
      {
        name: MeasurePerPublisher,
        requiredDimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
      },
    ],
  };
  const Explore: V1ExploreSpec = {
    ...AD_BIDS_EXPLORE_INIT,
    measures: MetricsView.measures!.map((m) => m.name!),
  };
  const timeRange = {
    start: "2024-01-01T00:00:00Z",
    end: "2024-02-01T00:00:00Z",
  };
  const comparisonTimeRange = {
    start: "2023-12-01T00:00:00Z",
    end: "2024-01-01T00:00:00Z",
  };

  function buildRequest(
    sortByMeasure: string,
    visibleMeasures = [
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    ],
  ) {
    const exploreState = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsView, Explore),
    );
    exploreState.selectedDimensionName = AD_BIDS_PUBLISHER_DIMENSION;
    exploreState.visibleMeasures = visibleMeasures;
    exploreState.leaderboardSortByMeasureName = sortByMeasure;
    exploreState.sortDirection = SortDirection.DESCENDING;

    return getDimensionTableAggregationRequestForTime({
      instanceId: "",
      metricsViewName: AD_BIDS_METRICS_NAME,
      exploreState,
      metricsViewSpec: MetricsView,
      timeRange,
      comparisonTimeRange,
      dimensionSearchText: "",
    });
  }

  it("omits a visible measure that needs the time dimension", () => {
    const request = buildRequest(AD_BIDS_IMPRESSIONS_MEASURE);
    expect(request.dimensions).toEqual([{ name: AD_BIDS_PUBLISHER_DIMENSION }]);
    expect(request.measures?.map((m) => m.name)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      "impressions_prev",
      "impressions_delta",
      "impressions_delta_perc",
    ]);
    expect(request.sort).toEqual([
      { name: AD_BIDS_IMPRESSIONS_MEASURE, desc: true },
    ]);
  });

  it("sorts by the dimension when the sort measure needs the time dimension", () => {
    const request = buildRequest(AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN);
    expect(request.measures?.map((m) => m.name)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
    expect(request.sort).toEqual([
      { name: AD_BIDS_PUBLISHER_DIMENSION, desc: true },
    ]);
  });

  it("keeps a measure that requires the selected dimension and sorts by it", () => {
    const request = buildRequest(MeasurePerPublisher, [
      AD_BIDS_IMPRESSIONS_MEASURE,
      MeasurePerPublisher,
    ]);
    expect(request.measures?.map((m) => m.name)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
      MeasurePerPublisher,
      "impressions_per_publisher_prev",
      "impressions_per_publisher_delta",
      "impressions_per_publisher_delta_perc",
    ]);
    expect(request.sort).toEqual([{ name: MeasurePerPublisher, desc: true }]);
  });
});
