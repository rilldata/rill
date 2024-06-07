import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
  AD_BIDS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
import {
  createAndExpression,
  createBinaryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  V1MetricsViewSpec,
  V1Operation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("AdvancedMeasureCorrector", () => {
  const MetricsView = {
    ...AD_BIDS_INIT,
    measures: AD_BIDS_ADVANCED_MEASURES,
  } as V1MetricsViewSpec;

  it("changing grain while in TDD for measure based on timestamp", () => {
    const dashboard = getDefaultMetricsExplorerEntity(
      "AdBids",
      MetricsView,
      undefined,
    );
    dashboard.tdd.expandedMeasureName = AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN;

    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    );

    // changing selected grain doesn't impact measure with no grain dependence
    dashboard.selectedTimeRange = {
      interval: V1TimeGrain.TIME_GRAIN_DAY,
    } as DashboardTimeControls;
    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    );

    dashboard.tdd.expandedMeasureName = AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN;
    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.tdd.expandedMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
    );

    // changing selected grain unsets measure with a particular grain dependence
    dashboard.selectedTimeRange = {
      interval: V1TimeGrain.TIME_GRAIN_WEEK,
    } as DashboardTimeControls;
    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.tdd.expandedMeasureName).toEqual("");
  });

  it("metrics view spec changed converting a measure to an advanced measure", () => {
    const dashboard = getDefaultMetricsExplorerEntity(
      "AdBids",
      MetricsView,
      undefined,
    );
    dashboard.leaderboardMeasureName = AD_BIDS_IMPRESSIONS_MEASURE;
    dashboard.dimensionThresholdFilters = [
      {
        name: AD_BIDS_PUBLISHER_DIMENSION,
        filter: createAndExpression([
          createBinaryExpression(
            AD_BIDS_IMPRESSIONS_MEASURE,
            V1Operation.OPERATION_GT,
            10,
          ),
        ]),
      },
    ];

    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.leaderboardMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
    expect(dashboard.dimensionThresholdFilters[0]?.filter).not.toBeUndefined();

    // metrics view spec updated to make AD_BIDS_IMPRESSIONS_MEASURE an advanced measure
    AdvancedMeasureCorrector.correct(dashboard, {
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
    });
    expect(dashboard.leaderboardMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    );
    expect(dashboard.dimensionThresholdFilters.length).toEqual(0);
  });
});
