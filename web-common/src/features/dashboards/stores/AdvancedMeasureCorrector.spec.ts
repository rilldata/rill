import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1TimeGrain,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  AD_BIDS_ADVANCED_MEASURES,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
  AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "./test-data/data";

describe("AdvancedMeasureCorrector", () => {
  // The advanced measures require the timestamp dimension, which is the metrics view's time dimension.
  const MetricsView = {
    ...AD_BIDS_METRICS_INIT_WITH_TIME,
    measures: AD_BIDS_ADVANCED_MEASURES,
  } as V1MetricsViewSpec;
  const Explore = {
    ...AD_BIDS_EXPLORE_INIT,
    measures: AD_BIDS_ADVANCED_MEASURES.map((m) => m.name!),
  } as V1ExploreSpec;

  it("changing grain while in TDD for measure based on timestamp", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsView, Explore),
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
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsView, Explore),
    );
    dashboard.leaderboardSortByMeasureName = AD_BIDS_IMPRESSIONS_MEASURE;

    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.leaderboardSortByMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );

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
    // no measure is valid for leaderboards now, so the sort measure is kept rather than cleared
    expect(dashboard.leaderboardSortByMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
    expect(dashboard.leaderboardMeasureNames).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
  });

  it("leaderboards drop measures that require the time dimension", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsView, Explore),
    );
    dashboard.leaderboardMeasureNames = [
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      AD_BIDS_IMPRESSIONS_MEASURE,
    ];
    dashboard.leaderboardSortByMeasureName =
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN;

    AdvancedMeasureCorrector.correct(dashboard, MetricsView);
    expect(dashboard.leaderboardSortByMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
    expect(dashboard.leaderboardMeasureNames).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
  });

  const MeasurePerDomain = "impressions_per_domain";
  const MetricsViewWithMeasurePerDomain = {
    ...MetricsView,
    measures: [
      ...AD_BIDS_ADVANCED_MEASURES,
      {
        name: MeasurePerDomain,
        requiredDimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
      },
    ],
  } as V1MetricsViewSpec;

  it("dimension table keeps a sort measure that requires only its dimension", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsViewWithMeasurePerDomain, {
        ...Explore,
        measures: [AD_BIDS_IMPRESSIONS_MEASURE, MeasurePerDomain],
      }),
    );
    dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
    dashboard.selectedDimensionName = AD_BIDS_DOMAIN_DIMENSION;
    dashboard.leaderboardMeasureNames = [MeasurePerDomain];
    dashboard.leaderboardSortByMeasureName = MeasurePerDomain;

    AdvancedMeasureCorrector.correct(
      dashboard,
      MetricsViewWithMeasurePerDomain,
    );
    expect(dashboard.leaderboardSortByMeasureName).toEqual(MeasurePerDomain);
    expect(dashboard.leaderboardMeasureNames).toEqual([MeasurePerDomain]);

    // leaderboards cannot query it
    dashboard.activePage = DashboardState_ActivePage.DEFAULT;
    dashboard.selectedDimensionName = undefined;
    AdvancedMeasureCorrector.correct(
      dashboard,
      MetricsViewWithMeasurePerDomain,
    );
    expect(dashboard.leaderboardSortByMeasureName).toEqual(
      AD_BIDS_IMPRESSIONS_MEASURE,
    );
  });

  it("TDD drops a measure that requires a dimension other than the time dimension", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsViewWithMeasurePerDomain, Explore),
    );
    dashboard.activePage = DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
    dashboard.tdd.expandedMeasureName = MeasurePerDomain;

    AdvancedMeasureCorrector.correct(
      dashboard,
      MetricsViewWithMeasurePerDomain,
    );
    expect(dashboard.tdd.expandedMeasureName).toEqual("");
    expect(dashboard.activePage).toEqual(DashboardState_ActivePage.DEFAULT);
  });

  it("pivot keeps measures with required dimensions, since its rows and columns can include them", () => {
    const dashboard = getFullInitExploreState(
      "AdBids",
      getInitExploreStateForTest(MetricsViewWithMeasurePerDomain, Explore),
    );
    dashboard.pivot.columns = [
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      MeasurePerDomain,
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
    ].map((id) => ({ id, title: id, type: PivotChipType.Measure }));

    AdvancedMeasureCorrector.correct(
      dashboard,
      MetricsViewWithMeasurePerDomain,
    );
    expect(dashboard.pivot.columns.map((c) => c.id)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
      MeasurePerDomain,
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
  });
});
