import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  V1TimeGrain,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "./test-data/data";
import { correctExploreState } from "@rilldata/web-common/features/dashboards/stores/correct-explore-state.ts";

describe("correctExploreState", () => {
  const MetricsView = {
    ...AD_BIDS_METRICS_INIT,
    measures: AD_BIDS_ADVANCED_MEASURES,
  } as V1MetricsViewSpec;
  const Explore = {
    ...AD_BIDS_EXPLORE_INIT,
    measures: AD_BIDS_ADVANCED_MEASURES.map((m) => m.name!),
  } as V1ExploreSpec;

  it("changing grain while in TDD for measure based on timestamp", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(
        MetricsView,
        Explore,
        AD_BIDS_TIME_RANGE_SUMMARY,
      ),
    );
    dashboard.tdd.expandedMeasureName = AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN;

    correctExploreState(MetricsView, dashboard);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    );

    // changing selected grain doesn't impact measure with no grain dependence
    dashboard.selectedTimeRange = {
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    } as DashboardTimeControls;
    correctExploreState(MetricsView, dashboard);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    );

    dashboard.tdd.expandedMeasureName = AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN;
    correctExploreState(MetricsView, dashboard);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
    );

    // changing selected grain unsets measure with a particular grain dependence
    dashboard.selectedTimeRange = {
      interval: V1TimeGrain.TIME_GRAIN_WEEK,
    } as DashboardTimeControls;
    correctExploreState(MetricsView, dashboard);
    expect(dashboard.tdd.expandedMeasureName).toEqual("");
  });

  it("metrics view spec changed converting a measure to an advanced measure", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsView, Explore),
    );
    dashboard.leaderboardSortByMeasureName = AD_BIDS_IMPRESSIONS_MEASURE;
    dashboard.dimensionThresholdFilters = [
      {
        name: AD_BIDS_PUBLISHER_DIMENSION,
        filters: [
          {
            measure: AD_BIDS_IMPRESSIONS_MEASURE,
            operation: MeasureFilterOperation.GreaterThan,
            type: MeasureFilterType.Value,
            value1: "10",
            value2: "",
          },
        ],
      },
    ];

    correctExploreState(MetricsView, dashboard);
    expect(dashboard.leaderboardSortByMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
    expect(dashboard.dimensionThresholdFilters[0]?.filters.length).toEqual(1);

    // metrics view spec updated to make AD_BIDS_IMPRESSIONS_MEASURE an advanced measure
    const updatedMetricsView = {
      ...MetricsView,
      measures: [
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE,
          expression: "count(*)",
          window: {
            partition: true,
          },
        },
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
          requiredDimensions: [
            {
              name: AD_BIDS_TIMESTAMP_DIMENSION,
              timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
            },
          ],
        },
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
          requiredDimensions: [
            {
              name: AD_BIDS_TIMESTAMP_DIMENSION,
              timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          ],
        },
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
          window: {
            partition: true,
          },
        },
      ],
    } as V1MetricsViewSpec;
    correctExploreState(updatedMetricsView, dashboard);
    // Should correct to a valid measure - either a non-advanced measure or clear invalid state
    expect(dashboard.leaderboardSortByMeasureName).toBeTruthy();
    expect([
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
    ]).toContain(dashboard.leaderboardSortByMeasureName);
    expect(dashboard.dimensionThresholdFilters.length).toEqual(0);
  });
});
