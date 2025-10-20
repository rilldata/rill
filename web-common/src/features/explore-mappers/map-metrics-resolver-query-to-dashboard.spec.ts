import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { describe, expect, it } from "vitest";

describe("mapMetricsResolverQueryToDashboard", () => {
  const TestCases: {
    title: string;
    query: MetricsResolverQuery;
    expectedPartialExplore: Partial<ExploreState>;
  }[] = [
    {
      title: "single measure and dimension, sort by compare measure value",
      query: {
        time_range: { start: "2022-01-01", end: "2022-01-07" },
        comparison_time_range: { start: "2022-02-01", end: "2022-02-07" },
        measures: [
          { name: AD_BIDS_IMPRESSIONS_MEASURE },
          {
            name: AD_BIDS_IMPRESSIONS_MEASURE + "_delta",
            compute: {
              comparison_delta: { measure: AD_BIDS_IMPRESSIONS_MEASURE },
            },
          },
        ],
        dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
        sort: [{ desc: true, name: AD_BIDS_IMPRESSIONS_MEASURE + "_delta" }],
      },
      expectedPartialExplore: {
        activePage: DashboardState_ActivePage.DIMENSION_TABLE,
        selectedTimeRange: {
          name: TimeRangePreset.CUSTOM,
          start: new Date("2022-01-01T00:00:00.000Z"),
          end: new Date("2022-01-07T00:00:00.000Z"),
        },
        selectedComparisonTimeRange: {
          name: TimeRangePreset.CUSTOM,
          start: new Date("2022-02-01T00:00:00.000Z"),
          end: new Date("2022-02-07T00:00:00.000Z"),
        },
        showTimeComparison: true,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        allDimensionsVisible: false,
        selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      },
    },

    {
      title: "single measure and dimensions with additional time dimension",
      query: {
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        dimensions: [
          { name: AD_BIDS_PUBLISHER_DIMENSION },
          {
            name: AD_BIDS_TIMESTAMP_DIMENSION,
            compute: {
              time_floor: {
                dimension: AD_BIDS_TIMESTAMP_DIMENSION,
                grain: "day",
              },
            },
          },
        ],
        sort: [{ desc: true, name: AD_BIDS_IMPRESSIONS_MEASURE }],
      },
      expectedPartialExplore: {
        activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
        selectedTimeRange: {
          name: TimeRangePreset.ALL_TIME,
          interval: V1TimeGrain.TIME_GRAIN_DAY,
        } as DashboardTimeControls,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
        allDimensionsVisible: false,
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,

        selectedComparisonDimension: AD_BIDS_PUBLISHER_DIMENSION,
        tdd: {
          expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
          chartType: TDDChart.DEFAULT,
          pinIndex: -1,
        },
      },
    },

    {
      title: "single measure and 2 dimensions",
      query: {
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        dimensions: [
          { name: AD_BIDS_PUBLISHER_DIMENSION },
          { name: AD_BIDS_DOMAIN_DIMENSION },
        ],
        sort: [{ desc: true, name: AD_BIDS_IMPRESSIONS_MEASURE }],
      },
      expectedPartialExplore: {
        activePage: DashboardState_ActivePage.PIVOT,
        selectedTimeRange: {
          name: TimeRangePreset.ALL_TIME,
        } as DashboardTimeControls,

        visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
        allMeasuresVisible: false,
        visibleDimensions: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_DOMAIN_DIMENSION,
        ],
        allDimensionsVisible: false,
        leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
        sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,

        pivot: {
          rows: [],
          columns: [
            {
              id: AD_BIDS_PUBLISHER_DIMENSION,
              title: AD_BIDS_PUBLISHER_DIMENSION,
              type: PivotChipType.Dimension,
            },
            {
              id: AD_BIDS_DOMAIN_DIMENSION,
              title: AD_BIDS_DOMAIN_DIMENSION,
              type: PivotChipType.Dimension,
            },
            {
              id: AD_BIDS_IMPRESSIONS_MEASURE,
              title: AD_BIDS_IMPRESSIONS_MEASURE,
              type: PivotChipType.Measure,
            },
          ],
          sorting: [{ desc: true, id: AD_BIDS_IMPRESSIONS_MEASURE }],
          expanded: {},
          columnPage: 0,
          rowPage: 0,
          enableComparison: false,
          tableMode: "flat",
          activeCell: null,
        },
      },
    },
  ];

  for (const { title, query, expectedPartialExplore } of TestCases) {
    it(title, () => {
      expect(
        mapMetricsResolverQueryToDashboard(
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
          AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
          query,
        ),
      ).toEqual(expectedPartialExplore);
    });
  }
});
