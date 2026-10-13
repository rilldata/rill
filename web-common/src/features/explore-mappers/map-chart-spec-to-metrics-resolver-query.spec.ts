import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types.ts";
import {
  type CreateChartSpec,
  mapChartSpecToMetricsResolverQuery,
} from "@rilldata/web-common/features/explore-mappers/map-chart-spec-to-metrics-resolver-query.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { describe, expect, it } from "vitest";

const timeRange = {
  start: "2024-01-01T00:00:00Z",
  end: "2024-12-31T23:59:59Z",
};
const comparisonTimeRange = {
  start: "2023-01-01T00:00:00Z",
  end: "2023-12-31T23:59:59Z",
};
// The generated type declares `val` as an object, but lists and scalars are what the resolver accepts.
const where = {
  cond: { op: "in", exprs: [{ name: "country" }, { val: ["US", "CA"] }] },
} as unknown as Expression;

describe("mapChartSpecToMetricsResolverQuery", () => {
  it("maps a bar chart over a nominal dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      where,
      // Ignored: there is no temporal field to floor.
      time_grain: V1TimeGrain.TIME_GRAIN_DAY,
      color: "primary",
      x: { field: "advertiser_name", type: "nominal", limit: 20, sort: "-y" },
      y: { field: "total_bids", type: "quantitative", zeroBasedOrigin: true },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "advertiser_name" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      where,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("maps a line chart over time with a grain and a comparison", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      time_grain: V1TimeGrain.TIME_GRAIN_MONTH,
      x: { field: "__time", type: "temporal" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "month" } },
        },
      ],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
    });
  });

  it("floors a temporal field by day when the chart has no grain", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "__time", type: "temporal" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec).dimensions).toEqual([
      {
        name: "__time",
        compute: { time_floor: { dimension: "__time", grain: "day" } },
      },
    ]);
  });

  it("maps a heatmap, skipping its explicit value order", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      color: { field: "total_bids", type: "quantitative" },
      x: {
        field: "day",
        type: "nominal",
        sort: ["Sunday", "Monday", "Tuesday"],
      },
      y: { field: "hour", type: "nominal", limit: 24, sort: "-color" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "day" }, { name: "hour" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("maps every measure of a multi-measure chart and ignores its measures legend", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      color: { field: "rill_measures", type: "value" },
      x: { field: "__time", type: "temporal" },
      y: {
        field: "clicks",
        fields: ["clicks", "impressions", "ctr"],
        type: "quantitative",
      },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "day" } },
        },
      ],
      measures: [{ name: "clicks" }, { name: "impressions" }, { name: "ctr" }],
      time_range: timeRange,
    });
  });

  it("maps a horizontal bar chart sorted by its measure on x", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "total_bids", type: "quantitative" },
      y: { field: "advertiser_name", type: "nominal", sort: "-x" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "advertiser_name" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("sorts by the comparison delta when the chart compares a single dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "-y_delta" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "advertiser_name" }],
      measures: [
        { name: "total_bids" },
        {
          name: "total_bids_delta",
          compute: { comparison_delta: { measure: "total_bids" } },
        },
      ],
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      sort: [{ name: "total_bids_delta", desc: true }],
    });
  });

  it("falls back to the measure for a delta sort without a comparison", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "y_delta" },
      y: { field: "total_bids", type: "quantitative" },
    };
    const query = mapChartSpecToMetricsResolverQuery(spec);
    expect(query.measures).toEqual([{ name: "total_bids" }]);
    expect(query.sort).toEqual([{ name: "total_bids", desc: false }]);
  });

  it("falls back to the measure for a delta sort when a color split adds a second dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "-y_delta" },
      y: { field: "total_bids", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 5 },
    };
    const query = mapChartSpecToMetricsResolverQuery(spec);
    expect(query.dimensions).toEqual([
      { name: "advertiser_name" },
      { name: "device" },
    ]);
    expect(query.measures).toEqual([{ name: "total_bids" }]);
    expect(query.sort).toEqual([{ name: "total_bids", desc: true }]);
  });

  it("maps a funnel in measures mode to measures only", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids",
      time_range: timeRange,
      breakdownMode: "measures",
      color: "name",
      mode: "width",
      measure: {
        field: "impressions",
        type: "quantitative",
        fields: ["impressions", "video_starts", "video_completes"],
      },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids",
      measures: [
        { name: "impressions" },
        { name: "video_starts" },
        { name: "video_completes" },
      ],
      time_range: timeRange,
    });
  });

  it("maps a funnel in dimension mode with its stage sorted by the measure", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids",
      time_range: timeRange,
      breakdownMode: "dimension",
      stage: { field: "funnel_stage", type: "nominal", sort: "-y" },
      measure: { field: "users", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids",
      dimensions: [{ name: "funnel_stage" }],
      measures: [{ name: "users" }],
      time_range: timeRange,
      sort: [{ name: "users", desc: true }],
    });
  });

  it("maps a donut chart sorted by its slices' dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      measure: { field: "total_bids", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 10, sort: "color" },
      innerRadius: 50,
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "device" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "device", desc: false }],
    });
  });

  it("maps a scatter plot's measures on both axes and its point dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "impressions", type: "quantitative" },
      y: { field: "clicks", type: "quantitative" },
      size: { field: "total_bids", type: "quantitative" },
      dimension: { field: "advertiser_name", type: "nominal", limit: 50 },
      color: "primary",
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [{ name: "advertiser_name" }],
      measures: [
        { name: "impressions" },
        { name: "clicks" },
        { name: "total_bids" },
      ],
      time_range: timeRange,
    });
  });

  it("maps a combo chart's two measures sorted by the first", () => {
    const spec: CreateChartSpec = {
      metrics_view: "auction_metrics",
      time_range: timeRange,
      x: { field: "app_site_name", type: "nominal", sort: "-y" },
      y1: { field: "requests", type: "quantitative", mark: "bar" },
      y2: { field: "bid_rate", type: "quantitative", mark: "line" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "auction_metrics",
      dimensions: [{ name: "app_site_name" }],
      measures: [{ name: "requests" }, { name: "bid_rate" }],
      time_range: timeRange,
      sort: [{ name: "requests", desc: true }],
    });
  });
});

// The replay route feeds the mapped query to the same explore mapper as a `query_metrics_view` call.
describe("mapChartSpecToMetricsResolverQuery opens an explore", () => {
  function exploreStateFor(spec: CreateChartSpec) {
    return mapMetricsResolverQueryToDashboard(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
      { query: mapChartSpecToMetricsResolverQuery(spec) },
    );
  }

  it("opens a bar chart as a filtered dimension table sorted by its measure", () => {
    const state = exploreStateFor({
      metrics_view: AD_BIDS_METRICS_NAME,
      time_range: timeRange,
      where: {
        cond: {
          op: "in",
          exprs: [
            { name: AD_BIDS_DOMAIN_DIMENSION },
            { val: ["a.com", "b.com"] },
          ],
        },
      } as unknown as Expression,
      x: { field: AD_BIDS_PUBLISHER_DIMENSION, type: "nominal", sort: "-y" },
      y: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
    });
    expect(state).toMatchObject({
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimensionName: AD_BIDS_PUBLISHER_DIMENSION,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,
      selectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date(timeRange.start),
        end: new Date(timeRange.end),
      },
      whereFilter: createInExpression(AD_BIDS_DOMAIN_DIMENSION, [
        "a.com",
        "b.com",
      ]),
    });
  });

  it("opens a line chart over time as a time dimension detail at the chart's grain", () => {
    const state = exploreStateFor({
      metrics_view: AD_BIDS_METRICS_NAME,
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      time_grain: V1TimeGrain.TIME_GRAIN_WEEK,
      x: { field: AD_BIDS_TIMESTAMP_DIMENSION, type: "temporal" },
      y: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
    });
    expect(state).toMatchObject({
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      tdd: {
        expandedMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
        chartType: TDDChart.DEFAULT,
      },
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      selectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        interval: V1TimeGrain.TIME_GRAIN_WEEK,
      },
      showTimeComparison: true,
    });
    // The time dimension is the chart's axis, not a dimension of the explore.
    expect(state.visibleDimensions).toBeUndefined();
  });

  it("opens a heatmap as a flat pivot of both dimensions", () => {
    const state = exploreStateFor({
      metrics_view: AD_BIDS_METRICS_NAME,
      time_range: timeRange,
      color: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
      x: {
        field: AD_BIDS_PUBLISHER_DIMENSION,
        type: "nominal",
        sort: ["Facebook", "Google"],
      },
      y: { field: AD_BIDS_DOMAIN_DIMENSION, type: "nominal", sort: "-color" },
    });
    expect(state).toMatchObject({
      activePage: DashboardState_ActivePage.PIVOT,
      pivot: {
        tableMode: "flat",
        rows: [],
        columns: [
          { id: AD_BIDS_PUBLISHER_DIMENSION, type: PivotChipType.Dimension },
          { id: AD_BIDS_DOMAIN_DIMENSION, type: PivotChipType.Dimension },
          { id: AD_BIDS_IMPRESSIONS_MEASURE, type: PivotChipType.Measure },
        ],
        sorting: [{ id: AD_BIDS_IMPRESSIONS_MEASURE, desc: true }],
      },
    });
  });
});
